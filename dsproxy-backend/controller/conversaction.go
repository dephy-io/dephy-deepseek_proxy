package controller

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "dsproxy-backend/logic"
)

// ConversationController handles HTTP requests
type ConversationController struct {
    logic *logic.ConversationLogic
}

func NewConversationController(logic *logic.ConversationLogic) *ConversationController {
    return &ConversationController{logic: logic}
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

    convo, err := c.logic.CreateConversation(req.PublicKey)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, convo)
}

// AddMessage handles POST /conversations/:id/messages
func (c *ConversationController) AddMessage(ctx *gin.Context) {
    type Request struct {
        Role    string `json:"role" binding:"required"`
        Content string `json:"content" binding:"required"`
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

    msg, err := c.logic.AddMessage(convoID, req.Role, req.Content)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, msg)
}

// GetMessages handles GET /conversations/:id/messages
func (c *ConversationController) GetMessages(ctx *gin.Context) {
    convoID, err := uuid.Parse(ctx.Param("id"))
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
        return
    }

    messages, err := c.logic.GetConversationMessages(convoID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, messages)
}

// StartNostrListener starts the Nostr event listener
func (c *ConversationController) StartNostrListener() {
    ctx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
    defer cancel()
    if err := c.logic.StartNostrListener(ctx); err != nil {
        log.Fatalf("Failed to start Nostr listener: %v", err)
    }
}