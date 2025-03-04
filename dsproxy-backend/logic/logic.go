package logic

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"dsproxy-backend/dao"
	"dsproxy-backend/models"
	"dsproxy-backend/pkg"

	"github.com/google/uuid"
	"github.com/nbd-wtf/go-nostr"
	"gorm.io/gorm"
)

const (
	MAX_CONTEXT_TOKENS = 4096 // Maximum allowed context tokens, adjust as needed
)

// ConversationLogic handles conversation-related business logic
type ConversationLogic struct {
	userDAO     *dao.UserDAO
	convoDAO    *dao.ConversationDAO
	messageDAO  *dao.MessageDAO
	txEventDAO  *dao.TransactionEventDAO
	chatClient  *pkg.ChatClient
	nostrClient *pkg.NostrClient
}

func NewConversationLogic(
	userDAO *dao.UserDAO,
	convoDAO *dao.ConversationDAO,
	messageDAO *dao.MessageDAO,
	txEventDAO *dao.TransactionEventDAO,
	chatClient *pkg.ChatClient,
	nostrClient *pkg.NostrClient,
) *ConversationLogic {
	return &ConversationLogic{
		userDAO:     userDAO,
		convoDAO:    convoDAO,
		messageDAO:  messageDAO,
		txEventDAO:  txEventDAO,
		chatClient:  chatClient,
		nostrClient: nostrClient,
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

// AddMessageAndCallChat adds a message and calls the chat API, only saves to DB on success
func (l *ConversationLogic) AddMessageAndCallChat(conversationID uuid.UUID, ask string, streamHandler func(string)) (*models.Message, error) {
    // Fetch all existing messages in the conversation
    messages, err := l.messageDAO.GetMessagesByConversationID(conversationID)
    if err != nil {
        return nil, err
    }

    // Get the user associated with the conversation
    conversation, err := l.convoDAO.GetConversationByID(conversationID)
    if err != nil {
        return nil, err
    }
    user, err := l.userDAO.GetUserByPublicKey(conversation.UserPubkey)
    if err != nil {
        return nil, err
    }

    // Check user's token balance
    if user.Tokens < 1 {
        return nil, errors.New("insufficient tokens")
    }

    // Calculate max_tokens: min(user.Tokens, MAX_CONTEXT_TOKENS)
    maxTokens := uint32(MAX_CONTEXT_TOKENS)
    if user.Tokens < uint64(MAX_CONTEXT_TOKENS) {
        maxTokens = uint32(user.Tokens)
    }

    // Prepare chat request with existing messages plus the new ask
    var chatMessages []pkg.RequestMessage
    for _, msg := range messages {
        chatMessages = append(chatMessages, pkg.RequestMessage{
            Role:    msg.Role,
            Content: msg.Content,
        })
    }
    chatMessages = append(chatMessages, pkg.RequestMessage{
        Role:    "user",
        Content: ask,
    })

    streamTrue := true
    req := pkg.ChatCompletionRequest{
        Model:     "some-model",
        Messages:  chatMessages,
        MaxTokens: maxTokens,
        Stream:    &streamTrue,
    }

    // Buffer to collect full response
    var fullResponse string

    // Call chat API with streaming
    err = l.chatClient.CreateChatCompletionStream(req, func(resp pkg.ChatCompletionResponse) error {
        for _, choice := range resp.Choices {
            if choice.Message.Content != "" {
                fullResponse += choice.Message.Content
                streamHandler(choice.Message.Content)
            }
        }
        return nil
    })
    if err != nil {
        return nil, err
    }

    // Only save messages to DB if API call succeeds
    // Save user's ask
    _, err = l.messageDAO.CreateMessage(conversationID, "user", ask)
    if err != nil {
        return nil, err
    }

    // Save assistant's response
    answer, err := l.messageDAO.CreateMessage(conversationID, "assistant", fullResponse)
    if err != nil {
        return nil, err
    }

    // Update user's consumed tokens (approximation based on response length)
    consumedTokens := uint64(len(fullResponse))
    if err := l.userDAO.UpdateUserTokens(user.PublicKey, -int64(consumedTokens), int64(consumedTokens)); err != nil {
        log.Printf("Failed to update user tokens: %v", err)
    }

    return answer, nil
}

// GetConversationMessages retrieves all messages in a conversation
func (l *ConversationLogic) GetConversationMessages(conversationID uuid.UUID) ([]models.Message, error) {
	return l.messageDAO.GetMessagesByConversationID(conversationID)
}

// SyncTransactionEvents syncs historical Transaction events at startup
func (l *ConversationLogic) SyncTransactionEvents(ctx context.Context) error {
	// Fetch stored Transaction events from DB
	storedEvents, err := l.txEventDAO.GetAllTransactionEvents()
	if err != nil {
		return err
	}
	storedEventMap := make(map[string]bool)
	for _, evt := range storedEvents {
		storedEventMap[evt.ID] = true
	}

	// Subscribe to historical Transaction events from Relay
	since := nostr.Timestamp(0) // Fetch all events from the beginning
	filters := nostr.Filters{{
		Kinds: []int{1573},
		Since: &since,
		Tags: nostr.TagMap{
			"s": []string{"dephy-dsproxy-controller"},
			"p": []string{l.nostrClient.MachinePubkey()},
		},
	}}

	sub, err := l.nostrClient.Relay().Subscribe(ctx, filters)
	if err != nil {
		return err
	}
	defer sub.Unsub()

	for ev := range sub.Events {
		var msg pkg.DephyDsProxyMessage
		if err := json.Unmarshal([]byte(ev.Content), &msg); err != nil {
			log.Printf("Failed to parse event content: %v", err)
			continue
		}

		if msg.Transaction != nil {
			eventID := ev.ID
			if !storedEventMap[eventID] {
				// New event, save to DB and update user
				txEvent := &models.TransactionEvent{
					ID:        eventID,
					User:      msg.Transaction.User,
					Lamports:  msg.Transaction.Lamports,
					CreatedAt: time.Unix(int64(ev.CreatedAt), 0),
				}
				if err := l.txEventDAO.SaveTransactionEvent(txEvent); err != nil {
					log.Printf("Failed to save Transaction event: %v", err)
					continue
				}
				if err := l.userDAO.UpdateUserTokens(msg.Transaction.User, int64(msg.Transaction.Lamports), 0); err != nil {
					log.Printf("Failed to update user tokens: %v", err)
				}
			}
		}
	}

	return nil
}

// StartNostrListener starts listening to Nostr Transaction events
func (l *ConversationLogic) StartNostrListener(ctx context.Context) error {
	return l.nostrClient.SubscribeTransactions(ctx, func(event nostr.Event) {
		var msg pkg.DephyDsProxyMessage
		if err := json.Unmarshal([]byte(event.Content), &msg); err != nil {
			log.Printf("Failed to parse event content: %v", err)
			return
		}

		if msg.Transaction != nil {
			txEvent := &models.TransactionEvent{
				ID:        event.ID,
				User:      msg.Transaction.User,
				Lamports:  msg.Transaction.Lamports,
				CreatedAt: time.Unix(int64(event.CreatedAt), 0),
			}
			if err := l.txEventDAO.SaveTransactionEvent(txEvent); err != nil {
				log.Printf("Failed to save Transaction event: %v", err)
				return
			}

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
