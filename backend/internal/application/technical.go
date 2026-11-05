package application

import (
	"context"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
)

type Technical interface {
	Ping(context.Context) error
	PutScan(context.Context, string, string, domain.ScanInput) (domain.Scan, bool, error)
	Scans(context.Context, string, domain.Page) (domain.List[domain.Scan], error)
	PutEvidence(context.Context, string, string, domain.EvidenceInput) (domain.Evidence, bool, error)
	EvidenceList(context.Context, string, domain.Page) (domain.List[domain.Evidence], error)
	File(context.Context, string, string) (domain.Evidence, []byte, error)
}
