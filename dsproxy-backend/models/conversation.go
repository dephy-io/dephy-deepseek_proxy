package models

import (
	"time"

	"github.com/google/uuid"
)

// Conversation represents a conversation
type Conversation struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserPubkey  string    `gorm:"not null" json:"user_pubkey"`
	TotalTokens uint64    `gorm:"default:0" json:"total_tokens"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
