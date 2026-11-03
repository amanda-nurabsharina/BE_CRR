package usecase

import (
	"context"
	"time"

	"be_crr/internal/domain"
	"be_crr/internal/domain/entity"
	"be_crr/internal/domain/repository"
	"be_crr/internal/infrastructure/storage"

	"github.com/google/uuid"
)

type CaseUsecase struct {
	caseRepo    repository.CaseRepository
	sessionRepo repository.SessionRepository
	audioRepo   repository.AudioRepository
	subRepo     repository.SubscriptionRepository
	storage     storage.Storage
	auditRepo   repository.AuditRepository
}

func NewCaseUsecase(
	caseRepo repository.CaseRepository,
	sessionRepo repository.SessionRepository,
	audioRepo repository.AudioRepository,
	subRepo repository.SubscriptionRepository,
	storage storage.Storage,
	auditRepo repository.AuditRepository,
) *CaseUsecase {
	return &CaseUsecase{
		caseRepo:    caseRepo,
		sessionRepo: sessionRepo,
		audioRepo:   audioRepo,
		subRepo:     subRepo,
		storage:     storage,
		auditRepo:   auditRepo,
	}
}

func (u *CaseUsecase) ListCases(ctx context.Context, tenantID uuid.UUID, search string, page, limit int) ([]entity.CourtCase, int64, error) {
	return u.caseRepo.ListByTenant(ctx, tenantID, search, page, limit)
}

func (u *CaseUsecase) CreateCase(ctx context.Context, c *entity.CourtCase) error {
	return u.caseRepo.Create(ctx, c)
}

func (u *CaseUsecase) GetCase(ctx context.Context, tenantID, caseID uuid.UUID) (*entity.CourtCase, error) {
	c, err := u.caseRepo.GetByID(ctx, tenantID, caseID)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	return c, nil
}

func (u *CaseUsecase) UpdateCase(ctx context.Context, c *entity.CourtCase) error {
	c.UpdatedAt = time.Now()
	return u.caseRepo.Update(ctx, c)
}

// ManualPurge removes a case and ALL associated media files, chunks, and database rows irreversibly
func (u *CaseUsecase) ManualPurge(ctx context.Context, tenantID, caseID, actorID uuid.UUID, ip string) error {
	courtCase, err := u.caseRepo.GetByID(ctx, tenantID, caseID)
	if err != nil {
		return domain.ErrNotFound
	}

	var freedBytes int64 = 0

	// Delete physical audio files
	for _, sess := range courtCase.Sessions {
		for _, audio := range sess.AudioEvidences {
			if audio.StoragePath != "" {
				_ = u.storage.DeleteFile(audio.StoragePath)
				freedBytes += audio.FileSizeBytes
			}
		}
		for _, rep := range sess.Reports {
			if rep.PdfStoragePath != "" {
				_ = u.storage.DeleteFile(rep.PdfStoragePath)
			}
		}
	}

	// Delete from DB (cascading deletes sessions, evidences, transcripts, reports)
	if err := u.caseRepo.Delete(ctx, tenantID, caseID); err != nil {
		return err
	}

	// Deduct freed storage from subscription
	sub, err := u.subRepo.GetByTenantID(ctx, tenantID)
	if err == nil {
		if sub.UsedStorageBytes >= freedBytes {
			sub.UsedStorageBytes -= freedBytes
		} else {
			sub.UsedStorageBytes = 0
		}
		_ = u.subRepo.Update(ctx, sub)
	}

	// Audit log entry
	_ = u.auditRepo.Create(ctx, &entity.AuditLog{
		TenantID:     &tenantID,
		ActorID:      actorID,
		Action:       "MANUAL_PURGE",
		ResourceType: "COURT_CASE",
		ResourceID:   caseID,
		IPAddress:    ip,
		DetailsJSON:  `{"message": "Irreversible manual purge completed"}`,
	})

	return nil
}

func (u *CaseUsecase) CreateSession(ctx context.Context, s *entity.CourtSession) error {
	return u.sessionRepo.Create(ctx, s)
}

func (u *CaseUsecase) GetSession(ctx context.Context, tenantID, sessionID uuid.UUID) (*entity.CourtSession, error) {
	s, err := u.sessionRepo.GetByID(ctx, tenantID, sessionID)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	return s, nil
}

