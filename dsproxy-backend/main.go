package main

import (
	"log"

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
	// Initialize database
	dsn := "host=localhost user=postgres password=your_password dbname=your_db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate models
	db.AutoMigrate(&models.User{}, &models.Conversation{}, &models.Message{})

	// Initialize DAOs
	userDAO := dao.NewUserDAO(db)
	convoDAO := dao.NewConversationDAO(db)
	messageDAO := dao.NewMessageDAO(db)
	txEventDAO := dao.NewTransactionEventDAO(db)

	// Initialize Nostr client (placeholder, use real implementation)
	nostrClient, err := pkg.NewNostrClient("wss://relay.stoner.com", "controller_pubkey", "machine_pubkey")
	if err != nil {
		log.Fatalf("Failed to initialize Nostr client: %v", err)
	}

	// Initialize Chat client (placeholder, use real implementation)
	chatClient := pkg.NewChatClient("api_key")

	// Initialize Logic
	logicLayer := logic.NewConversationLogic(userDAO, convoDAO, messageDAO, txEventDAO, chatClient, nostrClient)

	// Initialize Controller
	ctrl := controller.NewConversationController(logicLayer)

	// Start Nostr listener in a goroutine
	go ctrl.StartNostrListener()

	// Setup Gin router
	r := gin.Default()
	r.POST("/conversations", ctrl.CreateConversation)
	r.POST("/conversations/:id/messages", ctrl.AddMessage)
	r.GET("/conversations/:id/messages", ctrl.GetMessages)

	// Run server
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
