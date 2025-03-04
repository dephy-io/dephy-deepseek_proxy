package controller

import (
	"net/http"

	"dsproxy-backend/logic"

	"github.com/gin-gonic/gin"
)

// ConversationController handles HTTP requests
type ConversationController struct {
	convoLogic *logic.ConversationLogic
}

func NewConversationController(logic *logic.ConversationLogic) *ConversationController {
	return &ConversationController{convoLogic: logic}
}

// CreateConversation handles POST /conversations
func (c *ConversationController) CreateConversation(ctx *gin.Context) {
	type Request struct {
		PublicKey string `json:"public_key" binding:"required"`
	}
	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	convo, err := c.convoLogic.CreateConversation(req.PublicKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, convo)
}
