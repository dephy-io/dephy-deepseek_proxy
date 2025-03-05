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
	userPubkey := ctx.Query("user_pubkey")
	if userPubkey == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user_pubkey is required"})
		return
	}

	users, err := c.userLogic.GetUser(userPubkey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, users)
}
