package logic

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"dsproxy-backend/dao"
	"dsproxy-backend/models"
	"dsproxy-backend/pkg"

	"github.com/nbd-wtf/go-nostr"
)

// TxEventLogic handles conversation-related business logic
type TxEventLogic struct {
	userDAO     *dao.UserDAO
	txEventDAO  *dao.TransactionEventDAO
	nostrClient *pkg.NostrClient
}

func NewTxEventLogic(
	userDAO *dao.UserDAO,
	txEventDAO *dao.TransactionEventDAO,
	nostrClient *pkg.NostrClient,
) *TxEventLogic {
	return &TxEventLogic{
		userDAO:     userDAO,
		txEventDAO:  txEventDAO,
		nostrClient: nostrClient,
	}
}

// SyncTransactionEvents syncs historical Transaction events at startup
func (l *TxEventLogic) SyncTransactionEvents(ctx context.Context) error {
	// Fetch stored Transaction events from DB
	storedEvents, err := l.txEventDAO.GetAllTransactionEvents()
	if err != nil {
		return err
	}
	storedEventMap := make(map[string]bool)
	for _, evt := range storedEvents {
		storedEventMap[evt.ID] = true
	}

	// Subscribe to historical Transaction events from Relay
	since := nostr.Timestamp(0) // Fetch all events from the beginning
	filters := nostr.Filters{{
		Kinds: []int{1573},
		Since: &since,
		Tags: nostr.TagMap{
			"s": []string{"dephy-dsproxy-controller"},
			"p": []string{l.nostrClient.MachinePubkey()},
		},
	}}

	sub, err := l.nostrClient.Relay().Subscribe(ctx, filters)
	if err != nil {
		return err
	}
	defer sub.Unsub()

	for ev := range sub.Events {
		var msg pkg.DephyDsProxyMessage
		if err := json.Unmarshal([]byte(ev.Content), &msg); err != nil {
			log.Printf("Failed to parse event content: %v", err)
			continue
		}

		if msg.Transaction != nil {
			eventID := ev.ID
			if !storedEventMap[eventID] {
				// New event, save to DB and update user
				txEvent := &models.TransactionEvent{
					ID:        eventID,
					User:      msg.Transaction.User,
					Lamports:  msg.Transaction.Lamports,
					CreatedAt: time.Unix(int64(ev.CreatedAt), 0),
				}
				if err := l.txEventDAO.SaveTransactionEvent(txEvent); err != nil {
					log.Printf("Failed to save Transaction event: %v", err)
					continue
				}
				if err := l.userDAO.UpdateUserTokens(msg.Transaction.User, int64(msg.Transaction.Lamports), 0); err != nil {
					log.Printf("Failed to update user tokens: %v", err)
				}
			}
		}
	}

	return nil
}

// StartNostrListener starts listening to Nostr Transaction events
func (l *TxEventLogic) StartNostrListener(ctx context.Context) error {
	return l.nostrClient.SubscribeTransactions(ctx, func(event nostr.Event) {
		var msg pkg.DephyDsProxyMessage
		if err := json.Unmarshal([]byte(event.Content), &msg); err != nil {
			log.Printf("Failed to parse event content: %v", err)
			return
		}

		if msg.Transaction != nil {
			txEvent := &models.TransactionEvent{
				ID:        event.ID,
				User:      msg.Transaction.User,
				Lamports:  msg.Transaction.Lamports,
				CreatedAt: time.Unix(int64(event.CreatedAt), 0),
			}
			if err := l.txEventDAO.SaveTransactionEvent(txEvent); err != nil {
				log.Printf("Failed to save Transaction event: %v", err)
				return
			}

			// Update user tokens (1 Lamport = 1 Token)
			err := l.userDAO.UpdateUserTokens(msg.Transaction.User, int64(msg.Transaction.Lamports), 0)
			if err != nil {
				log.Printf("Failed to update user tokens: %v", err)
			} else {
				log.Printf("Updated tokens for user %s: +%d", msg.Transaction.User, msg.Transaction.Lamports)
			}
		}
	})
}
