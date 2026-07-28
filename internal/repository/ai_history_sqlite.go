package repository

import (
	"context"
	"database/sql"

	"wavearchive/internal/domain"
)

type AIHistorySQLite struct{ db *sql.DB }

func NewAIHistorySQLite(db *sql.DB) *AIHistorySQLite { return &AIHistorySQLite{db: db} }

func (r *AIHistorySQLite) List(ctx context.Context) ([]domain.AIConversation, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,title,context_type,context_id,provider,model,created_at,updated_at
		FROM ai_conversations ORDER BY updated_at DESC,id DESC`)
	if err != nil {
		return nil, err
	}
	conversations := []domain.AIConversation{}
	for rows.Next() {
		conversation, err := scanConversation(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		conversations = append(conversations, conversation)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	for index := range conversations {
		conversations[index].Messages, err = r.messages(ctx, conversations[index].ID)
		if err != nil {
			return nil, err
		}
	}
	return conversations, nil
}

func (r *AIHistorySQLite) Get(ctx context.Context, id int64) (domain.AIConversation, error) {
	conversation, err := scanConversation(r.db.QueryRowContext(ctx, `SELECT id,title,context_type,context_id,
		provider,model,created_at,updated_at FROM ai_conversations WHERE id=?`, id))
	if err != nil {
		return conversation, err
	}
	conversation.Messages, err = r.messages(ctx, id)
	return conversation, err
}

func (r *AIHistorySQLite) Create(ctx context.Context, conversation domain.AIConversation) (domain.AIConversation, error) {
	result, err := r.db.ExecContext(ctx, `INSERT INTO ai_conversations(title,context_type,context_id,provider,model)
		VALUES(?,?,?,?,?)`, conversation.Title, conversation.ContextType, nullableID(conversation.ContextID),
		conversation.Provider, conversation.Model)
	if err != nil {
		return conversation, err
	}
	conversation.ID, _ = result.LastInsertId()
	return r.Get(ctx, conversation.ID)
}

func (r *AIHistorySQLite) AddMessage(ctx context.Context, message domain.AIMessage) (domain.AIMessage, error) {
	result, err := r.db.ExecContext(ctx, "INSERT INTO ai_messages(conversation_id,role,content) VALUES(?,?,?)",
		message.ConversationID, message.Role, message.Content)
	if err != nil {
		return message, err
	}
	message.ID, _ = result.LastInsertId()
	_, _ = r.db.ExecContext(ctx, "UPDATE ai_conversations SET updated_at=CURRENT_TIMESTAMP WHERE id=?", message.ConversationID)
	err = r.db.QueryRowContext(ctx, "SELECT created_at FROM ai_messages WHERE id=?", message.ID).Scan(&message.CreatedAt)
	return message, err
}

func (r *AIHistorySQLite) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM ai_conversations WHERE id=?", id)
	return err
}

func (r *AIHistorySQLite) messages(ctx context.Context, conversationID int64) ([]domain.AIMessage, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id,conversation_id,role,content,created_at FROM ai_messages WHERE conversation_id=? ORDER BY id", conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	messages := []domain.AIMessage{}
	for rows.Next() {
		var message domain.AIMessage
		if err := rows.Scan(&message.ID, &message.ConversationID, &message.Role, &message.Content, &message.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, rows.Err()
}

func scanConversation(row scanner) (domain.AIConversation, error) {
	var conversation domain.AIConversation
	var contextID sql.NullInt64
	err := row.Scan(&conversation.ID, &conversation.Title, &conversation.ContextType, &contextID,
		&conversation.Provider, &conversation.Model, &conversation.CreatedAt, &conversation.UpdatedAt)
	if contextID.Valid {
		value := contextID.Int64
		conversation.ContextID = &value
	}
	return conversation, err
}
