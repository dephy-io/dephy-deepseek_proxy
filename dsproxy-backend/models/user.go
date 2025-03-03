package models

import (
	"time"
)

// User represents a user with token balances
type User struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	PublicKey      string    `gorm:"uniqueIndex;not null" json:"public_key"` // Nostr public key
	Tokens         uint64    `gorm:"default:0" json:"tokens"`                // Current token balance
	TokensConsumed uint64    `gorm:"default:0" json:"tokens_consumed"`       // Total tokens consumed
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
