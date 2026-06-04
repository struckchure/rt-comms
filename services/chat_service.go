package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"github.com/struckchure/rt-comms/internal"
	"github.com/struckchure/rt-comms/models"
	"github.com/struckchure/rt-comms/repository"
)

type IChatService interface {
	SendText(ctx context.Context, sender, to, text string) (*models.Message, error)
	HandleInbound(ctx context.Context, sender string, env internal.Envelope) error
	ListThread(ctx context.Context, user, peer string, limit int, before *uuid.UUID) ([]models.Message, error)
	RecordCallMessage(ctx context.Context, sender, recipient string, msgType models.MessageType, meta models.CallMetadata) (*models.Message, error)
	DeliverMessage(ctx context.Context, msg *models.Message) error
}

type ChatService struct {
	repo repository.IMessageRepository
	h internal.IHub
}

func NewChatService(i do.Injector) (IChatService, error) {
	repo := do.MustInvoke[repository.IMessageRepository](i)
	h := do.MustInvoke[internal.IHub](i)
	return &ChatService{repo: repo, h: h}, nil
}

func (s *ChatService) SendText(ctx context.Context, sender, to, text string) (*models.Message, error) {
	if sender == to {
		return nil, internal.ErrSelfMessage
	}
	msg := &models.Message{
		Sender:      sender,
		Recipient:   to,
		MessageType: models.MessageTypeText,
		Content:     text,
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.repo.InsertMessage(ctx, msg); err != nil {
		return nil, err
	}
	if err := s.deliverAndAck(ctx, sender, msg); err != nil {
		return msg, err
	}
	return msg, nil
}

func (s *ChatService) HandleInbound(ctx context.Context, sender string, env internal.Envelope) error {
	if env.Type != internal.TypeMessage {
		return internal.ErrInvalidInput
	}
	if env.MessageType != "" && env.MessageType != string(models.MessageTypeText) {
		return internal.ErrInvalidInput
	}
	_, err := s.SendText(ctx, sender, env.To, env.Text)
	return err
}

func (s *ChatService) ListThread(ctx context.Context, user, peer string, limit int, before *uuid.UUID) ([]models.Message, error) {
	return s.repo.ListThread(ctx, user, peer, limit, before)
}

func (s *ChatService) RecordCallMessage(ctx context.Context, sender, recipient string, msgType models.MessageType, meta models.CallMetadata) (*models.Message, error) {
	raw, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	msg := &models.Message{
		Sender:      sender,
		Recipient:   recipient,
		MessageType: msgType,
		Metadata:    raw,
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.repo.InsertMessage(ctx, msg); err != nil {
		return nil, err
	}
	_ = s.deliverAndAck(ctx, sender, msg)
	return msg, nil
}

func (s *ChatService) DeliverMessage(ctx context.Context, msg *models.Message) error {
	return s.deliverAndAck(ctx, msg.Sender, msg)
}

func (s *ChatService) deliverAndAck(ctx context.Context, sender string, msg *models.Message) error {
	payload, err := messagePayload(msg)
	if err != nil {
		return err
	}
	_ = s.h.DeliverTo(ctx, msg.Recipient, payload)
	ack, _ := json.Marshal(internal.Envelope{
		Type:      internal.TypeAck,
		ID:        msg.ID,
		CreatedAt: msg.CreatedAt.Format(time.RFC3339Nano),
	})
	_ = s.h.DeliverTo(ctx, sender, ack)
	return nil
}

func messagePayload(msg *models.Message) ([]byte, error) {
	body := map[string]any{
		"type":         internal.TypeMessage,
		"id":           msg.ID,
		"sender":       msg.Sender,
		"recipient":    msg.Recipient,
		"message_type": msg.MessageType,
		"content":      msg.Content,
		"metadata":     msg.Metadata,
		"created_at":   msg.CreatedAt.Format(time.RFC3339Nano),
	}
	return json.Marshal(body)
}
