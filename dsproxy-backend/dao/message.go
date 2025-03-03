package dao

import (
    "github.com/google/uuid"
    "gorm.io/gorm"
    "dsproxy-backend/models"
)

// MessageDAO handles message-related database operations
type MessageDAO struct {
    db *gorm.DB
}

func NewMessageDAO(db *gorm.DB) *MessageDAO {
    return &MessageDAO{db: db}
}

// CreateMessage adds a message to a conversation
func (d *MessageDAO) CreateMessage(conversationID uuid.UUID, role, content string) (*models.Message, error) {
    msg := &models.Message{
        ConversationID: conversationID,
        Role:           role,
        Content:        content,
    }
    if err := d.db.Create(msg).Error; err != nil {
        return nil, err
    }
    return msg, nil
}

// GetMessagesByConversationID retrieves all messages in a conversation
func (d *MessageDAO) GetMessagesByConversationID(conversationID uuid.UUID) ([]models.Message, error) {
    var messages []models.Message
    if err := d.db.Where("conversation_id = ?", conversationID).Order("created_at ASC").Find(&messages).Error; err != nil {
        return nil, err
    }
    return messages, nil
}