package mongodb

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

// Reconcile only removes old, unreferenced candidate files/chunks in this dedicated
// development bucket. Ready references and valid upload leases are always retained.
// No collection/database is dropped; callers explicitly choose dryRun or repair.
func (s *Store) Reconcile(ctx context.Context, dryRun bool) (int, error) {
	if err := s.Ping(ctx); err != nil {
		return 0, err
	}
	cutoff := time.Now().Add(-2 * LeaseTTL)
	ids := map[bson.ObjectID]bool{}
	cur, err := s.DB.Collection(bucketName+".files").Find(ctx, bson.M{"uploadDate": bson.M{"$lt": cutoff}, "metadata.managedBy": "elbravodeljac"})
	if err != nil {
		return 0, err
	}
	for cur.Next(ctx) {
		var f struct {
			ID bson.ObjectID `bson:"_id"`
		}
		if err = cur.Decode(&f); err != nil {
			cur.Close(ctx)
			return 0, err
		}
		ids[f.ID] = true
	}
	err = cur.Err()
	cur.Close(ctx)
	if err != nil {
		return 0, err
	}
	// GridFS can leave chunks without a files document after a lost connection.
	chunkCursor, err := s.DB.Collection(bucketName+".chunks").Aggregate(ctx, mongo.Pipeline{{{Key: "$group", Value: bson.M{"_id": "$files_id"}}}})
	if err != nil {
		return 0, err
	}
	for chunkCursor.Next(ctx) {
		var f struct {
			ID bson.ObjectID `bson:"_id"`
		}
		if err = chunkCursor.Decode(&f); err != nil {
			chunkCursor.Close(ctx)
			return 0, err
		}
		if f.ID.Timestamp().Before(cutoff) {
			ids[f.ID] = true
		}
	}
	err = chunkCursor.Err()
	chunkCursor.Close(ctx)
	if err != nil {
		return 0, err
	}
	pending, err := s.DB.Collection("evidencias_tecnicas").Find(ctx, bson.M{"state": "uploading", "leaseUntil": bson.M{"$lt": cutoff}})
	if err != nil {
		return 0, err
	}
	for pending.Next(ctx) {
		var r evidenceRecord
		if err = pending.Decode(&r); err != nil {
			pending.Close(ctx)
			return 0, err
		}
		id, parseErr := bson.ObjectIDFromHex(r.GridFSID)
		if parseErr != nil {
			pending.Close(ctx)
			return 0, parseErr
		}
		ids[id] = true
	}
	err = pending.Err()
	pending.Close(ctx)
	if err != nil {
		return 0, err
	}
	n := 0
	for id := range ids {
		var file struct {
			Metadata struct {
				ManagedBy string `bson:"managedBy"`
			} `bson:"metadata"`
		}
		fileErr := s.DB.Collection(bucketName+".files").FindOne(ctx, bson.M{"_id": id}).Decode(&file)
		if fileErr == nil && file.Metadata.ManagedBy != "elbravodeljac" {
			continue
		}
		if fileErr != nil && !errors.Is(fileErr, mongo.ErrNoDocuments) {
			return n, fileErr
		}
		var record evidenceRecord
		err = s.DB.Collection("evidencias_tecnicas").FindOne(ctx, bson.M{"gridfsFileId": id.Hex()}).Decode(&record)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return n, err
		}
		if err == nil && (record.State == "ready" || record.LeaseUntil.After(time.Now())) {
			continue
		}
		if !dryRun {
			bucket := s.DB.GridFSBucket(options.GridFSBucket().SetName(bucketName))
			if err = bucket.Delete(ctx, id); err != nil && !errors.Is(err, mongo.ErrFileNotFound) {
				return n, err
			}
			if record.ID != bson.NilObjectID {
				_, err = s.DB.Collection("evidencias_tecnicas").DeleteOne(ctx, bson.M{"_id": record.ID, "state": "uploading", "leaseUntil": bson.M{"$lt": cutoff}})
				if err != nil {
					return n, err
				}
			}
		}
		n++
	}
	return n, nil
}
