package usecase

import (
	"context"
	"fmt"
	"time"

	"be_crr/internal/domain"
	"be_crr/internal/domain/entity"
	"be_crr/internal/domain/repository"
	"be_crr/internal/infrastructure/hasher"
	"be_crr/internal/infrastructure/storage"

	"github.com/google/uuid"
)

type RecordingUsecase struct {
	audioRepo   repository.AudioRepository
	sessionRepo repository.SessionRepository
	subRepo     repository.SubscriptionRepository
	storage     storage.Storage
}

func NewRecordingUsecase(
	audioRepo repository.AudioRepository,
	sessionRepo repository.SessionRepository,
	subRepo repository.SubscriptionRepository,
	storage storage.Storage,
) *RecordingUsecase {
	return &RecordingUsecase{
		audioRepo:   audioRepo,
		sessionRepo: sessionRepo,
		subRepo:     subRepo,
		storage:     storage,
	}
}

type InitRecordingDTO struct {
	TenantID       uuid.UUID
	CourtSessionID uuid.UUID
	SourceType     entity.AudioSourceType
	MimeType       string
}

func (u *RecordingUsecase) InitRecording(ctx context.Context, dto InitRecordingDTO) (*entity.AudioEvidence, error) {
	// Verify session exists
	sess, err := u.sessionRepo.GetByID(ctx, dto.TenantID, dto.CourtSessionID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	// Update session status to recording
	sess.Status = entity.SessionStatusRecording
	now := time.Now()
	if sess.StartTime == nil {
		sess.StartTime = &now
	}
	_ = u.sessionRepo.Update(ctx, sess)

	evidence := &entity.AudioEvidence{
		TenantID:            dto.TenantID,
		CourtSessionID:      dto.CourtSessionID,
		SourceType:          dto.SourceType,
		OriginalFilename:    fmt.Sprintf("recording_%s_%d.webm", dto.CourtSessionID.String()[:8], time.Now().Unix()),
		StoragePath:         "",
		FileSizeBytes:       0,
		DurationSeconds:     0,
		MimeType:            dto.MimeType,
		Sha256Hash:          "",
		EncryptionAlgorithm: "AES-256-GCM",
		IngestStatus:        "ingesting",
	}

	if err := u.audioRepo.CreateEvidence(ctx, evidence); err != nil {
		return nil, err
	}

	return evidence, nil
}

type SaveChunkDTO struct {
	EvidenceID     uuid.UUID
	SequenceNumber int
	CRC32Checksum  string
	ChunkData      []byte
}

func (u *RecordingUsecase) SaveChunk(ctx context.Context, dto SaveChunkDTO) error {
	computedCRC := hasher.ComputeCRC32(dto.ChunkData)

	savedPath, err := u.storage.AppendChunk(dto.EvidenceID, dto.SequenceNumber, dto.ChunkData)
	if err != nil {
		return fmt.Errorf("failed to save chunk: %w", err)
	}

	chunk := &entity.AudioChunk{
		AudioEvidenceID: dto.EvidenceID,
		SequenceNumber:  dto.SequenceNumber,
		ChunkSizeBytes:  len(dto.ChunkData),
		Crc32Checksum:   computedCRC,
		StoragePath:     savedPath,
		SyncedAt:        time.Now(),
	}

	return u.audioRepo.CreateChunk(ctx, chunk)
}

func (u *RecordingUsecase) FinalizeSession(ctx context.Context, tenantID, evidenceID uuid.UUID, totalChunks int, estimatedDurationSec int) (*entity.AudioEvidence, error) {
	evidence, err := u.audioRepo.GetEvidenceByID(ctx, tenantID, evidenceID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	destFilename := fmt.Sprintf("audio_%s.webm", evidence.ID.String())
	finalPath, totalBytes, err := u.storage.AssembleChunks(evidence.ID, totalChunks, destFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to assemble recording chunks: %w", err)
	}

	masterHash, err := hasher.ComputeFileSHA256(finalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to compute master SHA-256: %w", err)
	}

	evidence.StoragePath = finalPath
	evidence.FileSizeBytes = totalBytes
	evidence.DurationSeconds = estimatedDurationSec
	evidence.Sha256Hash = masterHash
	evidence.IngestStatus = "ready"

	if err := u.audioRepo.UpdateEvidence(ctx, evidence); err != nil {
		return nil, err
	}

	// Update session status
	sess, err := u.sessionRepo.GetByID(ctx, tenantID, evidence.CourtSessionID)
	if err == nil {
		sess.Status = entity.SessionStatusCompleted
		now := time.Now()
		sess.EndTime = &now
		_ = u.sessionRepo.Update(ctx, sess)
	}

	// Deduct storage
	sub, err := u.subRepo.GetByTenantID(ctx, tenantID)
	if err == nil {
		sub.UsedStorageBytes += totalBytes
		_ = u.subRepo.Update(ctx, sub)
	}

	return evidence, nil
}

