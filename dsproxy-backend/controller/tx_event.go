package controller

import (
	"context"
	"log"
	"time"

	"dsproxy-backend/logic"
)

// TxEventController handles HTTP requests
type TxEventController struct {
	txEventLogic *logic.TxEventLogic
}

func NewTxEventController(logic *logic.TxEventLogic) *TxEventController {
	return &TxEventController{txEventLogic: logic}
}

// StartNostrListener starts the Nostr event listener
func (c *TxEventController) StartNostrListener() {
	ctx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
	defer cancel()
	if err := c.txEventLogic.StartNostrListener(ctx); err != nil {
		log.Fatalf("Failed to start Nostr listener: %v", err)
	}
}
