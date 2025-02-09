package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type WireMessage struct {
	ID          int       `json:"id"`
	Seq         int       `json:"seq"`
	SenderRTN   string    `json:"sender_rtn"`
	SenderAN    string    `json:"sender_an"`
	ReceiverRTN string    `json:"receiver_rtn"`
	ReceiverAN  string    `json:"receiver_an"`
	Amount      int       `json:"amount"`
	RawMessage  string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
}

type Handler struct {
	db *sql.DB
}

func main() {
	password := os.Getenv("DB_PASSWORD")
	connStr := fmt.Sprintf("postgres://postgres:%s@localhost:5432/pillar_bank?sslmode=disable", password)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	h := &Handler{
		db: db,
	}

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, gin.H{
			"message": "API is working",
		})
	})

	router.GET("/wire-messages", h.getWireMessages)
	router.GET("/wire-messages/:seq", h.getWireMessage)
	router.POST("/wire-messages", h.postWireMessage)

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

func parseWireMessage(message string) (WireMessage, error) {
	wireMessage := WireMessage{}
	parts := strings.Split(message, ";")

	if len(parts) != 6 {
		return wireMessage, fmt.Errorf("invalid message format: must contain all information")
	}

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

func (h *Handler) sequenceNumberExists(seq int) (bool, error) {
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM wire_messages WHERE seq = $1)", seq).Scan(&exists)
	return exists, err
}

func (h *Handler) postWireMessage(c *gin.Context) {
	var request struct {
		Message string `json:"message"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	wireMessage, err := parseWireMessage(request.Message)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	exists, err := h.sequenceNumberExists(wireMessage.Seq)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to check sequence number: %v", err)})
		return
	}
	if exists {
		c.IndentedJSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("sequence number %d already exists", wireMessage.Seq)})
		return
	}

	query := `INSERT INTO wire_messages (seq, sender_rtn, sender_an, receiver_rtn, receiver_an, amount, raw_message) 
			 VALUES ($1, $2, $3, $4, $5, $6, $7) 
			 RETURNING id, created_at`
	err = h.db.QueryRow(query, wireMessage.Seq, wireMessage.SenderRTN, wireMessage.SenderAN, wireMessage.ReceiverRTN, wireMessage.ReceiverAN, wireMessage.Amount, wireMessage.RawMessage).Scan(&wireMessage.ID, &wireMessage.CreatedAt)

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to insert wire message: %v", err)})
		return
	}

	c.IndentedJSON(http.StatusCreated, wireMessage)
}

func (h *Handler) getWireMessages(c *gin.Context) {
	var wireMessages []WireMessage
	limit := 10
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))

	if err != nil || page < 0 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
	}

	offset := (page - 1) * limit
	query := "SELECT * FROM wire_messages ORDER BY seq ASC LIMIT $1 OFFSET $2"
	rows, err := h.db.Query(query, limit, offset)

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var wm WireMessage
		err := rows.Scan(&wm.ID, &wm.Seq, &wm.SenderRTN, &wm.SenderAN, &wm.ReceiverRTN, &wm.ReceiverAN, &wm.Amount, &wm.RawMessage, &wm.CreatedAt)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		wireMessages = append(wireMessages, wm)
	}

	if len(wireMessages) == 0 {
		c.IndentedJSON(http.StatusOK, gin.H{"message": "No wire messages found"})
		return
	}

	c.IndentedJSON(http.StatusOK, wireMessages)
}

func (h *Handler) getWireMessage(c *gin.Context) {
	var wireMessage WireMessage
	seq := c.Param("seq")

	seqNum, err := strconv.Atoi(seq)
	if err != nil {
		fmt.Printf("conversion error: %v\n", err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid sequence number format"})
		return
	}

	query := "SELECT * FROM wire_messages WHERE seq = $1;"
	err = h.db.QueryRow(query, seqNum).Scan(
		&wireMessage.ID, &wireMessage.Seq, &wireMessage.SenderRTN,
		&wireMessage.SenderAN, &wireMessage.ReceiverRTN, &wireMessage.ReceiverAN,
		&wireMessage.Amount, &wireMessage.RawMessage, &wireMessage.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Wire message not found"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, wireMessage)
}
