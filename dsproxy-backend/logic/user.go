package logic

import (
	"crypto/ed25519"
	"dsproxy-backend/config"
	"dsproxy-backend/dao"
	"dsproxy-backend/models"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mr-tron/base58"
)

// UserLogic handles user-related business logic
type UserLogic struct {
	userDAO *dao.UserDAO
}

func NewUserLogic(
	userDAO *dao.UserDAO,
) *UserLogic {
	return &UserLogic{
		userDAO: userDAO,
	}
}

// GetUser retrieves user info
func (l *UserLogic) GetUser(userPubkey string) (*models.User, error) {
	return l.userDAO.GetUserByPublicKey(userPubkey)
}

func (l *UserLogic) verifySignature(publicKey, message, signature string) (bool, error) {
	pubKeyBytes, err := base58.Decode(publicKey)
	if err != nil {
		return false, err
	}
	if len(pubKeyBytes) != 32 {
		return false, fmt.Errorf("ed25519: bad public key length: %d", len(pubKeyBytes))
	}

	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, err
	}
	msgBytes := []byte(message)

	return ed25519.Verify(pubKeyBytes, msgBytes, sigBytes), nil
}

func (l *UserLogic) generateJWT(userPubkey string) (string, time.Time, error) {
	expireAt := time.Now().Add(time.Duration(config.GlobalConfig.Auth.ExpHour) * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_pubkey": userPubkey,
		"exp":         time.Now().Add(time.Hour * time.Duration(config.GlobalConfig.Auth.ExpHour)).Unix(),
	})
	tokenString, err := token.SignedString([]byte(config.GlobalConfig.Auth.Secret))
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenString, expireAt, nil
}

func (l *UserLogic) Login(userPubkey, message, signature string) (*models.User, string, time.Time, error) {
	isValid, err := l.verifySignature(userPubkey, message, signature)
	if err != nil || !isValid {
		return nil, "", time.Time{}, errors.New("invalid signature")
	}

	user, err := l.userDAO.GetUserByPublicKey(userPubkey)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	if user == nil {
		user, err = l.userDAO.CreateUser(userPubkey)
		if err != nil {
			return nil, "", time.Time{}, err
		}
	}

	token, expireAt, err := l.generateJWT(userPubkey)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	user.UpdatedAt = time.Now()
	if err := l.userDAO.SaveUser(user); err != nil {
		return nil, "", time.Time{}, err
	}

	return user, token, expireAt, nil
}
