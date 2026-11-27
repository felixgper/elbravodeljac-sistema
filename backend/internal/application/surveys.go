package application

import (
	"context"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"time"
)

type SurveyDeliveryReader interface {
	SurveyDelivery(context.Context, string) (domain.SurveyDelivery, error)
}

type Surveys interface {
	IssueSurvey(context.Context, domain.SurveyInvitation, bool) error
	SurveyByOrder(context.Context, string) (domain.SurveyInvitation, error)
	SurveyByHash(context.Context, string) (domain.SurveyInvitation, error)
	AnswerSurvey(context.Context, string, domain.SurveyResponse) error
	SurveyStatistics(context.Context, time.Time, time.Time) (domain.SurveyStats, error)
	SurveyTracking(context.Context, int) ([]domain.SurveyInvitation, error)
}
