package handlers

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/samber/do/v2"
	"github.com/struckchure/rt-comms/internal"
	"github.com/struckchure/rt-comms/models"
	"github.com/struckchure/rt-comms/services"
)

type IHistoryHandler interface {
	ListMessages(c *echo.Context) error
}

type HistoryHandler struct {
	chat services.IChatService
}

func NewHistoryHandler(i do.Injector) (IHistoryHandler, error) {
	chat := do.MustInvoke[services.IChatService](i)
	return &HistoryHandler{chat: chat}, nil
}

func (h *HistoryHandler) ListMessages(c *echo.Context) error {
	user := c.Param("user")
	peer := c.Param("peer")
	if user == "" || peer == "" {
		return writeHTTPError(c, internal.ErrInvalidInput)
	}
	limit := 50
	if l := c.QueryParam("limit"); l != "" {
		n, err := strconv.Atoi(l)
		if err != nil || n < 1 {
			return writeHTTPError(c, internal.ErrInvalidInput)
		}
		if n > 200 {
			n = 200
		}
		limit = n
	}
	var before *uuid.UUID
	if b := c.QueryParam("before"); b != "" {
		id, err := uuid.Parse(b)
		if err != nil {
			return writeHTTPError(c, internal.ErrInvalidInput)
		}
		before = &id
	}
	msgs, err := h.chat.ListThread(c.Request().Context(), user, peer, limit, before)
	if err != nil {
		return writeHTTPError(c, err)
	}
	if msgs == nil {
		msgs = []models.Message{}
	}
	return c.JSON(http.StatusOK, msgs)
}

func writeHTTPError(c *echo.Context, err error) error {
	status, body := internal.ToHTTP(err)
	return c.JSON(status, body)
}
