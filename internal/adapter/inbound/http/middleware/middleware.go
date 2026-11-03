package middleware

import (
	"strings"

	"be_crr/internal/domain/entity"
	"be_crr/internal/infrastructure/security"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func AuthMiddleware(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var tokenStr string
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenStr = parts[1]
			}
		}

		// Also allow token in query param (essential for HTML5 <audio> / <video> elements)
		if tokenStr == "" {
			tokenStr = c.Query("token")
		}

		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Missing authorization header or token parameter",
			})
		}

		claims, err := security.ValidateToken(tokenStr, jwtSecret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Token is invalid or expired",
			})
		}

		c.Locals("user_id", claims.UserID)
		if claims.TenantID != nil {
			c.Locals("tenant_id", *claims.TenantID)
		}
		c.Locals("role", claims.Role)
		c.Locals("email", claims.Email)

		return c.Next()
	}
}

func SuperAdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(entity.UserRole)
		if !ok || role != entity.RoleSuperAdmin {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error":   "Forbidden: Super Admin access required",
			})
		}
		return c.Next()
	}
}

func TenantUserOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		_, ok := c.Locals("tenant_id").(uuid.UUID)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error":   "Forbidden: Workspace context required",
			})
		}
		return c.Next()
	}
}

