package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/struckchure/rt-comms/internal"
	"github.com/struckchure/rt-comms/models"
	"github.com/struckchure/rt-comms/ws"
)

type offlineHub struct{}

func (offlineHub) Register(string, *ws.Client)              {}
func (offlineHub) Unregister(string)                        {}
func (offlineHub) IsOnline(string) bool                     { return false }
func (offlineHub) DeliverTo(context.Context, string, []byte) error {
	return nil
}
func (offlineHub) Shutdown(context.Context) error { return nil }

var _ internal.IHub = offlineHub{}

type recordingChat struct {
	recorded bool
}

func (c *recordingChat) SendText(context.Context, string, string, string) (*models.Message, error) {
	return nil, nil
}
func (c *recordingChat) HandleInbound(context.Context, string, internal.Envelope) error {
	return nil
}
func (c *recordingChat) ListThread(context.Context, string, string, int, *uuid.UUID) ([]models.Message, error) {
	return nil, nil
}
func (c *recordingChat) RecordCallMessage(ctx context.Context, sender, recipient string, msgType models.MessageType, meta models.CallMetadata) (*models.Message, error) {
	c.recorded = true
	return &models.Message{ID: "1", Sender: sender, Recipient: recipient, MessageType: msgType}, nil
}
func (c *recordingChat) DeliverMessage(context.Context, *models.Message) error { return nil }

var _ IChatService = (*recordingChat)(nil)

func TestCallOfferPeerOffline(t *testing.T) {
	store, _ := NewCallSessionStore(nil)
	svc := &CallService{h: offlineHub{}, chat: &recordingChat{}, store: store}
	err := svc.HandleSignaling(context.Background(), "alice", internal.Envelope{
		Type:  internal.TypeCallOffer,
		To:    "bob",
		Media: "voice",
	})
	if err != internal.ErrPeerOffline {
		t.Fatalf("expected peer offline, got %v", err)
	}
}

func TestCallHangupRecordsVoiceMessage(t *testing.T) {
	h, _ := internal.NewHub(nil)
	chat := &recordingChat{}
	store, _ := NewCallSessionStore(nil)
	svc := &CallService{h: h, chat: chat, store: store}
	sess := store.Create("alice", "bob", "voice")
	err := svc.HandleSignaling(context.Background(), "alice", internal.Envelope{
		Type:   internal.TypeCallHangup,
		CallID: sess.CallID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !chat.recorded {
		t.Fatal("expected call message recorded")
	}
}
