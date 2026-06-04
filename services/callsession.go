package services

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type CallSession struct {
	CallID    string
	Caller    string
	Callee    string
	Media     string
	StartedAt time.Time
}

type CallSessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*CallSession
	byUser   map[string]string
}

func NewCallSessionStore(_ do.Injector) (*CallSessionStore, error) {
	return &CallSessionStore{
		sessions: make(map[string]*CallSession),
		byUser:   make(map[string]string),
	}, nil
}

func (s *CallSessionStore) Create(caller, callee, media string) *CallSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := uuid.New().String()
	sess := &CallSession{
		CallID:    id,
		Caller:    caller,
		Callee:    callee,
		Media:     media,
		StartedAt: time.Now(),
	}
	s.sessions[id] = sess
	s.byUser[caller] = id
	s.byUser[callee] = id
	return sess
}

func (s *CallSessionStore) Get(callID string) (*CallSession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[callID]
	return sess, ok
}

func (s *CallSessionStore) GetByUser(userID string) (*CallSession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.byUser[userID]
	if !ok {
		return nil, false
	}
	return s.sessions[id], true
}

func (s *CallSessionStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions = make(map[string]*CallSession)
	s.byUser = make(map[string]string)
}

func (s *CallSessionStore) Delete(callID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[callID]
	if !ok {
		return
	}
	delete(s.sessions, callID)
	delete(s.byUser, sess.Caller)
	delete(s.byUser, sess.Callee)
}
