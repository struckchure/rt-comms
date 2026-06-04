package handlers

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v5"
	"github.com/samber/do/v2"
	"github.com/struckchure/rt-comms/internal"
	"github.com/struckchure/rt-comms/services"
)

type IWebSocketHandler interface {
	HandleWS(c *echo.Context) error
}

type WebSocketHandler struct {
	conn   services.IConnectionService
	upgrader websocket.Upgrader
}

func NewWebSocketHandler(i do.Injector) (IWebSocketHandler, error) {
	conn := do.MustInvoke[services.IConnectionService](i)
	return &WebSocketHandler{
		conn: conn,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}, nil
}

func (h *WebSocketHandler) HandleWS(c *echo.Context) error {
	user := c.QueryParam("user")
	if user == "" {
		return writeHTTPError(c, internal.ErrInvalidUser)
	}
	wsConn, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	h.conn.HandleConnection(c.Request().Context(), user, wsConn)
	return nil
}
