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
		UserPubkey string `json:"user_pubkey" binding:"required"`
	}
	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	convo, err := c.convoLogic.CreateConversation(req.UserPubkey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, convo)
}

// GetConversations handles GET /conversations
func (c *ConversationController) GetConversations(ctx *gin.Context) {
    type Query struct {
        UserPubkey string `form:"user_pubkey" binding:"required"`
    }
    var qry Query
    if err := ctx.ShouldBindQuery(&qry); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    conversations, err := c.convoLogic.GetConversations(qry.UserPubkey)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, conversations)
}