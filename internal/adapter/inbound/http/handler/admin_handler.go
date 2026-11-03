package handler

import (
	"strconv"

	"be_crr/internal/domain/entity"
	"be_crr/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AdminHandler struct {
	adminUsecase *usecase.AdminUsecase
}

func NewAdminHandler(adminUsecase *usecase.AdminUsecase) *AdminHandler {
	return &AdminHandler{adminUsecase: adminUsecase}
}

func (h *AdminHandler) GetMetrics(c *fiber.Ctx) error {
	metrics, err := h.adminUsecase.GetMetrics(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": metrics})
}

func (h *AdminHandler) ListTenants(c *fiber.Ctx) error {
	status := c.Query("status", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	tenants, total, err := h.adminUsecase.ListTenants(c.Context(), status, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    tenants,
		"meta": fiber.Map{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

type UpdateTenantStatusReq struct {
	Status entity.TenantStatus `json:"status"`
}

func (h *AdminHandler) UpdateTenantStatus(c *fiber.Ctx) error {
	idParam := c.Params("id")
	tenantID, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid tenant ID"})
	}

	var req UpdateTenantStatusReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request body"})
	}

	if err := h.adminUsecase.UpdateTenantStatus(c.Context(), tenantID, req.Status); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Tenant status updated successfully"})
}

func (h *AdminHandler) ListPendingPayments(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	payments, total, err := h.adminUsecase.ListPendingPayments(c.Context(), page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    payments,
		"meta": fiber.Map{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

func (h *AdminHandler) ApprovePayment(c *fiber.Ctx) error {
	adminID := c.Locals("user_id").(uuid.UUID)
	paymentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid payment ID"})
	}

	if err := h.adminUsecase.ApprovePayment(c.Context(), adminID, paymentID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Payment approved and tenant activated with new quotas"})
}

type RejectPaymentReq struct {
	Reason string `json:"reason"`
}

func (h *AdminHandler) RejectPayment(c *fiber.Ctx) error {
	adminID := c.Locals("user_id").(uuid.UUID)
	paymentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid payment ID"})
	}

	var req RejectPaymentReq
	_ = c.BodyParser(&req)

	if err := h.adminUsecase.RejectPayment(c.Context(), adminID, paymentID, req.Reason); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Payment rejected"})
}

type AdjustQuotaReq struct {
	AddMinutes   int `json:"add_minutes"`
	AddStorageGB int `json:"add_storage_gb"`
}

func (h *AdminHandler) AdjustQuota(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid tenant ID"})
	}

	var req AdjustQuotaReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid body"})
	}

	if err := h.adminUsecase.AdjustQuota(c.Context(), tenantID, req.AddMinutes, req.AddStorageGB); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Quota adjusted successfully"})
}

