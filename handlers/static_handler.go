package handlers

import (
	"github.com/labstack/echo/v5"
	"github.com/samber/do/v2"
)

type IStaticHandler interface {
	TestClient(c *echo.Context) error
}

type StaticHandler struct{}

func NewStaticHandler(_ do.Injector) (IStaticHandler, error) {
	return &StaticHandler{}, nil
}

func (h *StaticHandler) TestClient(c *echo.Context) error {
	return c.File("static/testclient.html")
}
