package pkg

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nbd-wtf/go-nostr"
)

// DephyDsProxyStatus 表示设备状态
type DephyDsProxyStatus string

const (
	StatusAvailable DephyDsProxyStatus = DephyDsProxyStatus(1)
	StatusWorking   DephyDsProxyStatus = DephyDsProxyStatus(2)
)

// UnmarshalJSON customizes JSON unmarshaling for DephyDsProxyStatus
func (s *DephyDsProxyStatus) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	switch str {
	case "Available":
		*s = StatusAvailable
	case "Working":
		*s = StatusWorking
	default:
		return fmt.Errorf("invalid DephyDsProxyStatus: %s", str)
	}
	return nil
}

// DephyDsProxyStatusReason 表示状态变更原因
type DephyDsProxyStatusReason uint8

const (
	ReasonUserRequest   DephyDsProxyStatusReason = DephyDsProxyStatusReason(1)
	ReasonAdminRequest  DephyDsProxyStatusReason = DephyDsProxyStatusReason(2)
	ReasonUserBehaviour DephyDsProxyStatusReason = DephyDsProxyStatusReason(3)
	ReasonReset         DephyDsProxyStatusReason = DephyDsProxyStatusReason(4)
	ReasonLockFailed    DephyDsProxyStatusReason = DephyDsProxyStatusReason(5)
)

// UnmarshalJSON customizes JSON unmarshaling for DephyDsProxyStatusReason
func (r *DephyDsProxyStatusReason) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	switch str {
	case "UserRequest":
		*r = ReasonUserRequest
	case "AdminRequest":
		*r = ReasonAdminRequest
	case "UserBehaviour":
		*r = ReasonUserBehaviour
	case "Reset":
		*r = ReasonReset
	case "LockFailed":
		*r = ReasonLockFailed

	default:
		return fmt.Errorf("invalid DephyDsProxyStatusReason: %s", str)
	}
	return nil
}

// DephyDsProxyMessage 表示 Nostr 事件的内容
type DephyDsProxyMessage struct {
	Request     *RequestPayload     `json:"Request,omitempty"`
	Status      *StatusPayload      `json:"Status,omitempty"`
	Transaction *TransactionPayload `json:"Transaction,omitempty"`
}

type RequestPayload struct {
	ToStatus       DephyDsProxyStatus       `json:"to_status"`
	Reason         DephyDsProxyStatusReason `json:"reason"`
	InitialRequest string                   `json:"initial_request"`
	Payload        string                   `json:"payload"`
}

type StatusPayload struct {
	Status         DephyDsProxyStatus       `json:"status"`
	Reason         DephyDsProxyStatusReason `json:"reason"`
	InitialRequest string                   `json:"initial_request"`
	Payload        string                   `json:"payload"`
}

type TransactionPayload struct {
	User   string `json:"user"`
	Tokens int64  `json:"tokens"`
}

type NostrClient struct {
	Relay         *nostr.Relay
	Session       string
	MachinePubkey string
	SecretKey     string
}

func NewNostrClient(relayURL, session, machinePubkey, secretKey string) (*NostrClient, error) {
	ctx := context.Background()
	relay, err := nostr.RelayConnect(ctx, relayURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to relay: %v", err)
	}

	return &NostrClient{
		Relay:         relay,
		Session:       session,
		MachinePubkey: machinePubkey,
		SecretKey:     secretKey,
	}, nil
}

func (c *NostrClient) Close() {
	c.Relay.Close()
}
