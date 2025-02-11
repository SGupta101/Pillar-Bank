package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pillar-bank/models"
	"pillar-bank/testdata"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestDB() *sql.DB {
	// Use environment variables for test database
	dbName := "pillar_bank_test"
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

	return db
}

func cleanTestDB(db *sql.DB) error {
	_, err := db.Exec("TRUNCATE wire_messages RESTART IDENTITY")
	return err
}

func TestPostWireMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB()

	err := cleanTestDB(db)
	if err != nil {
		t.Fatal(err)
	}

	h := &Handler{db: db}
	router := gin.Default()
	router.POST("/wire-messages", h.postWireMessage)

	for _, tt := range testdata.ValidMessages {
		t.Run(tt.Name, func(t *testing.T) {
			// Send the wire message directly without JSON wrapping
			req, _ := http.NewRequest(http.MethodPost, "/wire-messages", strings.NewReader(tt.WireMessage))
			req.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)
			var response models.WireMessage
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			// Rest of the assertions remain the same
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
			// Send invalid messages directly without JSON wrapping
			req, _ := http.NewRequest(http.MethodPost, "/wire-messages", strings.NewReader(tt.WireMessage))
			req.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.JSONEq(t, tt.ExpectedError, w.Body.String(), "response body mismatch")
		})
	}
}

func TestGetWireMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{db: setupTestDB()}
	router := gin.Default()
	router.GET("/wire-message/:seq", h.getWireMessage)

	t.Run("Get existing wire message", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/wire-message/2", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response models.WireMessage
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		expectedMessage := testdata.ValidMessages[1].Expected
		assert.Equal(t, expectedMessage.Seq, response.Seq)
		assert.Equal(t, expectedMessage.SenderRTN, response.SenderRTN)
		assert.Equal(t, expectedMessage.SenderAN, response.SenderAN)
		assert.Equal(t, expectedMessage.ReceiverRTN, response.ReceiverRTN)
		assert.Equal(t, expectedMessage.ReceiverAN, response.ReceiverAN)
		assert.Equal(t, expectedMessage.Amount, response.Amount)
	})

	t.Run("Get non-existent wire message", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/wire-message/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, `{"error": "Wire message not found"}`, w.Body.String())
	})
}

func TestGetWireMessages(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{db: setupTestDB()}
	router := gin.Default()
	router.GET("/wire-messages", h.getWireMessages)

	t.Run("Get all wire messages", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/wire-messages", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []models.WireMessage
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, len(testdata.ValidMessages), len(response))
		for i, msg := range response {
			assert.Equal(t, testdata.ValidMessages[i].Expected.Seq, msg.Seq)
			assert.Equal(t, testdata.ValidMessages[i].Expected.SenderRTN, msg.SenderRTN)
			assert.Equal(t, testdata.ValidMessages[i].Expected.SenderAN, msg.SenderAN)
			assert.Equal(t, testdata.ValidMessages[i].Expected.ReceiverRTN, msg.ReceiverRTN)
			assert.Equal(t, testdata.ValidMessages[i].Expected.ReceiverAN, msg.ReceiverAN)
			assert.Equal(t, testdata.ValidMessages[i].Expected.Amount, msg.Amount)
		}
	})

	t.Run("Get invalid page number", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/wire-messages?page=0&limit=2", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error": "Invalid page number"}`, w.Body.String())
	})

	t.Run("Get invalid limit number", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/wire-messages?page=1&limit=-2", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error": "Invalid limit number"}`, w.Body.String())
	})

	t.Run("Get wire messages page 1 with limit 2", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/wire-messages?page=1&limit=2", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []models.WireMessage
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(response))
		for i, msg := range response {
			assert.Equal(t, testdata.ValidMessages[i].Expected.Seq, msg.Seq)
			assert.Equal(t, testdata.ValidMessages[i].Expected.SenderRTN, msg.SenderRTN)
			assert.Equal(t, testdata.ValidMessages[i].Expected.SenderAN, msg.SenderAN)
			assert.Equal(t, testdata.ValidMessages[i].Expected.ReceiverRTN, msg.ReceiverRTN)
			assert.Equal(t, testdata.ValidMessages[i].Expected.ReceiverAN, msg.ReceiverAN)
			assert.Equal(t, testdata.ValidMessages[i].Expected.Amount, msg.Amount)
		}
	})

	t.Run("Get wire messages page 2 with limit 2", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/wire-messages?page=2&limit=2", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []models.WireMessage
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(response))
		for i, msg := range response {
			assert.Equal(t, testdata.ValidMessages[i+2].Expected.Seq, msg.Seq)
			assert.Equal(t, testdata.ValidMessages[i+2].Expected.SenderRTN, msg.SenderRTN)
			assert.Equal(t, testdata.ValidMessages[i+2].Expected.SenderAN, msg.SenderAN)
			assert.Equal(t, testdata.ValidMessages[i+2].Expected.ReceiverRTN, msg.ReceiverRTN)
			assert.Equal(t, testdata.ValidMessages[i+2].Expected.ReceiverAN, msg.ReceiverAN)
			assert.Equal(t, testdata.ValidMessages[i+2].Expected.Amount, msg.Amount)
		}
	})

	t.Run("Get wire messages page 3 with limit 2", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/wire-messages?page=3&limit=2", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []models.WireMessage
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(response))
		for i, msg := range response {
			assert.Equal(t, testdata.ValidMessages[i+4].Expected.Seq, msg.Seq)
			assert.Equal(t, testdata.ValidMessages[i+4].Expected.SenderRTN, msg.SenderRTN)
			assert.Equal(t, testdata.ValidMessages[i+4].Expected.SenderAN, msg.SenderAN)
			assert.Equal(t, testdata.ValidMessages[i+4].Expected.ReceiverRTN, msg.ReceiverRTN)
			assert.Equal(t, testdata.ValidMessages[i+4].Expected.ReceiverAN, msg.ReceiverAN)
			assert.Equal(t, testdata.ValidMessages[i+4].Expected.Amount, msg.Amount)
		}
	})
}

func TestLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/login", login)

	t.Run("Valid credentials", func(t *testing.T) {
		data := bytes.NewBufferString(`username=user1&password=password1`)
		req, _ := http.NewRequest(http.MethodPost, "/login", data)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"message":"Successfully logged in"}`, w.Body.String())

		cookie := w.Result().Cookies()
		assert.NotEmpty(t, cookie, "Cookie should be set")
		assert.Equal(t, "token", cookie[0].Name)
		assert.NotEmpty(t, cookie[0].Value, "Token should be set")
	})

	t.Run("Invalid credentials", func(t *testing.T) {
		data := bytes.NewBufferString(`username=user1&password=wrong-password`)
		req, _ := http.NewRequest(http.MethodPost, "/login", data)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.JSONEq(t, `{"error":"Invalid credentials"}`, w.Body.String())
	})

}
