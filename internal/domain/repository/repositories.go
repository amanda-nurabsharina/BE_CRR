package repository

import (
	"context"

	"be_crr/internal/domain/entity"

	"github.com/google/uuid"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant *entity.Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error)
	GetByEmail(ctx context.Context, email string) (*entity.Tenant, error)
	ListAll(ctx context.Context, status string, page, limit int) ([]entity.Tenant, int64, error)
	Update(ctx context.Context, tenant *entity.Tenant) error
}

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
}

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *entity.Subscription) error
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*entity.Subscription, error)
	Update(ctx context.Context, sub *entity.Subscription) error
	CreatePayment(ctx context.Context, payment *entity.ManualPayment) error
	GetPaymentByID(ctx context.Context, id uuid.UUID) (*entity.ManualPayment, error)
	ListPayments(ctx context.Context, status string, page, limit int) ([]entity.ManualPayment, int64, error)
	UpdatePayment(ctx context.Context, payment *entity.ManualPayment) error
}

type CaseRepository interface {
	Create(ctx context.Context, c *entity.CourtCase) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.CourtCase, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, search string, page, limit int) ([]entity.CourtCase, int64, error)
	Update(ctx context.Context, c *entity.CourtCase) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
}

type SessionRepository interface {
	Create(ctx context.Context, s *entity.CourtSession) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.CourtSession, error)
	ListByCase(ctx context.Context, tenantID, caseID uuid.UUID) ([]entity.CourtSession, error)
	Update(ctx context.Context, s *entity.CourtSession) error
}

type AudioRepository interface {
	CreateEvidence(ctx context.Context, audio *entity.AudioEvidence) error
	GetEvidenceByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.AudioEvidence, error)
	UpdateEvidence(ctx context.Context, audio *entity.AudioEvidence) error
	CreateChunk(ctx context.Context, chunk *entity.AudioChunk) error
	GetChunks(ctx context.Context, evidenceID uuid.UUID) ([]entity.AudioChunk, error)
	DeleteEvidence(ctx context.Context, tenantID, id uuid.UUID) error
}

type TranscriptRepository interface {
	Create(ctx context.Context, t *entity.Transcript) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.Transcript, error)
	GetByAudioID(ctx context.Context, tenantID, audioID uuid.UUID) (*entity.Transcript, error)
	Update(ctx context.Context, t *entity.Transcript) error
	SaveSegments(ctx context.Context, transcriptID uuid.UUID, segments []entity.TranscriptSegment) error
}

type ReportRepository interface {
	Create(ctx context.Context, r *entity.CourtReport) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.CourtReport, error)
	GetBySessionID(ctx context.Context, tenantID, sessionID uuid.UUID) ([]entity.CourtReport, error)
	Update(ctx context.Context, r *entity.CourtReport) error
	CreateSignature(ctx context.Context, sig *entity.ReportSignature) error
	GetSignatureByToken(ctx context.Context, token string) (*entity.ReportSignature, error)
}

type AuditRepository interface {
	Create(ctx context.Context, log *entity.AuditLog) error
}

