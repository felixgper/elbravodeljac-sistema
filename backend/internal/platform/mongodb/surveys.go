package mongodb

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/application"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

const surveyCollection = "encuestas_invitaciones"

//go:embed surveys-validator.json
var surveyValidatorJSON []byte
var _ application.Surveys = (*Store)(nil)

// InitSurveys is explicit provisioning, NEVER invoked by API startup. It does
// not change technical collections, GridFS or existing collection validators.
func (s *Store) InitSurveys(ctx context.Context) error {
	var validator bson.M
	if err := bson.UnmarshalExtJSON(surveyValidatorJSON, false, &validator); err != nil {
		return fmt.Errorf("leer validador de encuestas: %w", err)
	}
	err := s.DB.CreateCollection(ctx, surveyCollection, options.CreateCollection().SetValidator(validator).SetValidationLevel("strict").SetValidationAction("error"))
	var cmd mongo.CommandError
	if err != nil && (!errors.As(err, &cmd) || cmd.Code != 48) {
		return fmt.Errorf("crear colección de encuestas: %w", err)
	}
	if err = s.surveySchema(ctx); err != nil {
		return fmt.Errorf("validar colección de encuestas: %w", err)
	}
	_, err = s.DB.Collection(surveyCollection).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "deliveryId", Value: 1}}, Options: options.Index().SetName("survey_delivery_unique").SetUnique(true)},
		{Keys: bson.D{{Key: "tokenHash", Value: 1}}, Options: options.Index().SetName("survey_token_unique").SetUnique(true)},
		{Keys: bson.D{{Key: "response.at", Value: 1}}, Options: options.Index().SetName("survey_answer_date")},
		{Keys: bson.D{{Key: "createdAt", Value: -1}}, Options: options.Index().SetName("survey_tracking")},
	})
	if err != nil {
		return fmt.Errorf("crear índices de encuestas: %w", err)
	}
	return s.surveyReady(ctx)
}
func (s *Store) surveySchema(ctx context.Context) error {
	var expected, actual bson.M
	if err := bson.UnmarshalExtJSON(surveyValidatorJSON, false, &expected); err != nil {
		return err
	}
	specs, err := s.DB.ListCollectionSpecifications(ctx, bson.M{"name": surveyCollection})
	if err != nil {
		return err
	}
	if len(specs) != 1 {
		return errors.New("encuestas requieren inicialización explícita")
	}
	if err = bson.Unmarshal(specs[0].Options, &actual); err != nil {
		return err
	}
	if actual["validationLevel"] != "strict" || actual["validationAction"] != "error" || !sameJSON(actual["validator"], expected) {
		return errors.New("validador de encuestas distinto del aprobado")
	}
	return nil
}
func (s *Store) surveyReady(ctx context.Context) error {
	if err := s.surveySchema(ctx); err != nil {
		return fmt.Errorf("esquema de encuestas: %w", err)
	}
	cursor, err := s.DB.Collection(surveyCollection).Indexes().List(ctx)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	var indexes []surveyIndex
	if err = cursor.All(ctx, &indexes); err != nil {
		return fmt.Errorf("índices de encuestas: %w", err)
	}
	return validateSurveyIndexes(indexes)
}

type surveyIndex struct {
	Key                bson.D `bson:"key"`
	Unique             bool   `bson:"unique"`
	Sparse             bool   `bson:"sparse"`
	Partial            bson.M `bson:"partialFilterExpression"`
	ExpireAfterSeconds *int64 `bson:"expireAfterSeconds"`
}

func validateSurveyIndexes(indexes []surveyIndex) error {
	// Responses and their invitation share one document. A TTL index would
	// erase historical ratings, rather than merely invalidate a public link.
	// Refuse that configuration; provisioning must never drop existing indexes.
	for _, idx := range indexes {
		if idx.ExpireAfterSeconds != nil {
			return errors.New("índice TTL incompatible con el historial de encuestas")
		}
	}
	for _, key := range []string{"deliveryId", "tokenHash"} {
		found := false
		for _, idx := range indexes {
			if len(idx.Key) == 1 && idx.Key[0].Key == key && idx.Unique && !idx.Sparse && len(idx.Partial) == 0 {
				found = true
			}
		}
		if !found {
			return errors.New("falta índice único de encuestas")
		}
	}
	return nil
}
func surveyError(err error) error {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.ErrNotFound
	}
	if mongo.IsDuplicateKeyError(err) {
		return domain.ErrConflict
	}
	return err
}
func (s *Store) IssueSurvey(ctx context.Context, v domain.SurveyInvitation, replace bool) error {
	if err := s.surveyReady(ctx); err != nil {
		return err
	}
	c := s.DB.Collection(surveyCollection)
	if !replace {
		_, err := c.InsertOne(ctx, v)
		return surveyError(err)
	}
	// Explicit regeneration, at most once per minute, never after a response.
	// Immutable ID/delivery and unique index prevent concurrent active invitations.
	r, err := c.UpdateOne(ctx, bson.M{"deliveryId": v.DeliveryID, "orderId": v.OrderID, "response": bson.M{"$exists": false}, "createdAt": bson.M{"$lte": v.CreatedAt.Add(-time.Minute)}}, bson.M{"$set": bson.M{"tokenHash": v.TokenHash, "tokenCipher": v.TokenCipher, "createdAt": v.CreatedAt, "expiresAt": v.ExpiresAt, "version": v.Version}})
	if err != nil {
		return surveyError(err)
	}
	if r.MatchedCount != 1 {
		return domain.ErrConflict
	}
	return nil
}
func (s *Store) SurveyByOrder(ctx context.Context, id string) (domain.SurveyInvitation, error) {
	var v domain.SurveyInvitation
	err := s.DB.Collection(surveyCollection).FindOne(ctx, bson.M{"orderId": id}).Decode(&v)
	return v, surveyError(err)
}
func (s *Store) SurveyByHash(ctx context.Context, hash string) (domain.SurveyInvitation, error) {
	var v domain.SurveyInvitation
	err := s.DB.Collection(surveyCollection).FindOne(ctx, bson.M{"tokenHash": hash}).Decode(&v)
	return v, surveyError(err)
}
func (s *Store) AnswerSurvey(ctx context.Context, hash string, v domain.SurveyResponse) error {
	if err := v.Validate(); err != nil {
		return err
	}
	if err := s.surveyReady(ctx); err != nil {
		return err
	}
	// The server timestamp, not the browser, controls expiration and statistics.
	v.At = time.Now().UTC()
	r, err := s.DB.Collection(surveyCollection).UpdateOne(ctx, bson.M{"tokenHash": hash, "expiresAt": bson.M{"$gt": v.At}, "response": bson.M{"$exists": false}}, bson.M{"$set": bson.M{"response": v}})
	if err != nil {
		return surveyError(err)
	}
	if r.MatchedCount != 1 {
		return domain.ErrConflict
	}
	return nil
}
func (s *Store) SurveyTracking(ctx context.Context, page int) ([]domain.SurveyInvitation, error) {
	cursor, err := s.DB.Collection(surveyCollection).Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}, {Key: "_id", Value: 1}}).SetSkip(int64((page-1)*50)).SetLimit(50).SetProjection(bson.M{"tokenHash": 0, "tokenCipher": 0, "response.ratings": 0, "response.comment": 0}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	v := []domain.SurveyInvitation{}
	err = cursor.All(ctx, &v)
	return v, err
}
func (s *Store) SurveyStatistics(ctx context.Context, start, end time.Time) (domain.SurveyStats, error) {
	group := bson.M{"_id": nil, "total": bson.M{"$sum": 1}}
	for i := 0; i < 5; i++ {
		group[string(rune('a'+i))] = bson.M{"$avg": bson.M{"$arrayElemAt": bson.A{"$response.ratings", i}}}
		group[string(rune('f'+i))] = bson.M{"$sum": bson.M{"$cond": bson.A{bson.M{"$eq": bson.A{bson.M{"$arrayElemAt": bson.A{"$response.ratings", 4}}, i + 1}}, 1, 0}}}
	}
	cursor, err := s.DB.Collection(surveyCollection).Aggregate(ctx, mongo.Pipeline{bson.D{{Key: "$match", Value: bson.M{"response.at": bson.M{"$gte": start, "$lt": end}}}}, bson.D{{Key: "$group", Value: group}}})
	if err != nil {
		return domain.SurveyStats{}, err
	}
	defer cursor.Close(ctx)
	var rows []struct {
		Total         int `bson:"total"`
		A, B, C, D, E float64
		F, G, H, I, J int
	}
	if err = cursor.All(ctx, &rows); err != nil {
		return domain.SurveyStats{}, err
	}
	if len(rows) == 0 {
		return domain.SurveyStats{}, nil
	}
	r := rows[0]
	return domain.SurveyStats{Total: r.Total, Average: r.E, Aspects: [5]float64{r.A, r.B, r.C, r.D, r.E}, Distribution: [5]int{r.F, r.G, r.H, r.I, r.J}}, nil
}
