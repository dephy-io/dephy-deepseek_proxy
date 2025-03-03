package dao

import (
    "github.com/google/uuid"
    "gorm.io/gorm"
    "dsproxy-backend/models"
)

// UserDAO handles user-related database operations
type UserDAO struct {
    db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
    return &UserDAO{db: db}
}

// CreateUser creates a new user
func (d *UserDAO) CreateUser(publicKey string) (*models.User, error) {
    user := &models.User{PublicKey: publicKey}
    if err := d.db.Create(user).Error; err != nil {
        return nil, err
    }
    return user, nil
}

// GetUserByPublicKey retrieves a user by public key
func (d *UserDAO) GetUserByPublicKey(publicKey string) (*models.User, error) {
    var user models.User
    if err := d.db.Where("public_key = ?", publicKey).First(&user).Error; err != nil {
        return nil, err
    }
    return &user, nil
}

// UpdateUserTokens updates user's token balance and consumed tokens
func (d *UserDAO) UpdateUserTokens(publicKey string, tokensDelta int64, consumedDelta int64) error {
    return d.db.Model(&models.User{}).
        Where("public_key = ?", publicKey).
        Updates(map[string]interface{}{
            "tokens":          gorm.Expr("tokens + ?", tokensDelta),
            "tokens_consumed": gorm.Expr("tokens_consumed + ?", consumedDelta),
        }).Error
}
