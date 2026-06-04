package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/struckchure/rt-comms/models"
	"github.com/struckchure/rt-comms/internal"
	"github.com/struckchure/rt-comms/services"
)

type fakeChat struct{}

func (fakeChat) SendText(context.Context, string, string, string) (*models.Message, error) {
	return nil, nil
}
func (fakeChat) HandleInbound(context.Context, string, internal.Envelope) error { return nil }
func (fakeChat) ListThread(context.Context, string, string, int, *uuid.UUID) ([]models.Message, error) {
	return []models.Message{}, nil
}
func (fakeChat) RecordCallMessage(context.Context, string, string, models.MessageType, models.CallMetadata) (*models.Message, error) {
	return nil, nil
}
func (fakeChat) DeliverMessage(context.Context, *models.Message) error { return nil }

var _ services.IChatService = fakeChat{}

func TestHistoryHandlerListMessages(t *testing.T) {
	e := echo.New()
	h := &HistoryHandler{chat: fakeChat{}}
	e.GET("/users/:user/direct/:peer/messages", h.ListMessages)

	req := httptest.NewRequest(http.MethodGet, "/users/alice/direct/bob/messages", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var out []models.Message
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty array, got %d", len(out))
	}
}
