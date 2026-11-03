package handler

import (
	"strconv"

	"be_crr/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type SubscriptionHandler struct {
	subUsecase *usecase.SubscriptionUsecase
}

func NewSubscriptionHandler(subUsecase *usecase.SubscriptionUsecase) *SubscriptionHandler {
	return &SubscriptionHandler{subUsecase: subUsecase}
}

func (h *SubscriptionHandler) GetCurrent(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	usage, err := h.subUsecase.GetCurrentSubscription(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Subscription not found"})
	}
	return c.JSON(fiber.Map{"success": true, "data": usage})
}

func (h *SubscriptionHandler) SubmitManualPayment(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)

	amountStr := c.FormValue("amount")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Valid amount is required"})
	}

	bankDest := c.FormValue("bank_destination")
	senderName := c.FormValue("bank_sender_name")
	transferDate := c.FormValue("transfer_date")

	fileHeader, err := c.FormFile("proof_image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Transfer receipt image is required"})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Failed to read receipt file"})
	}
	defer file.Close()

	payment, err := h.subUsecase.SubmitManualPayment(c.Context(), usecase.SubmitPaymentDTO{
		TenantID:        tenantID,
		Amount:          amount,
		BankDestination: bankDest,
		BankSenderName:  senderName,
		TransferDate:    transferDate,
		ProofFilename:   fileHeader.Filename,
		ProofReader:     file,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    payment,
		"message": "Payment submitted. Awaiting Super Admin manual verification.",
	})
}

func (h *SubscriptionHandler) GetPaymentHistory(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	payments, err := h.subUsecase.GetPaymentHistory(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": payments})
}

