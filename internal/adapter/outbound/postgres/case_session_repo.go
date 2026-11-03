package postgres

import (
	"context"

	"be_crr/internal/domain/entity"
	"be_crr/internal/domain/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CaseRepo struct {
	db *gorm.DB
}

func NewCaseRepo(db *gorm.DB) repository.CaseRepository {
	return &CaseRepo{db: db}
}

func (r *CaseRepo) Create(ctx context.Context, c *entity.CourtCase) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *CaseRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.CourtCase, error) {
	var courtCase entity.CourtCase
	err := r.db.WithContext(ctx).
		Preload("Sessions", func(db *gorm.DB) *gorm.DB {
			return db.Order("court_sessions.session_date ASC")
		}).
		Preload("Sessions.AudioEvidences").
		Preload("Sessions.Reports").
		First(&courtCase, "tenant_id = ? AND id = ?", tenantID, id).Error
	if err != nil {
		return nil, err
	}
	return &courtCase, nil
}

func (r *CaseRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, search string, page, limit int) ([]entity.CourtCase, int64, error) {
	var cases []entity.CourtCase
	var total int64

	q := r.db.WithContext(ctx).Model(&entity.CourtCase{}).Where("tenant_id = ?", tenantID)
	if search != "" {
		s := "%" + search + "%"
		q = q.Where("case_number ILIKE ? OR client_name ILIKE ? OR court_name ILIKE ?", s, s, s)
	}

	q.Count(&total)

	offset := (page - 1) * limit
	err := q.Preload("Sessions").Offset(offset).Limit(limit).Order("created_at DESC").Find(&cases).Error
	return cases, total, err
}

func (r *CaseRepo) Update(ctx context.Context, c *entity.CourtCase) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", c.TenantID, c.ID).Save(c).Error
}

func (r *CaseRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&entity.CourtCase{}).Error
}

type SessionRepo struct {
	db *gorm.DB
}

func NewSessionRepo(db *gorm.DB) repository.SessionRepository {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) Create(ctx context.Context, s *entity.CourtSession) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *SessionRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.CourtSession, error) {
	var session entity.CourtSession
	err := r.db.WithContext(ctx).
		Preload("AudioEvidences").
		Preload("AudioEvidences.Transcripts").
		Preload("Reports").
		First(&session, "tenant_id = ? AND id = ?", tenantID, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepo) ListByCase(ctx context.Context, tenantID, caseID uuid.UUID) ([]entity.CourtSession, error) {
	var sessions []entity.CourtSession
	err := r.db.WithContext(ctx).
		Preload("AudioEvidences").
		Where("tenant_id = ? AND court_case_id = ?", tenantID, caseID).
		Order("session_date ASC").
		Find(&sessions).Error
	return sessions, err
}

func (r *SessionRepo) Update(ctx context.Context, s *entity.CourtSession) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", s.TenantID, s.ID).Save(s).Error
}

