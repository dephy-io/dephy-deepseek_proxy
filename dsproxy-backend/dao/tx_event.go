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

// GetLatestCreatedAt retrieves the latest created_at timestamp from stored events
func (d *TransactionEventDAO) GetLatestCreatedAt() (int64, error) {
	var latestEvent models.TransactionEvent
	err := d.db.Order("created_at DESC").First(&latestEvent).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// No events yet, return 0 to start from the beginning
			return 0, nil
		}
		return 0, err
	}
	return latestEvent.CreatedAt.Unix(), nil
}
