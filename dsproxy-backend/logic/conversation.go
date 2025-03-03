package logic

import (
    "context"
    "time"

    "github.com/google/uuid"
    "github.com/nbd-wtf/go-nostr"
    "your_project/dao"
    "your_project/models"
)

// ConversationLogic handles conversation-related business logic
type ConversationLogic struct {
    userDAO        *dao.UserDAO
    convoDAO       *dao.ConversationDAO
    messageDAO     *dao.MessageDAO
    nostrClient    *NostrClient // Assume NostrClient is defined elsewhere
}

func NewConversationLogic(userDAO *dao.UserDAO, convoDAO *dao.ConversationDAO, messageDAO *dao.MessageDAO, nostrClient *NostrClient) *ConversationLogic {
    return &ConversationLogic{
        userDAO:        userDAO,
        convoDAO:       convoDAO,
        messageDAO:     messageDAO,
        nostrClient:    nostrClient,
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

    return l.convoDAO.CreateConversation(user.ID)
}

// AddMessage adds a message to a conversation
func (l *ConversationLogic) AddMessage(conversationID uuid.UUID, role, content string) (*models.Message, error) {
    return l.messageDAO.CreateMessage(conversationID, role, content)
}

// GetConversationMessages retrieves all messages in a conversation
func (l *ConversationLogic) GetConversationMessages(conversationID uuid.UUID) ([]models.Message, error) {
    return l.messageDAO.GetMessagesByConversationID(conversationID)
}

// StartNostrListener starts listening to Nostr Transaction events
func (l *ConversationLogic) StartNostrListener(ctx context.Context) error {
    return l.nostrClient.SubscribeTransactions(ctx, func(event nostr.Event) {
        var msg models.DephyDsProxyMessage
        if err := json.Unmarshal([]byte(event.Content), &msg); err != nil {
            log.Printf("Failed to parse event content: %v", err)
            return
        }

        if msg.Transaction != nil {
            // Update user tokens (1 Lamport = 1 Token)
            err := l.userDAO.UpdateUserTokens(msg.Transaction.User, int64(msg.Transaction.Lamports), 0)
            if err != nil {
                log.Printf("Failed to update user tokens: %v", err)
            } else {
                log.Printf("Updated tokens for user %s: +%d", msg.Transaction.User, msg.Transaction.Lamports)
            }
        }
    })
}
