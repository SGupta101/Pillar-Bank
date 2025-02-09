package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"

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
	Amount      int    `json:"amount"`
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

func isInt(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

// parseWireMessage parses a wire message string and returns a WireMessage struct and error
func parseWireMessage(message string) (WireMessage, error) {
	wireMessage := WireMessage{}
	parts := strings.Split(message, ";")

	for _, part := range parts {
		keyValue := strings.Split(part, "=")
		if len(keyValue) != 2 {
			continue
		}

		key := strings.TrimSpace(keyValue[0])
		value := strings.TrimSpace(keyValue[1])

		switch key {
		case "SEQ":
			if !isInt(value) {
				return wireMessage, fmt.Errorf("invalid SEQ format: must be numeric")
			}
			seqNum, _ := strconv.Atoi(value)
			wireMessage.Seq = seqNum
		case "SENDER_RTN":
			if !isInt(value) || len(value) != 9 {
				return wireMessage, fmt.Errorf("invalid RTN format: must be exactly 9 digits")
			}
			wireMessage.SenderRTN = value
		case "SENDER_AN":
			if !isInt(value) {
				return wireMessage, fmt.Errorf("invalid AN format: must be numeric")
			}
			wireMessage.SenderAN = value
		case "RECEIVER_RTN":
			if !isInt(value) || len(value) != 9 {
				return wireMessage, fmt.Errorf("invalid RTN format: must be exactly 9 digits")
			}
			wireMessage.ReceiverRTN = value
		case "RECEIVER_AN":
			if !isInt(value) {
				return wireMessage, fmt.Errorf("invalid AN format: must be numeric")
			}
			wireMessage.ReceiverAN = value
		case "AMOUNT":
			if !isInt(value) {
				return wireMessage, fmt.Errorf("invalid amount format: must be numeric")
			}
			amount, _ := strconv.Atoi(value)
			wireMessage.Amount = amount
		}
	}

	wireMessage.RawMessage = message
	return wireMessage, nil
}

// getWireMessages responds with the list of all wire messages as JSON
func getWireMessages(c *gin.Context) {}

// getWireMessage locates the wire message whose Seq matches the seq
// parameter sent by the client, then returns that wire message as JSON
func getWireMessage(c *gin.Context) {}

// postWireMessage adds a new wire message from JSON received in the request body
func postWireMessage(c *gin.Context) {}
