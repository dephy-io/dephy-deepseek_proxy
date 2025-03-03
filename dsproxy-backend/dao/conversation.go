package dao

import (
	"dsproxy-backend/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConversationDAO handles conversation-related database operations
type ConversationDAO struct {
	db *gorm.DB
}

func NewConversationDAO(db *gorm.DB) *ConversationDAO {
	return &ConversationDAO{db: db}
}

// CreateConversation creates a new conversation
func (d *ConversationDAO) CreateConversation(userID uint64) (*models.Conversation, error) {
	convo := &models.Conversation{
		ID:     uuid.New(),
		UserID: userID,
	}
	if err := d.db.Create(convo).Error; err != nil {
		return nil, err
	}
	return convo, nil
}
