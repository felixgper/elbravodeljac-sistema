package mongodb

import (
	"context"
	"crypto/rand"
	"errors"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/auth"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSurveyMongoSchema(t *testing.T) {
	var schema bson.M
	if err := bson.UnmarshalExtJSON(surveyValidatorJSON, false, &schema); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(surveyValidatorJSON), `"maxItems": 5`) || !strings.Contains(string(surveyValidatorJSON), `"additionalProperties": false`) {
		t.Fatal("strict schema missing")
	}
	for _, term := range []string{"telefono", "dni", "password", "nombre", "gridfs"} {
		if strings.Contains(strings.ToLower(string(surveyValidatorJSON)), term) {
			t.Fatal("unnecessary personal field", term)
		}
	}
}

func TestSurveyIndexesPreserveHistory(t *testing.T) {
	encoded, err := bson.Marshal(bson.M{"key": bson.D{{Key: "expiresAt", Value: 1}}, "expireAfterSeconds": int32(0)})
	if err != nil {
		t.Fatal(err)
	}
	var decoded surveyIndex
	if err := bson.Unmarshal(encoded, &decoded); err != nil || decoded.ExpireAfterSeconds == nil {
		t.Fatalf("TTL metadata decode: %v", err)
	}
	indexes := []surveyIndex{
		{Key: bson.D{{Key: "deliveryId", Value: 1}}, Unique: true},
		{Key: bson.D{{Key: "tokenHash", Value: 1}}, Unique: true},
	}
	if err := validateSurveyIndexes(indexes); err != nil {
		t.Fatal(err)
	}
	for _, ttl := range []int64{0, 14 * 24 * 60 * 60} {
		unsafe := append(append([]surveyIndex{}, indexes...), surveyIndex{Key: bson.D{{Key: "expiresAt", Value: 1}}, ExpireAfterSeconds: &ttl})
		if err := validateSurveyIndexes(unsafe); err == nil {
			t.Fatal("TTL could erase answered surveys")
		}
	}
}

// Explicit, isolated QA only. No default URI or production environment fallback.
// Each run gets a fresh database; cleanup drops only its own survey collection.
func TestSurveyMongoAtomicIntegration(t *testing.T) {
	uri := os.Getenv("SURVEY_QA_URI")
	if uri == "" {
		t.Skip("requires dedicated disposable MongoDB QA; never use official MONGODB_URI")
	}
	if os.Getenv("SURVEY_QA_ISOLATED") != "yes" {
		t.Fatal("explicit isolated QA confirmation required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	client, err := mongo.Connect(options.Client().ApplyURI(uri).SetServerSelectionTimeout(3 * time.Second))
	if err != nil {
		t.Fatal("QA connection unavailable")
	}
	defer client.Disconnect(ctx)
	s := &Store{Client: client, DB: client.Database("elbravodeljac_encuestas_qa_" + strings.ToLower(rand.Text()))}
	if err = s.DB.RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		t.Fatal("QA MongoDB unavailable")
	}
	if err = s.InitSurveys(ctx); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanup, c := context.WithTimeout(context.Background(), 5*time.Second)
		defer c()
		if err := s.DB.Collection(surveyCollection).Drop(cleanup); err != nil {
			t.Error("QA collection cleanup failed")
		}
	}()
	if err = s.InitSurveys(ctx); err != nil {
		t.Fatal("idempotent provisioning", err)
	}
	// An unexpected TTL must fail closed before it can erase embedded answers.
	// This disposable collection is still empty; no historical data is at risk.
	if _, err = s.DB.Collection(surveyCollection).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "expiresAt", Value: 1}}, Options: options.Index().SetName("qa_unsafe_ttl").SetExpireAfterSeconds(0),
	}); err != nil {
		t.Fatal(err)
	}
	if err = s.InitSurveys(ctx); err == nil {
		t.Fatal("provisioning accepted a TTL that could erase historical responses")
	}
	if err = s.DB.Collection(surveyCollection).Indexes().DropOne(ctx, "qa_unsafe_ttl"); err != nil {
		t.Fatal(err)
	}
	if err = s.InitSurveys(ctx); err != nil {
		t.Fatal("provisioning after QA-only TTL removal", err)
	}
	_, hash, _ := domain.NewSurveyToken()
	now := time.Now().UTC().Truncate(time.Millisecond)
	v := domain.SurveyInvitation{ID: strings.Repeat("a", 32), DeliveryID: "1", OrderID: "1", TokenHash: hash, CreatedAt: now.Add(-2 * time.Minute), ExpiresAt: now.Add(time.Hour), Version: 1}
	key := []byte(strings.Repeat("k", 32))
	rawToken, hash, _ := domain.NewSurveyToken()
	v.TokenHash = hash
	v.TokenCipher, err = auth.SealSurveyLink(key, rawToken, "qa-link")
	if err != nil {
		t.Fatal(err)
	}
	if err = s.IssueSurvey(ctx, v, false); err != nil {
		t.Fatal(err)
	}
	persisted, readErr := s.SurveyByOrder(ctx, "1")
	if readErr != nil {
		t.Fatal(readErr)
	}
	recovered, readErr := auth.OpenSurveyLink(key, persisted.TokenCipher, "qa-link")
	if readErr != nil || recovered != rawToken {
		t.Fatal("persisted ciphertext not recoverable")
	}
	if err = s.IssueSurvey(ctx, v, false); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("unique delivery", err)
	}
	_, newHash, _ := domain.NewSurveyToken()
	replacement := v
	replacement.TokenHash = newHash
	replacement.CreatedAt = now
	if err = s.IssueSurvey(ctx, replacement, true); err != nil {
		t.Fatal(err)
	}
	if _, err = s.SurveyByHash(ctx, hash); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("old token still usable")
	}
	if err = s.IssueSurvey(ctx, replacement, true); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("cooldown")
	}
	var wg sync.WaitGroup
	results := make(chan error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- s.AnswerSurvey(ctx, newHash, domain.SurveyResponse{Ratings: [5]int{1, 2, 3, 4, 5}, Comment: "QA"})
		}()
	}
	wg.Wait()
	close(results)
	wins := 0
	for err := range results {
		if err == nil {
			wins++
		} else if !errors.Is(err, domain.ErrConflict) {
			t.Fatal(err)
		}
	}
	if wins != 1 {
		t.Fatalf("atomic winners=%d", wins)
	}
	if err = s.IssueSurvey(ctx, replacement, true); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("answered invitation regenerated")
	}
	stats, err := s.SurveyStatistics(ctx, now.Add(-time.Hour), now.Add(time.Hour))
	if err != nil || stats.Total != 1 || stats.Average != 5 || stats.Aspects != [5]float64{1, 2, 3, 4, 5} || stats.Distribution != [5]int{0, 0, 0, 0, 1} {
		t.Fatalf("aggregation: %+v %v", stats, err)
	}
	empty, err := s.SurveyStatistics(ctx, now.Add(time.Hour), now.Add(2*time.Hour))
	if err != nil || empty.Total != 0 {
		t.Fatal("period boundary")
	}
	v.ID = strings.Repeat("b", 32)
	v.DeliveryID = "2"
	v.OrderID = "2"
	_, v.TokenHash, _ = domain.NewSurveyToken()
	v.ExpiresAt = now.Add(-time.Second)
	if err = s.IssueSurvey(ctx, v, false); err != nil {
		t.Fatal(err)
	}
	if err = s.AnswerSurvey(ctx, v.TokenHash, domain.SurveyResponse{Ratings: [5]int{5, 5, 5, 5, 5}}); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("expired response accepted")
	}
	bad := bson.M{"_id": strings.Repeat("c", 32), "deliveryId": "3", "orderId": "3", "tokenHash": strings.Repeat("d", 64), "createdAt": now, "expiresAt": now.Add(time.Hour), "version": 1, "response": bson.M{"ratings": bson.A{5, 5, 5, 5, 6}, "at": now, "comment": "QA"}}
	if _, err = s.DB.Collection(surveyCollection).InsertOne(ctx, bad); err == nil {
		t.Fatal("validator accepted invalid rating")
	}
	// Q5, not the mean of all aspects; unanswered invitations are excluded.
	third := v
	third.ID = strings.Repeat("c", 32)
	third.DeliveryID = "3"
	third.OrderID = "3"
	third.ExpiresAt = now.Add(time.Hour)
	_, third.TokenHash, _ = domain.NewSurveyToken()
	if err = s.IssueSurvey(ctx, third, false); err != nil {
		t.Fatal(err)
	}
	if err = s.AnswerSurvey(ctx, third.TokenHash, domain.SurveyResponse{Ratings: [5]int{5, 5, 5, 5, 1}}); err != nil {
		t.Fatal(err)
	}
	stats, err = s.SurveyStatistics(ctx, now.Add(-time.Hour), now.Add(time.Hour))
	if err != nil || stats.Total != 2 || stats.Average != 3 || stats.Aspects != [5]float64{3, 3.5, 4, 4.5, 3} {
		t.Fatal("Q5 denominator or aspect contamination", stats, err)
	}
	// Synthetic historical times on this disposable QA collection only.
	beforeLimaYear := time.Date(2026, 1, 1, 4, 59, 59, 999000000, time.UTC)
	afterLimaYear := time.Date(2026, 1, 1, 5, 0, 0, 0, time.UTC)
	if _, err = s.DB.Collection(surveyCollection).UpdateOne(ctx, bson.M{"orderId": "1"}, bson.M{"$set": bson.M{"response.at": beforeLimaYear}}); err != nil {
		t.Fatal(err)
	}
	if _, err = s.DB.Collection(surveyCollection).UpdateOne(ctx, bson.M{"orderId": "3"}, bson.M{"$set": bson.M{"response.at": afterLimaYear}}); err != nil {
		t.Fatal(err)
	}
	for _, bounds := range []struct {
		from, to time.Time
		total    int
		average  float64
	}{
		{afterLimaYear, time.Date(2026, 2, 1, 5, 0, 0, 0, time.UTC), 1, 1},
		{time.Date(2025, 12, 1, 5, 0, 0, 0, time.UTC), afterLimaYear, 1, 5},
		{time.Date(2025, 12, 29, 5, 0, 0, 0, time.UTC), time.Date(2026, 1, 5, 5, 0, 0, 0, time.UTC), 2, 3},
		{afterLimaYear, time.Date(2027, 1, 1, 5, 0, 0, 0, time.UTC), 1, 1},
	} {
		got, e := s.SurveyStatistics(ctx, bounds.from, bounds.to)
		if e != nil || got.Total != bounds.total || got.Average != bounds.average {
			t.Fatal("Lima response-time boundary", got, e)
		}
	}
	tracking, err := s.SurveyTracking(ctx, 1)
	if err != nil || len(tracking) != 3 {
		t.Fatal("tracking", err)
	}
	for _, entry := range tracking {
		if entry.TokenCipher != "" || entry.TokenHash != "" || entry.Response != nil && entry.Response.Comment != "" {
			t.Fatal("tracking exposed secret/comment")
		}
	}
}
