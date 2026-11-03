package usecase

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"be_crr/internal/domain"
	"be_crr/internal/domain/entity"
	"be_crr/internal/domain/repository"
	"be_crr/internal/infrastructure/hasher"
	"be_crr/internal/infrastructure/storage"

	"github.com/google/uuid"
)

type IngestUsecase struct {
	audioRepo   repository.AudioRepository
	sessionRepo repository.SessionRepository
	subRepo     repository.SubscriptionRepository
	storage     storage.Storage
}

func NewIngestUsecase(
	audioRepo repository.AudioRepository,
	sessionRepo repository.SessionRepository,
	subRepo repository.SubscriptionRepository,
	storage storage.Storage,
) *IngestUsecase {
	return &IngestUsecase{
		audioRepo:   audioRepo,
		sessionRepo: sessionRepo,
		subRepo:     subRepo,
		storage:     storage,
	}
}

type UploadMediaDTO struct {
	TenantID       uuid.UUID
	CourtSessionID uuid.UUID
	OriginalName   string
	MimeType       string
	FileReader     io.Reader
	FileSizeBytes  int64
}

func (u *IngestUsecase) UploadMedia(ctx context.Context, dto UploadMediaDTO) (*entity.AudioEvidence, error) {
	// Verify session
	sess, err := u.sessionRepo.GetByID(ctx, dto.TenantID, dto.CourtSessionID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	// Check storage quota
	sub, err := u.subRepo.GetByTenantID(ctx, dto.TenantID)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	if sub.UsedStorageBytes+dto.FileSizeBytes > sub.MaxStorageBytes {
		return nil, domain.ErrQuotaExceeded
	}

	ext := strings.ToLower(filepath.Ext(dto.OriginalName))
	sourceType := entity.SourceUploadAudio
	if ext == ".mp4" {
		sourceType = entity.SourceExtractedMP4
	}

	subDir := fmt.Sprintf("evidences/%s", dto.TenantID.String())
	filename := fmt.Sprintf("evidence_%d_%s", time.Now().Unix(), dto.OriginalName)
	savedPath, savedBytes, err := u.storage.SaveFile(subDir, filename, dto.FileReader)
	if err != nil {
		return nil, fmt.Errorf("failed to save media file: %w", err)
	}

	// Calculate SHA-256
	fileHash, err := hasher.ComputeFileSHA256(savedPath)
	if err != nil {
		_ = u.storage.DeleteFile(savedPath)
		return nil, fmt.Errorf("failed to compute file hash: %w", err)
	}

	evidence := &entity.AudioEvidence{
		TenantID:            dto.TenantID,
		CourtSessionID:      dto.CourtSessionID,
		SourceType:          sourceType,
		OriginalFilename:    dto.OriginalName,
		StoragePath:         savedPath,
		FileSizeBytes:       savedBytes,
		DurationSeconds:     120, // default placeholder or extracted duration
		MimeType:            dto.MimeType,
		Sha256Hash:          fileHash,
		EncryptionAlgorithm: "AES-256-GCM",
		IngestStatus:        "ready",
	}

	if err := u.audioRepo.CreateEvidence(ctx, evidence); err != nil {
		return nil, err
	}

	// Update session status
	sess.Status = entity.SessionStatusCompleted
	now := time.Now()
	sess.EndTime = &now
	_ = u.sessionRepo.Update(ctx, sess)

	// Deduct storage
	sub.UsedStorageBytes += savedBytes
	_ = u.subRepo.Update(ctx, sub)

	return evidence, nil
}

func (u *IngestUsecase) GetAudioEvidence(ctx context.Context, tenantID, evidenceID uuid.UUID) (*entity.AudioEvidence, error) {
	return u.audioRepo.GetEvidenceByID(ctx, tenantID, evidenceID)
}

