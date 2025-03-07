package logic

import (
	"errors"
	"log"

	"dsproxy-backend/config"
	"dsproxy-backend/dao"
	"dsproxy-backend/models"
	"dsproxy-backend/pkg"

	"github.com/google/uuid"
)

// MessageLogic handles message-related business logic
type MessageLogic struct {
	userDAO    *dao.UserDAO
	convoDAO   *dao.ConversationDAO
	messageDAO *dao.MessageDAO
	chatClient *pkg.ChatClient
}

func NewMessageLogic(
	userDAO *dao.UserDAO,
	convoDAO *dao.ConversationDAO,
	messageDAO *dao.MessageDAO,
	chatClient *pkg.ChatClient,
) *MessageLogic {
	return &MessageLogic{
		userDAO:    userDAO,
		convoDAO:   convoDAO,
		messageDAO: messageDAO,
		chatClient: chatClient,
	}
}

// AddMessageAndCallChat adds a message and calls the chat API, only saves to DB on success
func (l *MessageLogic) AddMessageAndCallChat(conversationID uuid.UUID, model string, content string, streamHandler func(string)) (*models.Message, error) {
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

	// Check available context and user tokens
	remainingContextTokens := uint64(config.GlobalConfig.Chat.MaxContextTokens) - conversation.TotalTokens
	if remainingContextTokens < 1 {
		return nil, errors.New("conversation context limit exceeded")
	}

	// Check user's token balance
	if user.Tokens < 1 {
		return nil, errors.New("insufficient tokens")
	}

	// Calculate max_tokens: min(remaining context, user tokens)
	maxTokens := uint32(remainingContextTokens)
	if user.Tokens < remainingContextTokens {
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
		Content: content,
	})

	streamTrue := true
	streamOptions := pkg.StreamOptions{
		IncludeUsage: true,
	}
	req := pkg.ChatCompletionRequest{
		Model:         model,
		Messages:      chatMessages,
		MaxTokens:     maxTokens,
		Stream:        &streamTrue,
		StreamOptions: &streamOptions,
	}

	// Buffer to collect full response and track usage
	var fullResponse string
	var finalUsage *pkg.Usage

	// Call chat API with streaming
	err = l.chatClient.CreateChatCompletionStream(req, func(resp *pkg.StreamChatCompletionResponse) error {
		for _, choice := range resp.Choices {
			if choice.Delta.Content != "" {
				fullResponse += choice.Delta.Content
				streamHandler(choice.Delta.Content)
			}
		}
		if resp.Usage != nil {
			finalUsage = resp.Usage // 捕获最后一个块的 usage
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Validate usage data
	if finalUsage.TotalTokens == 0 {
		return nil, errors.New("invalid usage data from chat API")
	}

	// Only save messages to DB if API call succeeds
	// Save user's ask
	_, err = l.messageDAO.CreateMessage(conversationID, "user", content)
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

	// Update conversation's TotalTokens
	newTotalTokens := conversation.TotalTokens + uint64(finalUsage.TotalTokens)
	if err := l.convoDAO.UpdateTotalTokens(conversationID, newTotalTokens); err != nil {
		log.Printf("Failed to update conversation total tokens: %v", err)
	}

	return answer, nil
}

// GetConversationMessages retrieves all messages in a conversation
func (l *MessageLogic) GetConversationMessages(conversationID uuid.UUID) ([]models.Message, error) {
	return l.messageDAO.GetMessagesByConversationID(conversationID)
}
