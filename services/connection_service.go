package services

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/samber/do/v2"
	"github.com/struckchure/rt-comms/internal"
	"github.com/struckchure/rt-comms/ws"
)

type IConnectionService interface {
	HandleConnection(ctx context.Context, userID string, conn *websocket.Conn)
	DeliverTo(ctx context.Context, userID string, payload []byte) error
	Shutdown(ctx context.Context) error
}

type ConnectionService struct {
	h    internal.IHub
	chat IChatService
	call ICallService
	mu   sync.Mutex
}

func NewConnectionService(i do.Injector) (IConnectionService, error) {
	h := do.MustInvoke[internal.IHub](i)
	chat := do.MustInvoke[IChatService](i)
	call := do.MustInvoke[ICallService](i)
	return &ConnectionService{h: h, chat: chat, call: call}, nil
}

func (s *ConnectionService) HandleConnection(ctx context.Context, userID string, conn *websocket.Conn) {
	client := ws.NewClient(userID, conn)
	s.h.Register(userID, client)
	go client.WritePump()
	client.ReadPump(func(data []byte) {
		env, parseErr := internal.ParseInbound(data)
		if parseErr != nil {
			s.sendWSError(ctx, userID, internal.ErrInvalidInput)
			return
		}
		var handleErr error
		if isCallEnvelope(env.Type) {
			handleErr = s.call.HandleSignaling(ctx, userID, env)
		} else {
			handleErr = s.chat.HandleInbound(ctx, userID, env)
		}
		if handleErr != nil {
			s.sendWSError(ctx, userID, handleErr)
		}
	})
	s.h.Unregister(userID)
	s.call.OnUserDisconnect(userID)
}

func (s *ConnectionService) sendWSError(ctx context.Context, userID string, err error) {
	wsErr := internal.ToWS(err)
	b, _ := json.Marshal(wsErr)
	_ = s.h.DeliverTo(ctx, userID, b)
}

func (s *ConnectionService) DeliverTo(ctx context.Context, userID string, payload []byte) error {
	return s.h.DeliverTo(ctx, userID, payload)
}

func (s *ConnectionService) Shutdown(ctx context.Context) error {
	return s.h.Shutdown(ctx)
}

func isCallEnvelope(t string) bool {
	switch t {
	case internal.TypeCallOffer, internal.TypeCallAnswer, internal.TypeCallReject, internal.TypeCallHangup, internal.TypeCallICE, internal.TypeCallStarted:
		return true
	default:
		return false
	}
}
