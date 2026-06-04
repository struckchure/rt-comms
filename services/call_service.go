package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/samber/do/v2"
	"github.com/struckchure/rt-comms/internal"
	"github.com/struckchure/rt-comms/models"
)

type ICallService interface {
	HandleSignaling(ctx context.Context, sender string, env internal.Envelope) error
	EndAll()
	OnUserDisconnect(userID string)
}

type CallService struct {
	h     internal.IHub
	chat  IChatService
	store *CallSessionStore
}

func NewCallService(i do.Injector) (ICallService, error) {
	h := do.MustInvoke[internal.IHub](i)
	chat := do.MustInvoke[IChatService](i)
	store := do.MustInvoke[*CallSessionStore](i)
	return &CallService{h: h, chat: chat, store: store}, nil
}

func (s *CallService) HandleSignaling(ctx context.Context, sender string, env internal.Envelope) error {
	switch env.Type {
	case internal.TypeCallOffer:
		return s.handleOffer(ctx, sender, env)
	case internal.TypeCallAnswer, internal.TypeCallReject, internal.TypeCallHangup, internal.TypeCallICE:
		return s.forward(ctx, sender, env)
	default:
		return internal.ErrInvalidInput
	}
}

func (s *CallService) handleOffer(ctx context.Context, sender string, env internal.Envelope) error {
	if env.To == "" {
		return internal.ErrInvalidInput
	}
	if !s.h.IsOnline(env.To) {
		return internal.ErrPeerOffline
	}
	sess := s.store.Create(sender, env.To, env.Media)
	env.CallID = sess.CallID
	env.From = sender

	toCallee, _ := json.Marshal(env)
	if err := s.h.DeliverTo(ctx, env.To, toCallee); err != nil {
		return err
	}

	started, _ := json.Marshal(internal.Envelope{
		Type:   internal.TypeCallStarted,
		CallID: sess.CallID,
		To:     env.To,
		Media:  env.Media,
	})
	return s.h.DeliverTo(ctx, sender, started)
}

func (s *CallService) forward(ctx context.Context, sender string, env internal.Envelope) error {
	callID := env.CallID
	if callID == "" {
		return internal.ErrInvalidInput
	}
	sess, ok := s.store.Get(callID)
	if !ok {
		return internal.ErrNotFound
	}
	peer := sess.Callee
	if sender == sess.Callee {
		peer = sess.Caller
	}
	payload, _ := json.Marshal(env)
	_ = s.h.DeliverTo(ctx, peer, payload)

	switch env.Type {
	case internal.TypeCallReject:
		s.finalize(ctx, sess, sender, peer, "rejected", 0)
	case internal.TypeCallHangup:
		dur := int(time.Since(sess.StartedAt).Seconds())
		s.finalize(ctx, sess, sender, peer, "completed", dur)
	}
	return nil
}

func (s *CallService) finalize(ctx context.Context, sess *CallSession, actor, peer, status string, dur int) {
	s.store.Delete(sess.CallID)
	msgType := models.MessageTypeVoice
	if sess.Media == "video" {
		msgType = models.MessageTypeVideo
	}
	meta := models.CallMetadata{
		CallID:      sess.CallID,
		Status:      status,
		DurationSec: dur,
		Media:       sess.Media,
	}
	msg, err := s.chat.RecordCallMessage(ctx, actor, peer, msgType, meta)
	if err == nil && msg != nil {
		_ = s.chat.DeliverMessage(ctx, msg)
	}
}

func (s *CallService) EndAll() {
	s.store.Clear()
}

func (s *CallService) OnUserDisconnect(userID string) {
	if sess, ok := s.store.GetByUser(userID); ok {
		peer := sess.Callee
		if userID == sess.Callee {
			peer = sess.Caller
		}
		ctx := context.Background()
		s.finalize(ctx, sess, userID, peer, "missed", 0)
	}
}
