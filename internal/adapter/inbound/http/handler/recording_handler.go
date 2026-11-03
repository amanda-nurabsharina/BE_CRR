package handler

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"

	"be_crr/internal/domain/entity"
	"be_crr/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type RecordingHandler struct {
	recUsecase    *usecase.RecordingUsecase
	ingestUsecase *usecase.IngestUsecase
}

func NewRecordingHandler(recUsecase *usecase.RecordingUsecase, ingestUsecase *usecase.IngestUsecase) *RecordingHandler {
	return &RecordingHandler{
		recUsecase:    recUsecase,
		ingestUsecase: ingestUsecase,
	}
}

type InitRecordingReq struct {
	CourtSessionID string                 `json:"court_session_id"`
	SourceType     entity.AudioSourceType `json:"source_type"`
	MimeType       string                 `json:"mime_type"`
}

func (h *RecordingHandler) InitSession(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	var req InitRecordingReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request body"})
	}

	sessionID, err := uuid.Parse(req.CourtSessionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid session ID"})
	}

	mimeType := req.MimeType
	if mimeType == "" {
		mimeType = "audio/webm;codecs=opus"
	}

	evidence, err := h.recUsecase.InitRecording(c.Context(), usecase.InitRecordingDTO{
		TenantID:       tenantID,
		CourtSessionID: sessionID,
		SourceType:     req.SourceType,
		MimeType:       mimeType,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    evidence,
	})
}

type UploadChunkReq struct {
	EvidenceID     string `json:"evidence_id"`
	SequenceNumber int    `json:"sequence_number"`
	CRC32Checksum  string `json:"crc32_checksum"`
	Base64Data     string `json:"base64_data"`
}

func (h *RecordingHandler) UploadChunk(c *fiber.Ctx) error {
	var req UploadChunkReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid body"})
	}

	evidenceID, err := uuid.Parse(req.EvidenceID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid evidence ID"})
	}

	chunkData, err := base64.StdEncoding.DecodeString(req.Base64Data)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid base64 chunk payload"})
	}

	err = h.recUsecase.SaveChunk(c.Context(), usecase.SaveChunkDTO{
		EvidenceID:     evidenceID,
		SequenceNumber: req.SequenceNumber,
		CRC32Checksum:  req.CRC32Checksum,
		ChunkData:      chunkData,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success":         true,
		"sequence_number": req.SequenceNumber,
		"message":         "Chunk synced and verified",
	})
}

type FinalizeSessionReq struct {
	TotalChunks          int `json:"total_chunks"`
	EstimatedDurationSec int `json:"estimated_duration_sec"`
}

func (h *RecordingHandler) FinalizeSession(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	evidenceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid evidence ID"})
	}

	var req FinalizeSessionReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request body"})
	}

	if req.TotalChunks <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Total chunks must be > 0"})
	}

	evidence, err := h.recUsecase.FinalizeSession(c.Context(), tenantID, evidenceID, req.TotalChunks, req.EstimatedDurationSec)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    evidence,
		"message": "Recording session finalized. Master SHA-256 generated.",
	})
}

func (h *RecordingHandler) UploadMediaFile(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	sessionIDStr := c.FormValue("court_session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid court session ID"})
	}

	fileHeader, err := c.FormFile("media_file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Media file is required"})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Failed to read file"})
	}
	defer file.Close()

	evidence, err := h.ingestUsecase.UploadMedia(c.Context(), usecase.UploadMediaDTO{
		TenantID:       tenantID,
		CourtSessionID: sessionID,
		OriginalName:   fileHeader.Filename,
		MimeType:       fileHeader.Header.Get("Content-Type"),
		FileReader:     file,
		FileSizeBytes:  fileHeader.Size,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    evidence,
		"message": "Media uploaded and cryptographic fingerprint calculated.",
	})
}

// StreamAudio handles authenticated Range-Request audio streaming without leaking S3/direct paths
func (h *RecordingHandler) StreamAudio(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	evidenceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid audio evidence ID"})
	}

	evidence, err := h.ingestUsecase.GetAudioEvidence(c.Context(), tenantID, evidenceID)
	if err != nil || evidence.StoragePath == "" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Audio file not found"})
	}

	if _, err := os.Stat(evidence.StoragePath); os.IsNotExist(err) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Physical audio file not found on disk"})
	}

	ext := strings.ToLower(filepath.Ext(evidence.StoragePath))
	switch ext {
	case ".m4a":
		c.Set("Content-Type", "audio/mp4")
	case ".mp4":
		c.Set("Content-Type", "video/mp4")
	case ".mp3":
		c.Set("Content-Type", "audio/mpeg")
	case ".wav":
		c.Set("Content-Type", "audio/wav")
	case ".webm":
		c.Set("Content-Type", "audio/webm")
	default:
		if evidence.MimeType != "" {
			c.Set("Content-Type", evidence.MimeType)
		}
	}
	c.Set("Accept-Ranges", "bytes")

	return c.SendFile(evidence.StoragePath)
}

