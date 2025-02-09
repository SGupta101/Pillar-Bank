package models

import "time"

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
