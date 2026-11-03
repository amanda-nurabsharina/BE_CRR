package handler

import (
	"be_crr/internal/domain"
	"be_crr/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authUsecase *usecase.AuthUsecase
}

func NewAuthHandler(authUsecase *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
}

type RegisterReq struct {
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	PhoneNumber string `json:"phone_number"`
	LawFirmName string `json:"law_firm_name"`
	BarNumber   string `json:"bar_number"`
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request body"})
	}

	if req.Email == "" || req.Password == "" || req.FullName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Full name, email and password are required"})
	}

	res, err := h.authUsecase.Register(c.Context(), usecase.RegisterDTO{
		FullName:    req.FullName,
		Email:       req.Email,
		Password:    req.Password,
		PhoneNumber: req.PhoneNumber,
		LawFirmName: req.LawFirmName,
		BarNumber:   req.BarNumber,
	})
	if err != nil {
		if err == domain.ErrEmailAlreadyExists {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "error": "Email is already registered"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    res,
	})
}

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request body"})
	}

	res, err := h.authUsecase.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		if err == domain.ErrInvalidCredentials {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "Invalid email or password"})
		}
		if err == domain.ErrTenantInactive {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Account is suspended or inactive"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    res,
	})
}

func (h *AuthHandler) GetMe(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "Unauthorized"})
	}

	user, err := h.authUsecase.GetMe(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "User not found"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    user,
	})
}

type UpdateProfileReq struct {
	FullName    string `json:"full_name"`
	PhoneNumber string `json:"phone_number"`
	BarNumber   string `json:"bar_number"`
	LawFirmName string `json:"law_firm_name"`
	Password    string `json:"password"`
}

func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "Unauthorized"})
	}

	var req UpdateProfileReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request body"})
	}

	user, err := h.authUsecase.UpdateProfile(c.Context(), usecase.UpdateProfileDTO{
		UserID:      userID,
		FullName:    req.FullName,
		PhoneNumber: req.PhoneNumber,
		BarNumber:   req.BarNumber,
		LawFirmName: req.LawFirmName,
		Password:    req.Password,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    user,
		"message": "Profil berhasil diperbarui",
	})
}

