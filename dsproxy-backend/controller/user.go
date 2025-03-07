package controller

import (
	"dsproxy-backend/logic"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UserController handles HTTP requests
type UserController struct {
	userLogic *logic.UserLogic
}

func NewUserController(logic *logic.UserLogic) *UserController {
	return &UserController{userLogic: logic}
}

// GetUser handles GET /user
func (c *UserController) GetUser(ctx *gin.Context) {
	userPubkey, err := extractUserPubkey(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	users, err := c.userLogic.GetUser(userPubkey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, users)
}

func (c *UserController) Login(ctx *gin.Context) {
	type Request struct {
		UserPubkey string `json:"user_pubkey" binding:"required"`
		Message    string `json:"message" binding:"required"`
		Signature  string `json:"signature" binding:"required"`
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, expireAt, err := c.userLogic.Login(req.UserPubkey, req.Message, req.Signature)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
		"expire_at": expireAt,
	})
}
