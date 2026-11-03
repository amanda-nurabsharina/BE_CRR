package handler

import (
	"be_crr/internal/domain/entity"
	"be_crr/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TranscriptHandler struct {
	transcriptUsecase *usecase.TranscriptUsecase
}

func NewTranscriptHandler(transcriptUsecase *usecase.TranscriptUsecase) *TranscriptHandler {
	return &TranscriptHandler{transcriptUsecase: transcriptUsecase}
}

type TriggerASRReq struct {
	DurationSeconds int `json:"duration_seconds"`
}

func (h *TranscriptHandler) TriggerASR(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	audioID, err := uuid.Parse(c.Params("audioId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid audio ID"})
	}

	var req TriggerASRReq
	_ = c.BodyParser(&req)

	t, err := h.transcriptUsecase.TriggerASR(c.Context(), tenantID, audioID, req.DurationSeconds)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    t,
		"message": "Indonesian ASR transcription generated successfully",
	})
}

func (h *TranscriptHandler) GetTranscript(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	audioID, err := uuid.Parse(c.Params("audioId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid audio ID"})
	}

	t, err := h.transcriptUsecase.GetTranscript(c.Context(), tenantID, audioID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": t})
}

type SaveSegmentsReq struct {
	TranscriptID string                     `json:"transcript_id"`
	Segments     []entity.TranscriptSegment `json:"segments"`
}

func (h *TranscriptHandler) SaveSegments(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	var req SaveSegmentsReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request body"})
	}

	tID, err := uuid.Parse(req.TranscriptID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid transcript ID"})
	}

	updated, err := h.transcriptUsecase.SaveSegments(c.Context(), usecase.UpdateSegmentsDTO{
		TenantID:     tenantID,
		TranscriptID: tID,
		Segments:     req.Segments,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    updated,
		"message": "Transcript segments updated and version incremented",
	})
}

