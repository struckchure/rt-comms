package internal

import (
	"encoding/json"
	"fmt"
)

const (
	TypeMessage     = "message"
	TypeError       = "error"
	TypeAck         = "ack"
	TypeCallOffer   = "call_offer"
	TypeCallAnswer  = "call_answer"
	TypeCallReject  = "call_reject"
	TypeCallHangup  = "call_hangup"
	TypeCallICE       = "call_ice"
	TypeCallStarted   = "call_started"
)

type Envelope struct {
	Type        string          `json:"type"`
	MessageType string          `json:"message_type,omitempty"`
	From        string          `json:"from,omitempty"`
	To          string          `json:"to,omitempty"`
	Text        string          `json:"text,omitempty"`
	CallID      string          `json:"call_id,omitempty"`
	Media       string          `json:"media,omitempty"`
	SDP         json.RawMessage `json:"sdp,omitempty"`
	Candidate   json.RawMessage `json:"candidate,omitempty"`
	Reason      string          `json:"reason,omitempty"`
	Code        string          `json:"code,omitempty"`
	Message     string          `json:"message,omitempty"`
	ID          string          `json:"id,omitempty"`
	CreatedAt   string          `json:"created_at,omitempty"`
}

func ParseInbound(data []byte) (Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return Envelope{}, fmt.Errorf("%w: %v", errInvalid, err)
	}
	if env.Type == "" {
		return Envelope{}, errInvalid
	}
	return env, nil
}

var errInvalid = fmt.Errorf("invalid message")

func (e Envelope) Marshal() ([]byte, error) {
	return json.Marshal(e)
}
