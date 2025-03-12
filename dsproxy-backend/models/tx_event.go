package models

import "time"

type TransactionEvent struct {
	ID        string    `gorm:"primaryKey" json:"id"` // Nostr event ID
	User      string    `gorm:"not null" json:"user"`
	Tokens    int64     `gorm:"not null" json:"tokens"`
	CreatedAt time.Time `json:"created_at"`
}
