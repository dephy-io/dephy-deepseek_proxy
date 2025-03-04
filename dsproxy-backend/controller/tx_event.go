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

// StartServices starts Nostr listener
func (c *TxEventController) StartNostrServices() {
	ctx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
	defer cancel()

	if err := c.txEventLogic.StartNostrListener(ctx); err != nil {
		log.Printf("Nostr listener failed: %v", err)
	}
	// if err := c.txEventLogic.StartNostrListener(ctx); err != nil {
	// 	log.Printf("Nostr listener failed: %v", err)
	// }
}
