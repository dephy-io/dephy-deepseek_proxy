package main

import (
	"fmt"
	"log"
	"os"

	"dsproxy-backend/config"
	"dsproxy-backend/controller"
	"dsproxy-backend/dao"
	"dsproxy-backend/logic"
	"dsproxy-backend/models"
	"dsproxy-backend/pkg"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Initialize config
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <config.yaml>")
	}
	configFile := os.Args[1]
	if err := config.LoadConfig(configFile); err != nil {
		log.Fatalf("Failed to load config from %s: %v", configFile, err)
	}

	// Initialize database
	db, err := gorm.Open(postgres.Open(config.GlobalConfig.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	db.AutoMigrate(&models.User{}, &models.Conversation{}, &models.Message{}, &models.TransactionEvent{})

	// Initialize Nostr client (placeholder, use real implementation)
	nostrClient, err := pkg.NewNostrClient(config.GlobalConfig.Nostr.RelayURL, config.GlobalConfig.Nostr.Session, config.GlobalConfig.Nostr.MachinePubkey)
	if err != nil {
		log.Fatalf("Failed to initialize Nostr client: %v", err)
	}

	// Initialize Chat client (placeholder, use real implementation)
	chatClient := pkg.NewChatClient(config.GlobalConfig.Chat.APIKey)

	// Initialize DAOs
	userDAO := dao.NewUserDAO(db)
	convoDAO := dao.NewConversationDAO(db)
	messageDAO := dao.NewMessageDAO(db)
	txEventDAO := dao.NewTransactionEventDAO(db)

	// Initialize Logics
	userLogic := logic.NewUserLogic(userDAO)
	convoLogic := logic.NewConversationLogic(userDAO, convoDAO)
	messageLogic := logic.NewMessageLogic(userDAO, convoDAO, messageDAO, chatClient)
	txEventLogic := logic.NewTxEventLogic(userDAO, txEventDAO, nostrClient)

	// Initialize Controllers
	userCtrl := controller.NewUserController(userLogic)
	convoCtrl := controller.NewConversationController(convoLogic)
	messageCtrl := controller.NewMessageController(messageLogic)
	txEventCtrl := controller.NewTxEventController(txEventLogic)

	// Start Nostr event listener in a goroutine
	go txEventCtrl.StartNostrServices()

	// Setup Gin router
	r := gin.Default()
	r.GET("/user", userCtrl.GetUser)
	r.POST("/conversations", convoCtrl.CreateConversation)
	r.GET("/conversations", convoCtrl.GetConversations)
	r.POST("/messages", messageCtrl.AddMessage)
	r.GET("/messages", messageCtrl.GetMessages)

	// Run server
	if err := r.Run(fmt.Sprintf(":%d", config.GlobalConfig.Server.Port)); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
