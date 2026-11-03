package handler

import (
	"be_crr/internal/domain/entity"
	"be_crr/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ReportHandler struct {
	reportUsecase *usecase.ReportUsecase
}

func NewReportHandler(reportUsecase *usecase.ReportUsecase) *ReportHandler {
	return &ReportHandler{reportUsecase: reportUsecase}
}

type CreateReportReq struct {
	CourtSessionID string `json:"court_session_id"`
	Title          string `json:"title"`
}

func (h *ReportHandler) CreateReport(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	var req CreateReportReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request body"})
	}

	sessionID, err := uuid.Parse(req.CourtSessionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid session ID"})
	}

	report, err := h.reportUsecase.CreateReport(c.Context(), tenantID, sessionID, req.Title)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    report,
	})
}

func (h *ReportHandler) GetReport(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	reportID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid report ID"})
	}

	report, err := h.reportUsecase.GetReport(c.Context(), tenantID, reportID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Report not found"})
	}

	return c.JSON(fiber.Map{"success": true, "data": report})
}

func (h *ReportHandler) UpdateReport(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	reportID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid report ID"})
	}

	var report entity.CourtReport
	if err := c.BodyParser(&report); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request body"})
	}
	report.ID = reportID

	if err := h.reportUsecase.UpdateReport(c.Context(), tenantID, &report); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Report draft saved", "data": report})
}

type SignReportReq struct {
	SignerName         string `json:"signer_name"`
	SignerIdentifier   string `json:"signer_identifier"` // NIA Peradi / NIK
	SignerRole         string `json:"signer_role"`       // Penasihat Hukum / Notulis
	SignatureImageData string `json:"signature_image_data"`
}

func (h *ReportHandler) SignReport(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	reportID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid report ID"})
	}

	var req SignReportReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request body"})
	}

	if req.SignerName == "" || req.SignatureImageData == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Signer name and signature canvas data are required"})
	}

	sig, err := h.reportUsecase.SignReport(c.Context(), usecase.SignReportDTO{
		TenantID:           tenantID,
		ReportID:           reportID,
		SignerName:         req.SignerName,
		SignerIdentifier:   req.SignerIdentifier,
		SignerRole:         req.SignerRole,
		SignatureImageData: req.SignatureImageData,
		SignerIPAddress:    c.IP(),
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    sig,
		"message": "Report digitally signed and bound with SHA-256 integrity manifest",
	})
}

func (h *ReportHandler) PublicVerify(c *fiber.Ctx) error {
	token := c.Params("token")
	res, err := h.reportUsecase.PublicVerify(c.Context(), token)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": res})
}

