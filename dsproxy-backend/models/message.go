package models

import (
	"time"

	"github.com/google/uuid"
)

// Message represents a message in a conversation
type Message struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ConversationID uuid.UUID `gorm:"type:uuid;not null" json:"conversation_id"`
	Role           string    `gorm:"not null" json:"role"` // "user" for ask, "assistant" for answer
	Content        string    `gorm:"not null" json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}
