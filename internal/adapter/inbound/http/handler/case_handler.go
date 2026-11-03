package handler

import (
	"strconv"
	"time"

	"be_crr/internal/domain/entity"
	"be_crr/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CaseHandler struct {
	caseUsecase *usecase.CaseUsecase
}

func NewCaseHandler(caseUsecase *usecase.CaseUsecase) *CaseHandler {
	return &CaseHandler{caseUsecase: caseUsecase}
}

func (h *CaseHandler) ListCases(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	search := c.Query("search", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	cases, total, err := h.caseUsecase.ListCases(c.Context(), tenantID, search, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    cases,
		"meta": fiber.Map{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

type CreateCaseReq struct {
	CaseNumber       string `json:"case_number"`
	CourtName        string `json:"court_name"`
	CaseCategory     string `json:"case_category"`
	ClientName       string `json:"client_name"`
	CounterPartyName string `json:"counter_party_name"`
	JudgePanelNames  string `json:"judge_panel_names"`
	ProsecutorNames  string `json:"prosecutor_names"`
	ClerkName        string `json:"clerk_name"`
	Notes            string `json:"notes"`
}

func (h *CaseHandler) CreateCase(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	var req CreateCaseReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request body"})
	}

	if req.CaseNumber == "" || req.CourtName == "" || req.ClientName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Case number, court name and client name are required"})
	}

	courtCase := &entity.CourtCase{
		TenantID:         tenantID,
		CaseNumber:       req.CaseNumber,
		CourtName:        req.CourtName,
		CaseCategory:     req.CaseCategory,
		ClientName:       req.ClientName,
		CounterPartyName: req.CounterPartyName,
		JudgePanelNames:  req.JudgePanelNames,
		ProsecutorNames:  req.ProsecutorNames,
		ClerkName:        req.ClerkName,
		Notes:            req.Notes,
	}

	if err := h.caseUsecase.CreateCase(c.Context(), courtCase); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    courtCase,
	})
}

func (h *CaseHandler) GetCase(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	caseID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid case ID"})
	}

	courtCase, err := h.caseUsecase.GetCase(c.Context(), tenantID, caseID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Case not found"})
	}

	return c.JSON(fiber.Map{"success": true, "data": courtCase})
}

func (h *CaseHandler) ManualPurge(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	actorID := c.Locals("user_id").(uuid.UUID)
	caseID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid case ID"})
	}

	ip := c.IP()
	if err := h.caseUsecase.ManualPurge(c.Context(), tenantID, caseID, actorID, ip); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Court case, all associated audio files, transcripts, and records have been permanently purged.",
	})
}

type CreateSessionReq struct {
	SessionTitle string `json:"session_title"`
	SessionDate  string `json:"session_date"`
}

func (h *CaseHandler) CreateSession(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	caseID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid case ID"})
	}

	var req CreateSessionReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request body"})
	}

	sessDate := time.Now()
	if req.SessionDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.SessionDate); err == nil {
			sessDate = parsed
		}
	}

	session := &entity.CourtSession{
		TenantID:     tenantID,
		CourtCaseID:  caseID,
		SessionTitle: req.SessionTitle,
		SessionDate:  sessDate,
		Status:       entity.SessionStatusScheduled,
	}
	if session.SessionTitle == "" {
		session.SessionTitle = "Sidang Agenda Pemeriksaan"
	}

	if err := h.caseUsecase.CreateSession(c.Context(), session); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    session,
	})
}

func (h *CaseHandler) GetSession(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	sessID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid session ID"})
	}

	sess, err := h.caseUsecase.GetSession(c.Context(), tenantID, sessID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Session not found"})
	}

	return c.JSON(fiber.Map{"success": true, "data": sess})
}

