package models

import (
	"encoding/json"
	"time"
)

type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeVoice MessageType = "voice"
	MessageTypeVideo MessageType = "video"
)

type Message struct {
	ID          string          `json:"id"`
	Sender      string          `json:"sender"`
	Recipient   string          `json:"recipient"`
	MessageType MessageType     `json:"message_type"`
	Content     string          `json:"content,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

type CallMetadata struct {
	CallID      string `json:"call_id"`
	Status      string `json:"status"`
	DurationSec int    `json:"duration_sec,omitempty"`
	Media       string `json:"media"`
}
