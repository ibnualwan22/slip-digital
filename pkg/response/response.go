package response

import (
	"github.com/labstack/echo/v4"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c echo.Context, statusCode int, message string, data interface{}) error {
	return c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(c echo.Context, statusCode int, message string) error {
	return c.JSON(statusCode, APIResponse{
		Success: false,
		Message: message,
	})
}
