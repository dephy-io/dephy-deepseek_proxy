package logic

import (
	"context"
	"encoding/json"
	"fmt"
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

// StartNostrListener starts listening to Nostr Transaction events
func (l *TxEventLogic) StartNostrListener(ctx context.Context) error {
	// Get the latest created_at timestamp from the database
	latestCreatedAt, err := l.txEventDAO.GetLatestCreatedAt()
	if err != nil {
		return fmt.Errorf("failed to get latest created_at: %v", err)
	}

	log.Printf("Nostr events subscribe since: $v", latestCreatedAt)

	since := nostr.Timestamp(latestCreatedAt+1)

	filters := nostr.Filters{{
		Kinds: []int{1573},
		Since: &since,
		Tags: nostr.TagMap{
			"s": []string{l.nostrClient.Session},
			"p": []string{l.nostrClient.MachinePubkey},
		},
	}}

	sub, err := l.nostrClient.Relay.Subscribe(ctx, filters)
	if err != nil {
		return err
	}
	defer sub.Unsub()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ev, ok := <-sub.Events:
			if !ok {
				log.Println("Event channel closed")
				return nil
			}
			var msg pkg.DephyDsProxyMessage
			if err := json.Unmarshal([]byte(ev.Content), &msg); err != nil {
				log.Printf("Failed to parse event content: %v", err)
				continue
			}

			
			if msg.Transaction != nil {
				log.Printf("Transaction event received: %v", ev.ID)
				txEvent := &models.TransactionEvent{
					ID:        ev.ID,
					User:      msg.Transaction.User,
					Lamports:  msg.Transaction.Lamports,
					CreatedAt: time.Unix(int64(ev.CreatedAt), 0),
				}
				if err := l.txEventDAO.SaveTransactionEvent(txEvent); err != nil {
					log.Printf("Failed to save Transaction event: %v", err)
				}

				// Update user tokens (1 Lamport = 1 Token)
				err := l.userDAO.UpdateUserTokens(msg.Transaction.User, int64(msg.Transaction.Lamports), 0)
				if err != nil {
					log.Printf("Failed to update user tokens: %v", err)
				} else {
					log.Printf("Updated tokens for user %s: +%d", msg.Transaction.User, msg.Transaction.Lamports)
				}
			}
		case <-sub.EndOfStoredEvents:
			log.Println("Received EOSE")
		}
	}
}
