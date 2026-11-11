package datamigration

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"

	"github.com/jackc/pgx/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func verifyMongo(ctx context.Context, c Config, pg pgx.Tx, dst *sql.Tx) (int, int, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(c.Mongo))
	if err != nil {
		return 0, 0, fmt.Errorf("configuración MongoDB inválida")
	}
	defer client.Disconnect(context.Background())
	db := client.Database(MongoName)
	if err = db.Collection("migration_control").FindOne(ctx, bson.M{"_id": c.Marker}).Err(); err != nil {
		return 0, 0, fmt.Errorf("copia MongoDB sin marcador de esta ejecución")
	}
	names, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return 0, 0, fmt.Errorf("inventario MongoDB ilegible")
	}
	need := map[string]bool{"resultados_escaner": true, "evidencias_tecnicas": true, "archivos_tecnicos.files": true, "archivos_tecnicos.chunks": true}
	for _, n := range names {
		delete(need, n)
	}
	if len(need) != 0 {
		return 0, 0, fmt.Errorf("faltan colecciones técnicas o GridFS")
	}
	bucket := db.GridFSBucket(options.GridFSBucket().SetName("archivos_tecnicos"))
	seen := map[bson.ObjectID]bool{}
	checked := map[[2]int64]bool{}
	counts := map[string]int{}
	for _, collection := range []string{"resultados_escaner", "evidencias_tecnicas"} {
		cursor, err := db.Collection(collection).Find(ctx, bson.D{})
		if err != nil {
			return 0, 0, fmt.Errorf("lectura MongoDB fallida: %s", collection)
		}
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var doc struct {
				ID        bson.ObjectID `bson:"_id"`
				Diagnosis string        `bson:"diagnosticoId"`
				Actor     string        `bson:"capturadoPor"`
				File      string        `bson:"gridfsFileId"`
				State     string        `bson:"state"`
				Hash      string        `bson:"sha256"`
				Size      int64         `bson:"sizeBytes"`
			}
			if cursor.Decode(&doc) != nil {
				return 0, 0, fmt.Errorf("documento técnico ilegible: %s", collection)
			}
			diagnosis, de := strconv.ParseInt(doc.Diagnosis, 10, 64)
			actor, ae := strconv.ParseInt(doc.Actor, 10, 64)
			if de != nil || ae != nil || diagnosis <= 0 || actor <= 0 || strconv.FormatInt(diagnosis, 10) != doc.Diagnosis || strconv.FormatInt(actor, 10) != doc.Actor {
				return 0, 0, fmt.Errorf("referencia técnica no canónica en %s", collection)
			}
			pair := [2]int64{diagnosis, actor}
			if !checked[pair] {
				var a, b bool
				e1 := pg.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM public.diagnosticos WHERE id_diagnostico=$1) AND EXISTS(SELECT 1 FROM public.usuarios WHERE id_usuario=$2)`, diagnosis, actor).Scan(&a)
				e2 := dst.QueryRowContext(ctx, `SELECT CONVERT(bit,CASE WHEN EXISTS(SELECT 1 FROM dbo.diagnosticos WHERE id_diagnostico=@p1)
    AND EXISTS(SELECT 1 FROM dbo.usuarios WHERE id_usuario=@p2) THEN 1 ELSE 0 END)`, diagnosis, actor).Scan(&b)
				if e1 != nil || e2 != nil || !a || !b {
					return 0, 0, fmt.Errorf("referencia técnica huérfana o ilegible en %s", collection)
				}
				checked[pair] = true
			}
			if collection == "evidencias_tecnicas" {
				id, err := bson.ObjectIDFromHex(doc.File)
				if err != nil || doc.State != "ready" || seen[id] || doc.Size <= 0 || doc.Size > 10<<20 {
					return 0, 0, fmt.Errorf("evidencia incompleta, fuera de tamaño o duplicada")
				}
				var file struct {
					Length   int64 `bson:"length"`
					Metadata struct {
						Evidence  bson.ObjectID `bson:"evidenceId"`
						Diagnosis string        `bson:"diagnosticoId"`
						ManagedBy string        `bson:"managedBy"`
					} `bson:"metadata"`
				}
				if err = db.Collection("archivos_tecnicos.files").FindOne(ctx, bson.M{"_id": id}).Decode(&file); err != nil || file.Length != doc.Size || file.Metadata.Evidence != doc.ID || file.Metadata.Diagnosis != doc.Diagnosis || file.Metadata.ManagedBy != "elbravodeljac" {
					return 0, 0, fmt.Errorf("metadatos de archivo GridFS inconsistentes")
				}
				stream, err := bucket.OpenDownloadStream(ctx, id)
				if err != nil {
					return 0, 0, fmt.Errorf("archivo GridFS ausente")
				}
				sum := sha256.New()
				n, readErr := io.Copy(sum, io.LimitReader(stream, doc.Size+1))
				closeErr := stream.Close()
				if readErr != nil || closeErr != nil || n != doc.Size || hex.EncodeToString(sum.Sum(nil)) != doc.Hash {
					return 0, 0, fmt.Errorf("archivo GridFS alterado o incompleto")
				}
				seen[id] = true
			}
			counts[collection]++
		}
		if cursor.Err() != nil {
			return 0, 0, fmt.Errorf("inventario MongoDB interrumpido")
		}
	}
	files, err := db.Collection("archivos_tecnicos.files").CountDocuments(ctx, bson.D{})
	if err != nil || files != int64(len(seen)) {
		return 0, 0, fmt.Errorf("archivos GridFS sin evidencia asociada")
	}
	chunks, err := db.Collection("archivos_tecnicos.chunks").Aggregate(ctx, mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.M{"_id": "$files_id"}}},
	})
	if err != nil {
		return 0, 0, fmt.Errorf("inventario de fragmentos GridFS ilegible")
	}
	defer chunks.Close(ctx)
	for chunks.Next(ctx) {
		var item struct {
			ID bson.ObjectID `bson:"_id"`
		}
		if chunks.Decode(&item) != nil || !seen[item.ID] {
			return 0, 0, fmt.Errorf("fragmentos GridFS sin evidencia asociada")
		}
	}
	if chunks.Err() != nil {
		return 0, 0, fmt.Errorf("inventario de fragmentos interrumpido")
	}
	return counts["resultados_escaner"], counts["evidencias_tecnicas"], nil
}
