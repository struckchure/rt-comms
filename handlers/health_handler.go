package handlers

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/samber/do/v2"
)

type IHealthHandler interface {
	Health(c *echo.Context) error
}

type HealthHandler struct{}

func NewHealthHandler(_ do.Injector) (IHealthHandler, error) {
	return &HealthHandler{}, nil
}

func (h *HealthHandler) Health(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
