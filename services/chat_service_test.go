package services

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/struckchure/rt-comms/internal"
	"github.com/struckchure/rt-comms/models"
	"github.com/struckchure/rt-comms/repository"
	"github.com/struckchure/rt-comms/ws"
)

type fakeRepo struct {
	mu      sync.Mutex
	inserts int
}

func (f *fakeRepo) InsertMessage(ctx context.Context, msg *models.Message) error {
	f.mu.Lock()
	f.inserts++
	f.mu.Unlock()
	msg.ID = uuid.New().String()
	return nil
}

func (f *fakeRepo) ListThread(ctx context.Context, user, peer string, limit int, before *uuid.UUID) ([]models.Message, error) {
	return nil, nil
}

type recordingHub struct {
	mu       sync.Mutex
	delivers int
}

func (h *recordingHub) Register(string, *ws.Client) {}
func (h *recordingHub) Unregister(string)          {}
func (h *recordingHub) IsOnline(string) bool       { return true }
func (h *recordingHub) Shutdown(context.Context) error {
	return nil
}

func (h *recordingHub) DeliverTo(ctx context.Context, userID string, payload []byte) error {
	h.mu.Lock()
	h.delivers++
	h.mu.Unlock()
	return nil
}

var _ internal.IHub = (*recordingHub)(nil)
var _ repository.IMessageRepository = (*fakeRepo)(nil)

func TestChatServicePersistBeforeDeliver(t *testing.T) {
	repo := &fakeRepo{}
	h := &recordingHub{}
	svc := &ChatService{repo: repo, h: h}
	_, err := svc.SendText(context.Background(), "alice", "bob", "hi")
	if err != nil {
		t.Fatal(err)
	}
	if repo.inserts != 1 {
		t.Fatalf("expected 1 insert, got %d", repo.inserts)
	}
	if h.delivers < 1 {
		t.Fatalf("expected deliver after insert, delivers=%d", h.delivers)
	}
}
