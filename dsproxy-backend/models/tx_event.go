package models

import "time"

type TransactionEvent struct {
	ID        string    `gorm:"primaryKey" json:"id"` // Nostr event ID
	User      string    `gorm:"not null" json:"user"`
	Lamports  uint64    `gorm:"not null" json:"lamports"`
	CreatedAt time.Time `json:"created_at"`
}
