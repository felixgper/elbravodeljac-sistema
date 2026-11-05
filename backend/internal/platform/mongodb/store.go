package mongodb

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"io"
	"time"
)

//go:embed validators.json
var validatorJSON []byte

const bucketName = "archivos_tecnicos"
const LeaseTTL = 5 * time.Minute

type Store struct {
	Client *mongo.Client
	DB     *mongo.Database
}

func ValidateDatabase(name string) error {
	if name != "elbravodeljac_tecnico_test" && name != "elbravodeljac_tecnico_dev" {
		return errors.New("MongoDB solo permite elbravodeljac_tecnico_test/elbravodeljac_tecnico_dev")
	}
	return nil
}
func Open(uri, name string) (*Store, error) {
	if err := ValidateDatabase(name); err != nil {
		return nil, err
	}
	if uri == "" {
		return nil, errors.New("MONGODB_URI es obligatorio")
	}
	c, err := mongo.Connect(options.Client().ApplyURI(uri).SetMaxPoolSize(10).SetServerSelectionTimeout(2 * time.Second).SetTimeout(5 * time.Second).SetBSONOptions(&options.BSONOptions{DefaultDocumentM: true}))
	if err != nil {
		return nil, errors.New("Configuración MongoDB inválida")
	}
	return &Store{Client: c, DB: c.Database(name)}, nil
}
func (s *Store) Close(ctx context.Context) error { return s.Client.Disconnect(ctx) }
func schemas() (map[string]bson.M, error) {
	var m map[string]bson.M
	err := bson.UnmarshalExtJSON(validatorJSON, false, &m)
	return m, err
}
func sameJSON(a, b any) bool {
	x, err := json.Marshal(a)
	if err != nil {
		return false
	}
	y, err := json.Marshal(b)
	return err == nil && bytes.Equal(x, y)
}

// Init is an explicit environment provisioning operation, never called by API startup.
func (s *Store) Init(ctx context.Context) error {
	specs, err := schemas()
	if err != nil {
		return err
	}
	for name, validator := range specs {
		err = s.DB.CreateCollection(ctx, name, options.CreateCollection().SetValidator(validator).SetValidationLevel("strict").SetValidationAction("error"))
		var cmd mongo.CommandError
		if err != nil && (!errors.As(err, &cmd) || cmd.Code != 48) {
			return err
		}
		indexes := []mongo.IndexModel{{Keys: bson.D{{Key: "diagnosticoId", Value: 1}, {Key: "requestId", Value: 1}}, Options: options.Index().SetName("diagnosis_request_unique").SetUnique(true)}, {Keys: bson.D{{Key: "diagnosticoId", Value: 1}, {Key: "capturadoEn", Value: -1}}, Options: options.Index().SetName("diagnosis_capture")}}
		if name == "evidencias_tecnicas" {
			indexes = append(indexes, mongo.IndexModel{Keys: bson.D{{Key: "gridfsFileId", Value: 1}}, Options: options.Index().SetName("gridfs_unique").SetUnique(true)})
		}
		if _, err = s.DB.Collection(name).Indexes().CreateMany(ctx, indexes); err != nil {
			return err
		}
	}
	for _, suffix := range []string{"files", "chunks"} {
		err = s.DB.CreateCollection(ctx, bucketName+"."+suffix)
		var cmd mongo.CommandError
		if err != nil && (!errors.As(err, &cmd) || cmd.Code != 48) {
			return err
		}
	}
	if _, err = s.DB.Collection(bucketName+".files").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "filename", Value: 1}, {Key: "uploadDate", Value: 1}}, Options: options.Index().SetName("filename_1_uploadDate_1")}); err != nil {
		return err
	}
	if _, err = s.DB.Collection(bucketName+".chunks").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "files_id", Value: 1}, {Key: "n", Value: 1}}, Options: options.Index().SetName("files_id_1_n_1").SetUnique(true)}); err != nil {
		return err
	}
	return s.Ping(ctx)
}
func (s *Store) Ping(ctx context.Context) error {
	if err := s.DB.RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		return err
	}
	expected, err := schemas()
	if err != nil {
		return err
	}
	for name, validator := range expected {
		specs, err := s.DB.ListCollectionSpecifications(ctx, bson.M{"name": name})
		if err != nil {
			return err
		}
		if len(specs) != 1 {
			return errors.New("Colecciones técnicas no inicializadas")
		}
		var actual bson.M
		if err = bson.Unmarshal(specs[0].Options, &actual); err != nil {
			return err
		}
		if actual["validationLevel"] != "strict" || actual["validationAction"] != "error" || !sameJSON(actual["validator"], validator) {
			return errors.New("Validador MongoDB distinto del aprobado; requiere revisión explícita")
		}
		indexes, err := s.DB.Collection(name).Indexes().ListSpecifications(ctx)
		if err != nil {
			return err
		}
		found := false
		fileUnique := name != "evidencias_tecnicas"
		for _, index := range indexes {
			if index.Unique == nil || !*index.Unique {
				continue
			}
			var keys bson.D
			if err = bson.Unmarshal(index.KeysDocument, &keys); err != nil {
				return err
			}
			if sameJSON(keys, bson.D{{Key: "diagnosticoId", Value: int32(1)}, {Key: "requestId", Value: int32(1)}}) {
				found = true
			}
			if sameJSON(keys, bson.D{{Key: "gridfsFileId", Value: int32(1)}}) {
				fileUnique = true
			}
		}
		if !found || !fileUnique {
			return errors.New("Faltan índices únicos técnicos")
		}
	}
	return nil
}
func fingerprint(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:]), nil
}

type scanRecord struct {
	ID          bson.ObjectID `bson:"_id"`
	PayloadHash string        `bson:"payloadHash"`
	domain.Scan `bson:",inline"`
}
type evidenceRecord struct {
	ID              bson.ObjectID `bson:"_id"`
	PayloadHash     string        `bson:"payloadHash"`
	State           string        `bson:"state"`
	LeaseOwner      string        `bson:"leaseOwner"`
	LeaseUntil      time.Time     `bson:"leaseUntil"`
	domain.Evidence `bson:",inline"`
}

func (r scanRecord) public() domain.Scan         { r.Scan.ID = r.ID.Hex(); return r.Scan }
func (r evidenceRecord) public() domain.Evidence { r.Evidence.ID = r.ID.Hex(); return r.Evidence }
func (s *Store) PutScan(ctx context.Context, diagnosis, actor string, input domain.ScanInput) (domain.Scan, bool, error) {
	if err := input.Validate(); err != nil {
		return domain.Scan{}, false, err
	}
	if err := s.Ping(ctx); err != nil {
		return domain.Scan{}, false, err
	}
	hash, err := fingerprint(input)
	if err != nil {
		return domain.Scan{}, false, err
	}
	col := s.DB.Collection("resultados_escaner")
	key := bson.M{"diagnosticoId": diagnosis, "requestId": input.RequestID}
	var old scanRecord
	err = col.FindOne(ctx, key).Decode(&old)
	if err == nil {
		if old.PayloadHash != hash {
			return domain.Scan{}, false, domain.ErrConflict
		}
		return old.public(), false, nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return domain.Scan{}, false, err
	}
	now := time.Now().UTC().Truncate(time.Millisecond)
	captured := now
	if input.CapturedAt != nil {
		captured = *input.CapturedAt
	}
	record := scanRecord{ID: bson.NewObjectID(), PayloadHash: hash, Scan: domain.Scan{DiagnosisID: diagnosis, ActorID: actor, CapturedAt: captured, CreatedAt: now, SchemaVersion: 1, RequestID: input.RequestID, Equipment: input.Equipment, DTCs: input.DTCs, Measurements: input.Measurements}}
	if _, err = col.InsertOne(ctx, record); err == nil {
		return record.public(), true, nil
	}
	// Resolve duplicate or an ambiguous acknowledgement before reporting a failure.
	lookup := col.FindOne(ctx, key).Decode(&old)
	if lookup == nil {
		if old.PayloadHash != hash {
			return domain.Scan{}, false, domain.ErrConflict
		}
		return old.public(), false, nil
	}
	return domain.Scan{}, false, err
}
func (s *Store) Scans(ctx context.Context, id string, p domain.Page) (domain.List[domain.Scan], error) {
	out := domain.List[domain.Scan]{Data: []domain.Scan{}, Meta: domain.Meta{Page: p.Number, PageSize: p.Size}}
	if err := s.Ping(ctx); err != nil {
		return out, err
	}
	filter := bson.M{"diagnosticoId": id}
	col := s.DB.Collection("resultados_escaner")
	n, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return out, err
	}
	out.Meta.Total = int(n)
	cur, err := col.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "capturadoEn", Value: -1}, {Key: "_id", Value: -1}}).SetSkip(int64((p.Number-1)*p.Size)).SetLimit(int64(p.Size)))
	if err != nil {
		return out, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var r scanRecord
		if err = cur.Decode(&r); err != nil {
			return out, err
		}
		out.Data = append(out.Data, r.public())
	}
	return out, cur.Err()
}
func evidenceHash(v domain.EvidenceInput) (string, error) {
	return fingerprint(struct {
		Description, MIME, SHA string
		Size                   int
	}{v.Description, v.MIME, v.SHA256, len(v.Bytes)})
}
func (s *Store) PutEvidence(ctx context.Context, diagnosis, actor string, input domain.EvidenceInput) (domain.Evidence, bool, error) {
	if err := input.Validate(); err != nil {
		return domain.Evidence{}, false, err
	}
	if err := s.Ping(ctx); err != nil {
		return domain.Evidence{}, false, err
	}
	hash, err := evidenceHash(input)
	if err != nil {
		return domain.Evidence{}, false, err
	}
	col := s.DB.Collection("evidencias_tecnicas")
	now := time.Now().UTC().Truncate(time.Millisecond)
	fileID := bson.NewObjectID()
	owner := bson.NewObjectID().Hex()
	record := evidenceRecord{ID: bson.NewObjectID(), PayloadHash: hash, State: "uploading", LeaseOwner: owner, LeaseUntil: now.Add(LeaseTTL), Evidence: domain.Evidence{DiagnosisID: diagnosis, ActorID: actor, CapturedAt: now, CreatedAt: now, SchemaVersion: 1, RequestID: input.RequestID, Description: input.Description, MIME: input.MIME, Size: int64(len(input.Bytes)), SHA256: input.SHA256, GridFSID: fileID.Hex()}}
	key := bson.M{"diagnosticoId": diagnosis, "requestId": input.RequestID}
	if _, err = col.InsertOne(ctx, record); err != nil {
		var old evidenceRecord
		if lookup := col.FindOne(ctx, key).Decode(&old); lookup != nil {
			return domain.Evidence{}, false, err
		}
		if old.PayloadHash != hash {
			return domain.Evidence{}, false, domain.ErrConflict
		}
		if old.State == "ready" {
			return old.public(), false, nil
		}
		if old.LeaseUntil.After(now) {
			return domain.Evidence{}, false, domain.ErrConflict
		}
		filter := bson.M{"_id": old.ID, "state": "uploading", "leaseUntil": bson.M{"$lte": now}, "payloadHash": hash}
		update := bson.M{"$set": bson.M{"leaseOwner": owner, "leaseUntil": now.Add(LeaseTTL), "gridfsFileId": fileID.Hex()}}
		if err = col.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&record); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return domain.Evidence{}, false, domain.ErrConflict
			}
			return domain.Evidence{}, false, err
		}
	}
	// A separate bucket instance per upload avoids shared driver read-buffer state.
	bucket := s.DB.GridFSBucket(options.GridFSBucket().SetName(bucketName))
	metadata := bson.M{"evidenceId": record.ID, "leaseOwner": owner, "diagnosticoId": diagnosis, "managedBy": "elbravodeljac"}
	if err = bucket.UploadFromStreamWithID(ctx, fileID, "evidence-"+record.ID.Hex(), bytes.NewReader(input.Bytes), options.GridFSUpload().SetMetadata(metadata)); err != nil {
		s.cleanupAttempt(record, fileID)
		return domain.Evidence{}, false, err
	}
	filter := bson.M{"_id": record.ID, "leaseOwner": owner, "state": "uploading", "leaseUntil": bson.M{"$gt": time.Now()}}
	result, err := col.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"state": "ready"}})
	if err == nil && result.ModifiedCount == 1 {
		record.State = "ready"
		return record.public(), true, nil
	}
	// If acknowledgement was lost, never delete a file possibly referenced by ready metadata.
	checkCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var resolved evidenceRecord
	lookup := col.FindOne(checkCtx, bson.M{"_id": record.ID}).Decode(&resolved)
	if lookup == nil && resolved.State == "ready" && resolved.GridFSID == fileID.Hex() {
		return resolved.public(), false, nil
	}
	if lookup == nil {
		s.cleanupAttempt(record, fileID)
	}
	if err != nil {
		return domain.Evidence{}, false, err
	}
	return domain.Evidence{}, false, domain.ErrConflict
}
func (s *Store) cleanupAttempt(record evidenceRecord, fileID bson.ObjectID) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	col := s.DB.Collection("evidencias_tecnicas")
	var current evidenceRecord
	err := col.FindOne(ctx, bson.M{"_id": record.ID}).Decode(&current)
	if err != nil {
		return
	}
	if current.State == "ready" && current.GridFSID == fileID.Hex() {
		return
	}
	bucket := s.DB.GridFSBucket(options.GridFSBucket().SetName(bucketName))
	if err = bucket.Delete(ctx, fileID); err != nil && !errors.Is(err, mongo.ErrFileNotFound) {
		return
	}
	_, _ = col.DeleteOne(ctx, bson.M{"_id": record.ID, "state": "uploading", "leaseOwner": record.LeaseOwner})
}
func (s *Store) EvidenceList(ctx context.Context, id string, p domain.Page) (domain.List[domain.Evidence], error) {
	out := domain.List[domain.Evidence]{Data: []domain.Evidence{}, Meta: domain.Meta{Page: p.Number, PageSize: p.Size}}
	if err := s.Ping(ctx); err != nil {
		return out, err
	}
	filter := bson.M{"diagnosticoId": id, "state": "ready"}
	col := s.DB.Collection("evidencias_tecnicas")
	n, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return out, err
	}
	out.Meta.Total = int(n)
	cur, err := col.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "capturadoEn", Value: -1}, {Key: "_id", Value: -1}}).SetSkip(int64((p.Number-1)*p.Size)).SetLimit(int64(p.Size)))
	if err != nil {
		return out, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var r evidenceRecord
		if err = cur.Decode(&r); err != nil {
			return out, err
		}
		out.Data = append(out.Data, r.public())
	}
	return out, cur.Err()
}
func (s *Store) File(ctx context.Context, diagnosis, id string) (domain.Evidence, []byte, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return domain.Evidence{}, nil, domain.Invalid("evidenciaId", "Identificador MongoDB inválido.")
	}
	if err = s.Ping(ctx); err != nil {
		return domain.Evidence{}, nil, err
	}
	var record evidenceRecord
	err = s.DB.Collection("evidencias_tecnicas").FindOne(ctx, bson.M{"_id": objectID, "diagnosticoId": diagnosis, "state": "ready"}).Decode(&record)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.Evidence{}, nil, domain.ErrNotFound
	}
	if err != nil {
		return domain.Evidence{}, nil, err
	}
	fileID, err := bson.ObjectIDFromHex(record.GridFSID)
	if err != nil || record.Size > domain.MaxEvidenceBytes || record.Size <= 0 {
		return domain.Evidence{}, nil, errors.New("Metadatos GridFS inconsistentes")
	}
	stream, err := s.DB.GridFSBucket(options.GridFSBucket().SetName(bucketName)).OpenDownloadStream(ctx, fileID)
	if err != nil {
		return domain.Evidence{}, nil, err
	}
	defer stream.Close()
	var buf bytes.Buffer
	buf.Grow(int(record.Size))
	limited := &limitedWriter{w: &buf, left: domain.MaxEvidenceBytes}
	if _, err = io.Copy(limited, stream); err != nil {
		return domain.Evidence{}, nil, err
	}
	sum := sha256.Sum256(buf.Bytes())
	if int64(buf.Len()) != record.Size || hex.EncodeToString(sum[:]) != record.SHA256 {
		return domain.Evidence{}, nil, errors.New("Integridad GridFS inválida")
	}
	return record.public(), buf.Bytes(), nil
}

type limitedWriter struct {
	w    *bytes.Buffer
	left int
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	if len(p) > w.left {
		return 0, fmt.Errorf("Archivo GridFS supera el límite")
	}
	n, err := w.w.Write(p)
	w.left -= n
	return n, err
}
