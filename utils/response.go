package utils

import (
	"github.com/gofiber/fiber/v2"
)

type APIResponse struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   any    `json:"error"`
	Data    any    `json:"data"`
}

func SendSuccess(c *fiber.Ctx, code int, message string, data any) error {
	return c.Status(code).JSON(APIResponse{
		Success: true,
		Code:    code,
		Message: message,
		Error:   nil,
		Data:    data,
	})
}

func SendError(c *fiber.Ctx, code int, massage string, errorDetail any) error {
	return c.Status(code).JSON(APIResponse{
		Success: false,
		Code:    code,
		Message: massage,
		Error:   errorDetail,
		Data:    nil,
	})
}

func GetPaginationParams(c *fiber.Ctx) (int, int) {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return page, limit
}
