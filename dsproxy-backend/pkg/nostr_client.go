package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/nbd-wtf/go-nostr"
)

// DephyDsProxyStatus 表示设备状态
type DephyDsProxyStatus uint8

const (
	StatusAvailable DephyDsProxyStatus = 1
	StatusWorking   DephyDsProxyStatus = 2
)

// DephyDsProxyStatusReason 表示状态变更原因
type DephyDsProxyStatusReason uint8

const (
	ReasonUserRequest   DephyDsProxyStatusReason = 1
	ReasonAdminRequest  DephyDsProxyStatusReason = 2
	ReasonUserBehaviour DephyDsProxyStatusReason = 3
	ReasonReset         DephyDsProxyStatusReason = 4
	ReasonLockFailed    DephyDsProxyStatusReason = 5
)

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
	User     string `json:"user"`
	Lamports uint64 `json:"lamports"`
}

type NostrClient struct {
	relay            *nostr.Relay
	controllerPubkey string
	machinePubkey    string
}

func NewNostrClient(relayURL, controllerPubkey, machinePubkey string) (*NostrClient, error) {
	ctx := context.Background()
	relay, err := nostr.RelayConnect(ctx, relayURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to relay: %v", err)
	}

	return &NostrClient{
		relay:            relay,
		controllerPubkey: controllerPubkey,
		machinePubkey:    machinePubkey,
	}, nil
}

func (c *NostrClient) Relay() *nostr.Relay {
	return c.relay
}

func (c *NostrClient) MachinePubkey() string {
	return c.machinePubkey
}

func (c *NostrClient) SubscribeTransactions(ctx context.Context, handler func(event nostr.Event)) error {
	since := nostr.Now()

	filters := nostr.Filters{{
		Kinds: []int{1573}, 
		Since: &since,
		Tags: nostr.TagMap{
			"s": []string{"dephy-dsproxy-controller"},
			"p": []string{c.machinePubkey},
		},
	}}

	sub, err := c.relay.Subscribe(ctx, filters)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %v", err)
	}

	go func() {
		defer sub.Unsub() 
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-sub.Events:
				if !ok {
					log.Println("Event channel closed")
					return
				}
				var msg DephyDsProxyMessage
				if err := json.Unmarshal([]byte(ev.Content), &msg); err != nil {
					log.Printf("Failed to parse event content: %v", err)
					continue
				}

				// 只处理 Transaction 事件
				if msg.Transaction != nil {
					handler(*ev)
				}
			case <-sub.EndOfStoredEvents:
				log.Println("Received EOSE")
			}
		}
	}()

	return nil
}

// Close 关闭客户端连接
func (c *NostrClient) Close() {
	c.relay.Close()
}

// // 示例使用
// func main() {
//     // 配置参数
//     relayURL := "wss://relay.stoner.com"
//     controllerPubkey := "your_controller_pubkey" // 替换为实际的公钥
//     machinePubkey := "your_machine_pubkey"      // 替换为实际的机器公钥

//     // 创建客户端
//     client, err := NewNostrClient(relayURL, controllerPubkey, machinePubkey)
//     if err != nil {
//         log.Fatalf("Failed to create client: %v", err)
//     }
//     defer client.Close()

//     // 创建带有超时的上下文
//     ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//     defer cancel()

//     // 订阅 Transaction 事件
//     err = client.SubscribeTransactions(ctx, func(event nostr.Event) {
//         var msg DephyDsProxyMessage
//         if err := json.Unmarshal([]byte(event.Content), &msg); err != nil {
//             log.Printf("Failed to parse event: %v", err)
//             return
//         }

//         if msg.Transaction != nil {
//             fmt.Printf("Received Transaction event:\n")
//             fmt.Printf("User: %s\n", msg.Transaction.User)
//             fmt.Printf("Lamports: %d\n", msg.Transaction.Lamports)
//             fmt.Printf("Event ID: %s\n", event.ID)
//             fmt.Printf("Created at: %s\n", time.Unix(int64(event.CreatedAt), 0))
//         }
//     })
//     if err != nil {
//         log.Fatalf("Failed to subscribe: %v", err)
//     }

//     // 等待订阅完成或超时
//     <-ctx.Done()
//     fmt.Println("Subscription ended")
// }
