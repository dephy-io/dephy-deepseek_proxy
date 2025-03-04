package dao

import (
	"dsproxy-backend/models"

	"gorm.io/gorm"
)

// TransactionEventDAO handles Transaction event storage
type TransactionEventDAO struct {
	db *gorm.DB
}

func NewTransactionEventDAO(db *gorm.DB) *TransactionEventDAO {
	return &TransactionEventDAO{db: db}
}

func (d *TransactionEventDAO) SaveTransactionEvent(event *models.TransactionEvent) error {
	return d.db.Create(event).Error
}

func (d *TransactionEventDAO) GetAllTransactionEvents() ([]models.TransactionEvent, error) {
	var events []models.TransactionEvent
	if err := d.db.Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}
