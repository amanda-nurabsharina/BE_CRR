package postgres

import (
	"context"

	"be_crr/internal/domain/entity"
	"be_crr/internal/domain/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AudioRepo struct {
	db *gorm.DB
}

func NewAudioRepo(db *gorm.DB) repository.AudioRepository {
	return &AudioRepo{db: db}
}

func (r *AudioRepo) CreateEvidence(ctx context.Context, audio *entity.AudioEvidence) error {
	return r.db.WithContext(ctx).Create(audio).Error
}

func (r *AudioRepo) GetEvidenceByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.AudioEvidence, error) {
	var audio entity.AudioEvidence
	err := r.db.WithContext(ctx).
		Preload("Transcripts").
		Preload("Transcripts.Segments", func(db *gorm.DB) *gorm.DB {
			return db.Order("sequence_order ASC")
		}).
		First(&audio, "tenant_id = ? AND id = ?", tenantID, id).Error
	if err != nil {
		return nil, err
	}
	return &audio, nil
}

func (r *AudioRepo) UpdateEvidence(ctx context.Context, audio *entity.AudioEvidence) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", audio.TenantID, audio.ID).Save(audio).Error
}

func (r *AudioRepo) CreateChunk(ctx context.Context, chunk *entity.AudioChunk) error {
	return r.db.WithContext(ctx).Create(chunk).Error
}

func (r *AudioRepo) GetChunks(ctx context.Context, evidenceID uuid.UUID) ([]entity.AudioChunk, error) {
	var chunks []entity.AudioChunk
	err := r.db.WithContext(ctx).Where("audio_evidence_id = ?", evidenceID).Order("sequence_number ASC").Find(&chunks).Error
	return chunks, err
}

func (r *AudioRepo) DeleteEvidence(ctx context.Context, tenantID, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&entity.AudioEvidence{}).Error
}

type TranscriptRepo struct {
	db *gorm.DB
}

func NewTranscriptRepo(db *gorm.DB) repository.TranscriptRepository {
	return &TranscriptRepo{db: db}
}

func (r *TranscriptRepo) Create(ctx context.Context, t *entity.Transcript) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *TranscriptRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.Transcript, error) {
	var t entity.Transcript
	err := r.db.WithContext(ctx).
		Preload("Segments", func(db *gorm.DB) *gorm.DB {
			return db.Order("sequence_order ASC")
		}).
		First(&t, "tenant_id = ? AND id = ?", tenantID, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TranscriptRepo) GetByAudioID(ctx context.Context, tenantID, audioID uuid.UUID) (*entity.Transcript, error) {
	var t entity.Transcript
	err := r.db.WithContext(ctx).
		Preload("Segments", func(db *gorm.DB) *gorm.DB {
			return db.Order("sequence_order ASC")
		}).
		First(&t, "tenant_id = ? AND audio_evidence_id = ?", tenantID, audioID).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TranscriptRepo) Update(ctx context.Context, t *entity.Transcript) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", t.TenantID, t.ID).Save(t).Error
}

func (r *TranscriptRepo) SaveSegments(ctx context.Context, transcriptID uuid.UUID, segments []entity.TranscriptSegment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Replace all segments atomically
		if err := tx.Where("transcript_id = ?", transcriptID).Delete(&entity.TranscriptSegment{}).Error; err != nil {
			return err
		}
		if len(segments) > 0 {
			for i := range segments {
				segments[i].TranscriptID = transcriptID
			}
			if err := tx.Create(&segments).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

type ReportRepo struct {
	db *gorm.DB
}

func NewReportRepo(db *gorm.DB) repository.ReportRepository {
	return &ReportRepo{db: db}
}

func (r *ReportRepo) Create(ctx context.Context, rep *entity.CourtReport) error {
	return r.db.WithContext(ctx).Create(rep).Error
}

func (r *ReportRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.CourtReport, error) {
	var rep entity.CourtReport
	err := r.db.WithContext(ctx).
		Preload("Signatures").
		Preload("Session").
		First(&rep, "tenant_id = ? AND id = ?", tenantID, id).Error
	if err != nil {
		return nil, err
	}
	return &rep, nil
}

func (r *ReportRepo) GetBySessionID(ctx context.Context, tenantID, sessionID uuid.UUID) ([]entity.CourtReport, error) {
	var reps []entity.CourtReport
	err := r.db.WithContext(ctx).
		Preload("Signatures").
		Where("tenant_id = ? AND court_session_id = ?", tenantID, sessionID).
		Find(&reps).Error
	return reps, err
}

func (r *ReportRepo) Update(ctx context.Context, rep *entity.CourtReport) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", rep.TenantID, rep.ID).Save(rep).Error
}

func (r *ReportRepo) CreateSignature(ctx context.Context, sig *entity.ReportSignature) error {
	return r.db.WithContext(ctx).Create(sig).Error
}

func (r *ReportRepo) GetSignatureByToken(ctx context.Context, token string) (*entity.ReportSignature, error) {
	var sig entity.ReportSignature
	err := r.db.WithContext(ctx).First(&sig, "verification_token = ?", token).Error
	if err != nil {
		return nil, err
	}
	return &sig, nil
}

type AuditRepo struct {
	db *gorm.DB
}

func NewAuditRepo(db *gorm.DB) repository.AuditRepository {
	return &AuditRepo{db: db}
}

func (r *AuditRepo) Create(ctx context.Context, log *entity.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

