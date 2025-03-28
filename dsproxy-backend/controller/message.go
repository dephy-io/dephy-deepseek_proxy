package controller

import (
	"fmt"
	"net/http"

	"dsproxy-backend/logic"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MessageController handles HTTP requests
type MessageController struct {
	messageLogic *logic.MessageLogic
	convoLogic   *logic.ConversationLogic
}

func NewMessageController(messageLogic *logic.MessageLogic, convoLogic *logic.ConversationLogic) *MessageController {
	return &MessageController{messageLogic: messageLogic, convoLogic: convoLogic}
}

// AddMessage handles POST /messages
func (c *MessageController) AddMessage(ctx *gin.Context) {
	type Request struct {
		ConvoID uuid.UUID `json:"conversation_id" binding:"required"`
		Content string    `json:"content" binding:"required"`
		Model   string    `json:"model" binding:"required"`
	}
	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Stream response to client using Server-Sent Events
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")

	msg, err := c.messageLogic.AddMessageAndCallChat(req.ConvoID, req.Model, req.Content, func(content string) {
		ctx.SSEvent("message", gin.H{
			"content": content,
		})
		ctx.Writer.Flush()
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.SSEvent("done", msg)
	ctx.Writer.Flush()
}

// GetMessages handles GET /messages
func (c *MessageController) GetMessages(ctx *gin.Context) {
	convoIDStr := ctx.Query("conversation_id")
	if convoIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "conversation_id is required"})
		return
	}

	// Parse UUID manually
	convoID, err := uuid.Parse(convoIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid conversation_id: %v", err)})
		return
	}

	userPubkey, err := extractUserPubkey(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conversation, err := c.convoLogic.GetConversationByID(convoID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("no conversation found: %v", err)})
		return
	}

	if conversation.UserPubkey != userPubkey {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("the conversation not matched to the user pubkey: %v", err)})
		return
	}

	messages, err := c.messageLogic.GetConversationMessages(convoID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, messages)
}
