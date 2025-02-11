package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"pillar-bank/auth"
	"pillar-bank/models"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Handler struct {
	db *sql.DB
}

func handleError(c *gin.Context, status int, message string) {
	c.IndentedJSON(status, gin.H{"error": message})
}

func main() {
	connStr := "postgres://postgres@localhost:5432/pillar_bank?sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create table if it doesn't exist
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS wire_messages (
            id SERIAL PRIMARY KEY,
            seq INTEGER UNIQUE NOT NULL,
            sender_rtn VARCHAR(9) NOT NULL,
            sender_an VARCHAR(255) NOT NULL,
            receiver_rtn VARCHAR(9) NOT NULL,
            receiver_an VARCHAR(255) NOT NULL,
            amount INTEGER NOT NULL,
            raw_message TEXT NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
	if err != nil {
		log.Fatal(err)
	}

	h := &Handler{
		db: db,
	}

	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	router.GET("/health", func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, gin.H{
			"message": "API is working",
		})
	})

	// Create route for user authentication
	router.POST("/login", login)
	router.GET("/wire-messages", auth.AuthenticateMiddleware, h.getWireMessages)
	router.GET("/wire-message/:seq", auth.AuthenticateMiddleware, h.getWireMessage)
	router.POST("/wire-messages", auth.AuthenticateMiddleware, h.postWireMessage)

	router.Run("localhost:8080")
}

func login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	// Consider using a map for credentials (still not secure, but cleaner)
	validCredentials := map[string]string{
		"user1": "password1",
		"user2": "password2",
	}

	if storedPassword, exists := validCredentials[username]; exists && storedPassword == password {
		tokenString, err := auth.CreateToken(username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating token"})
			return
		}

		c.SetCookie("token", tokenString, 900, "/", "localhost", false, true) // 900 seconds = 15 minutes
		c.JSON(http.StatusOK, gin.H{"message": "Successfully logged in"})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
	}
}

func isInt(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

func parseWireMessage(message string) (models.WireMessage, error) {
	wireMessage := models.WireMessage{}
	parts := strings.Split(message, ";")
	fmt.Println("Received message:", message)
	fmt.Println("Parts:", parts)

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
		case "seq":
			if !isInt(value) {
				return wireMessage, fmt.Errorf("invalid SEQ format: must be numeric")
			}
			seqNum, _ := strconv.Atoi(value)
			wireMessage.Seq = seqNum
		case "sender_rtn":
			if !isInt(value) || len(value) != 9 {
				return wireMessage, fmt.Errorf("invalid RTN format: must be exactly 9 digits")
			}
			wireMessage.SenderRTN = value
		case "sender_an":
			if !isInt(value) {
				return wireMessage, fmt.Errorf("invalid AN format: must be numeric")
			}
			wireMessage.SenderAN = value
		case "receiver_rtn":
			if !isInt(value) || len(value) != 9 {
				return wireMessage, fmt.Errorf("invalid RTN format: must be exactly 9 digits")
			}
			wireMessage.ReceiverRTN = value
		case "receiver_an":
			if !isInt(value) {
				return wireMessage, fmt.Errorf("invalid AN format: must be numeric")
			}
			wireMessage.ReceiverAN = value
		case "amount":
			if !isInt(value) {
				return wireMessage, fmt.Errorf("invalid amount format: must be numeric")
			}
			amount, _ := strconv.Atoi(value)

			if amount < 0 {
				return wireMessage, fmt.Errorf("invalid amount format: must be positive")
			}

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
	// Read the raw message directly from the body
	message, err := c.GetRawData()
	if err != nil {
		handleError(c, http.StatusBadRequest, "Failed to read message")
		return
	}

	// Parse the wire message from the raw string
	wireMessage, err := parseWireMessage(string(message))
	if err != nil {
		handleError(c, http.StatusBadRequest, err.Error())
		return
	}

	exists, err := h.sequenceNumberExists(wireMessage.Seq)
	if err != nil {
		handleError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check sequence number: %v", err))
		return
	}
	if exists {
		handleError(c, http.StatusBadRequest, fmt.Sprintf("duplicate sequence number %d", wireMessage.Seq))
		return
	}

	query := `INSERT INTO wire_messages (seq, sender_rtn, sender_an, receiver_rtn, receiver_an, amount, raw_message) 
			 VALUES ($1, $2, $3, $4, $5, $6, $7) 
			 RETURNING id, created_at`
	err = h.db.QueryRow(query, wireMessage.Seq, wireMessage.SenderRTN, wireMessage.SenderAN, wireMessage.ReceiverRTN, wireMessage.ReceiverAN, wireMessage.Amount, wireMessage.RawMessage).Scan(&wireMessage.ID, &wireMessage.CreatedAt)

	if err != nil {
		handleError(c, http.StatusInternalServerError, fmt.Sprintf("failed to insert wire message: %v", err))
		return
	}

	c.IndentedJSON(http.StatusCreated, wireMessage)
}

func (h *Handler) getWireMessages(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		handleError(c, http.StatusBadRequest, "Invalid page number")
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		handleError(c, http.StatusBadRequest, "Invalid limit number")
		return
	}

	offset := (page - 1) * limit
	query := "SELECT * FROM wire_messages ORDER BY seq ASC LIMIT $1 OFFSET $2"
	rows, err := h.db.Query(query, limit, offset)

	if err != nil {
		handleError(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var wireMessages []models.WireMessage
	for rows.Next() {
		var wm models.WireMessage
		err := rows.Scan(&wm.ID, &wm.Seq, &wm.SenderRTN, &wm.SenderAN, &wm.ReceiverRTN, &wm.ReceiverAN, &wm.Amount, &wm.RawMessage, &wm.CreatedAt)
		if err != nil {
			handleError(c, http.StatusInternalServerError, err.Error())
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
	var wireMessage models.WireMessage
	seq := c.Param("seq")

	seqNum, err := strconv.Atoi(seq)
	if err != nil {
		handleError(c, http.StatusBadRequest, "Invalid sequence number format")
		return
	}

	query := "SELECT * FROM wire_messages WHERE seq = $1;"
	err = h.db.QueryRow(query, seqNum).Scan(
		&wireMessage.ID, &wireMessage.Seq, &wireMessage.SenderRTN,
		&wireMessage.SenderAN, &wireMessage.ReceiverRTN, &wireMessage.ReceiverAN,
		&wireMessage.Amount, &wireMessage.RawMessage, &wireMessage.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			handleError(c, http.StatusNotFound, "Wire message not found")
			return
		}
		handleError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.IndentedJSON(http.StatusOK, wireMessage)
}
