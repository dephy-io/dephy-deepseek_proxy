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
func (d *ConversationDAO) CreateConversation(userPubkey string) (*models.Conversation, error) {
	convo := &models.Conversation{
		ID:     uuid.New(),
		UserPubkey: userPubkey,
	}
	if err := d.db.Create(convo).Error; err != nil {
		return nil, err
	}
	return convo, nil
}


// GetConversationsByUserPubkey retrieves all conversations for a given user public key
func (d *ConversationDAO) GetConversationsByUserPubkey(userPubkey string) ([]models.Conversation, error) {
    var conversations []models.Conversation
    if err := d.db.Where("user_pubkey = ?", userPubkey).Order("created_at DESC").Find(&conversations).Error; err != nil {
        return nil, err
    }
    return conversations, nil
}

// GetConversationByID retrieves a conversation by its ID
func (d *ConversationDAO) GetConversationByID(conversationID uuid.UUID) (*models.Conversation, error) {
    var conversation models.Conversation
    if err := d.db.Where("id = ?", conversationID).First(&conversation).Error; err != nil {
        return nil, err
    }
    return &conversation, nil
}