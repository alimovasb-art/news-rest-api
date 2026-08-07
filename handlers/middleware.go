package handlers

import (
	"news-restapi/utils"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
)

var jwtSecretKey = []byte(os.Getenv("JWT_SECRET"))

func AuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return utils.SendError(c, fiber.StatusUnauthorized, "Missing Authorization header", nil)
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return utils.SendError(c, fiber.StatusUnauthorized, "Invalid Authorization format", nil)
	}

	tokenString := parts[1]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecretKey, nil
	})
	if err != nil || !token.Valid {
		return utils.SendError(c, fiber.StatusUnauthorized, "Invalid or expired token", nil)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return utils.SendError(c, fiber.StatusUnauthorized, "Invalid token claims", nil)
	}

	authorID := int(claims["author_id"].(float64))

	c.Locals("author_id", authorID)

	return c.Next()
}
