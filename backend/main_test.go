package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"pillar-bank/models"
	"pillar-bank/testdata"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestDB() *sql.DB {
	// Use environment variables for test database
	dbName := os.Getenv("TEST_DB_NAME")
	connStr := fmt.Sprintf("postgres://postgres@localhost:5432/%s?sslmode=disable", dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Create table if not exists
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

	// Clean the database
	_, err = db.Exec("TRUNCATE wire_messages RESTART IDENTITY")
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func TestPostWireMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{db: setupTestDB()}
	router := gin.Default()
	router.POST("/wire-messages", h.postWireMessage)

	for _, tt := range testdata.ValidMessages {
		t.Run(tt.Name, func(t *testing.T) {
			reqBody := fmt.Sprintf(`{"message": "%s"}`, tt.WireMessage)
			req, _ := http.NewRequest(http.MethodPost, "/wire-messages", strings.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)
			var response models.WireMessage
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			// Check all fields
			assert.Equal(t, tt.Expected.Seq, response.Seq, "seq mismatch")
			assert.Equal(t, tt.Expected.SenderRTN, response.SenderRTN, "senderRTN mismatch")
			assert.Equal(t, tt.Expected.SenderAN, response.SenderAN, "senderAN mismatch")
			assert.Equal(t, tt.Expected.ReceiverRTN, response.ReceiverRTN, "receiverRTN mismatch")
			assert.Equal(t, tt.Expected.ReceiverAN, response.ReceiverAN, "receiverAN mismatch")
			assert.Equal(t, tt.Expected.Amount, response.Amount, "amount mismatch")
		})
	}

	for _, tt := range testdata.InvalidMessages {
		t.Run(tt.Name, func(t *testing.T) {
			reqBody := fmt.Sprintf(`{"message": "%s"}`, tt.WireMessage)
			req, _ := http.NewRequest(http.MethodPost, "/wire-messages", strings.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.JSONEq(t, tt.ExpectedError, w.Body.String(), "response body mismatch")
		})
	}
}
