package controller

import (
	"net/http"

	"dsproxy-backend/logic"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MessageController handles HTTP requests
type MessageController struct {
	messageLogic *logic.MessageLogic
}

func NewMessageController(logic *logic.MessageLogic) *MessageController {
	return &MessageController{messageLogic: logic}
}

// AddMessage handles POST /conversations/:id/messages
func (c *MessageController) AddMessage(ctx *gin.Context) {
	type Request struct {
		Ask   string `json:"ask" binding:"required"`
		Model string `json:"model" binding:"required"`
	}
	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	convoID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	// Stream response to client using Server-Sent Events
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")

	msg, err := c.messageLogic.AddMessageAndCallChat(convoID, req.Model, req.Ask, func(content string) {
		ctx.SSEvent("message", content)
		ctx.Writer.Flush()
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.SSEvent("done", msg)
	ctx.Writer.Flush()
}

// GetMessages handles GET /conversations/:id/messages
func (c *MessageController) GetMessages(ctx *gin.Context) {
	convoID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	messages, err := c.messageLogic.GetConversationMessages(convoID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, messages)
}
