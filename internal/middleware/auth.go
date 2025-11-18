package middleware

import (
	"strings"

	"github.com/MatiasTelo/stockgo/internal/service"
	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware valida el token de autenticación usando el servicio de auth
func AuthMiddleware(authService *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Extraer token del header Authorization
		token := extractToken(c)
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// 2. Validar token (con caché en Redis)
		user, err := authService.ValidateToken(c.Context(), token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// 3. Almacenar información del usuario en el contexto
		c.Locals("token", token)
		c.Locals("user_id", user.ID)
		c.Locals("username", user.Username)
		c.Locals("user", user)

		// 4. Continuar con el siguiente handler
		return c.Next()
	}
}

// extractToken extrae el token del header Authorization
func extractToken(c *fiber.Ctx) string {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Formato esperado: "Bearer <token>"
	if strings.HasPrefix(strings.ToUpper(authHeader), "BEARER ") {
		return authHeader[7:]
	}

	return ""
}
