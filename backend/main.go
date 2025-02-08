package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type WireMessage struct {
	ID          int    `json:"id"`
	Seq         int    `json:"seq"`
	SenderRTN   string `json:"sender_rtn"`
	SenderAN    string `json:"sender_an"`
	ReceiverRTN string `json:"receiver_rtn"`
	ReceiverAN  string `json:"receiver_an"`
	Amount      string `json:"amount"`
	RawMessage  string `json:"message"`
}

func main() {
	// Get the environment variable
	password := os.Getenv("DB_PASSWORD")
	connStr := fmt.Sprintf("postgres://postgres:%s@localhost:5432/pillar_bank?sslmode=disable", password)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()

	// Wire message routes
	router.GET("/wire-messages", getWireMessages)    // Get all wire messages
	router.GET("/wire-messages/:id", getWireMessage) // Get single wire message
	router.POST("/wire-messages", postWireMessage)   // Create new wire message

	router.Run("localhost:8080")
}

// getWireMessages responds with the list of all wire messages as JSON
func getWireMessages(c *gin.Context) {}

// getWireMessage locates the wire message whose Seq matches the seq
// parameter sent by the client, then returns that wire message as JSON
func getWireMessage(c *gin.Context) {}

// postWireMessage adds a new wire message from JSON received in the request body
func postWireMessage(c *gin.Context) {}
