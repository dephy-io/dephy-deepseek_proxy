package controller

import (
	"errors"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func extractUserPubkey(c *gin.Context) (string, error) {
	userClaims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return "", errors.New("user not found in context")
	}

	claims, ok := userClaims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user claims"})
		return "", errors.New("invalid user claims")
	}

	userPubkey, ok := claims["user_pubkey"].(string)
	if !ok || userPubkey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user_pubkey not found in token"})
		return "", errors.New("user_pubkey not found in token")
	}

	return userPubkey, nil
}
