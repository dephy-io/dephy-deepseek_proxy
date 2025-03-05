package logic

import (
	"dsproxy-backend/dao"
	"dsproxy-backend/models"
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
