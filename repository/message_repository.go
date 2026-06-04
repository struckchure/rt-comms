package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
	"github.com/struckchure/rt-comms/models"
)

type IMessageRepository interface {
	InsertMessage(ctx context.Context, msg *models.Message) error
	ListThread(ctx context.Context, user, peer string, limit int, before *uuid.UUID) ([]models.Message, error)
}

type MessageRepository struct {
	pool *pgxpool.Pool
}

func nullJSON(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	return raw
}

func NewMessageRepository(i do.Injector) (IMessageRepository, error) {
	pool := do.MustInvoke[*pgxpool.Pool](i)
	return &MessageRepository{pool: pool}, nil
}

func (r *MessageRepository) InsertMessage(ctx context.Context, msg *models.Message) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	var createdAt any
	if !msg.CreatedAt.IsZero() {
		createdAt = msg.CreatedAt
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO messages (id, sender, recipient, message_type, content, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, COALESCE($7, NOW()))
	`, msg.ID, msg.Sender, msg.Recipient, msg.MessageType, msg.Content, nullJSON(msg.Metadata), createdAt)
	return err
}

func (r *MessageRepository) ListThread(ctx context.Context, user, peer string, limit int, before *uuid.UUID) ([]models.Message, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id::text, sender, recipient, message_type, content, metadata, created_at
		FROM messages
		WHERE (sender = $1 AND recipient = $2) OR (sender = $2 AND recipient = $1)
	`
	args := []any{user, peer}
	argN := 3
	if before != nil {
		query += fmt.Sprintf(" AND created_at < (SELECT created_at FROM messages WHERE id = $%d)", argN)
		args = append(args, before.String())
		argN++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argN)
	args = append(args, limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Message
	for rows.Next() {
		var m models.Message
		var meta []byte
		if err := rows.Scan(&m.ID, &m.Sender, &m.Recipient, &m.MessageType, &m.Content, &meta, &m.CreatedAt); err != nil {
			return nil, err
		}
		if len(meta) > 0 {
			m.Metadata = meta
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
