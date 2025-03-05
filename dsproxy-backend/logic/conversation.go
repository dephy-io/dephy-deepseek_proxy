package logic

import (
	"dsproxy-backend/dao"
	"dsproxy-backend/models"

	"gorm.io/gorm"
)

// ConversationLogic handles conversation-related business logic
type ConversationLogic struct {
	userDAO  *dao.UserDAO
	convoDAO *dao.ConversationDAO
}

func NewConversationLogic(
	userDAO *dao.UserDAO,
	convoDAO *dao.ConversationDAO,
) *ConversationLogic {
	return &ConversationLogic{
		userDAO:  userDAO,
		convoDAO: convoDAO,
	}
}

// CreateConversation creates a new conversation for a user
func (l *ConversationLogic) CreateConversation(publicKey string) (*models.Conversation, error) {
	user, err := l.userDAO.GetUserByPublicKey(publicKey)
	if err != nil {
		// If user doesn't exist, create one
		if err == gorm.ErrRecordNotFound {
			user, err = l.userDAO.CreateUser(publicKey)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return l.convoDAO.CreateConversation(user.PublicKey)
}

// GetConversations retrieves all conversations for a user by public key
func (l *ConversationLogic) GetConversations(publicKey string) ([]models.Conversation, error) {
	// Optional: Verify user exists
	_, err := l.userDAO.GetUserByPublicKey(publicKey)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return []models.Conversation{}, nil // Return empty list if user doesn't exist
		}
		return nil, err
	}

	return l.convoDAO.GetConversationsByUserPubkey(publicKey)
}
