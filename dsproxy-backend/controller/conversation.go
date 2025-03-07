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
	userPubkey, err := extractUserPubkey(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	convo, err := c.convoLogic.CreateConversation(userPubkey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, convo)
}

// GetConversations handles GET /conversations
func (c *ConversationController) GetConversations(ctx *gin.Context) {
	userPubkey, err := extractUserPubkey(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

    conversations, err := c.convoLogic.GetConversations(userPubkey)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, conversations)
}