package testdata

import (
	"pillar-bank/models"
)

var ValidMessages = []struct {
	Name        string
	WireMessage string
	Expected    models.WireMessage
}{
	{
		Name:        "Valid wire message seq 5",
		WireMessage: "seq=5;sender_rtn=021000021;sender_an=629385443170308;receiver_rtn=121145307;receiver_an=136657407199052;amount=6666",
		Expected: models.WireMessage{
			Seq:         5,
			SenderRTN:   "021000021",
			SenderAN:    "629385443170308",
			ReceiverRTN: "121145307",
			ReceiverAN:  "136657407199052",
			Amount:      6666,
		},
	},
	{
		Name:        "Valid wire message seq 1",
		WireMessage: "seq=1;sender_rtn=021000021;sender_an=537646894897833;receiver_rtn=121145307;receiver_an=669907820975207;amount=3424",
		Expected: models.WireMessage{
			Seq:         1,
			SenderRTN:   "021000021",
			SenderAN:    "537646894897833",
			ReceiverRTN: "121145307",
			ReceiverAN:  "669907820975207",
			Amount:      3424,
		},
	},
	{
		Name:        "Valid wire message seq 2",
		WireMessage: "seq=2;sender_rtn=121000248;sender_an=349848983426759;receiver_rtn=121145307;receiver_an=160661577716921;amount=2123",
		Expected: models.WireMessage{
			Seq:         2,
			SenderRTN:   "121000248",
			SenderAN:    "349848983426759",
			ReceiverRTN: "121145307",
			ReceiverAN:  "160661577716921",
			Amount:      2123,
		},
	},
	{
		Name:        "Valid wire message seq 3",
		WireMessage: "seq=3;sender_rtn=121000248;sender_an=608884434554320;receiver_rtn=121145307;receiver_an=136657407199052;amount=2123",
		Expected: models.WireMessage{
			Seq:         3,
			SenderRTN:   "121000248",
			SenderAN:    "608884434554320",
			ReceiverRTN: "121145307",
			ReceiverAN:  "136657407199052",
			Amount:      2123,
		},
	},
	{
		Name:        "Valid wire message seq 4",
		WireMessage: "seq=4;sender_rtn=021000021;sender_an=629385443170308;receiver_rtn=121145307;receiver_an=136657407199052;amount=1034",
		Expected: models.WireMessage{
			Seq:         4,
			SenderRTN:   "021000021",
			SenderAN:    "629385443170308",
			ReceiverRTN: "121145307",
			ReceiverAN:  "136657407199052",
			Amount:      1034,
		},
	},
}

var InvalidMessages = []struct {
	Name          string
	WireMessage   string
	ExpectedError string
}{
	{
		Name:          "Empty message",
		WireMessage:   "",
		ExpectedError: `{"error": "invalid message format: must contain all information"}`,
	},
	{
		Name:          "Invalid SEQ",
		WireMessage:   "seq=hello world;sender_rtn=1234;sender_an=12345678;receiver_rtn=987654321;receiver_an=87654321;amount=1000",
		ExpectedError: `{"error": "invalid SEQ format: must be numeric"}`,
	},
	{
		Name:          "Duplicate SEQ",
		WireMessage:   "seq=1;sender_rtn=021000021;sender_an=537646894897833;receiver_rtn=121145307;receiver_an=669907820975207;amount=3424",
		ExpectedError: `{"error": "duplicate sequence number 1"}`,
	},
	{
		Name:          "Invalid Sender RTN length",
		WireMessage:   "seq=6;sender_rtn=0021000021;sender_an=12345678;receiver_rtn=121145307;receiver_an=87654321;amount=1000",
		ExpectedError: `{"error": "invalid RTN format: must be exactly 9 digits"}`,
	},
	{
		Name:          "Invalid Sender RTN",
		WireMessage:   "seq=7;sender_rtn=hello world;sender_an=12345678;receiver_rtn=121145307;receiver_an=87654321;amount=1000",
		ExpectedError: `{"error": "invalid RTN format: must be exactly 9 digits"}`,
	},
	{
		Name:          "Invalid Receiver RTN length",
		WireMessage:   "seq=8;sender_rtn=021000021;sender_an=12345678;receiver_rtn=21145307;receiver_an=87654321;amount=1000",
		ExpectedError: `{"error": "invalid RTN format: must be exactly 9 digits"}`,
	},
	{
		Name:          "Invalid Receiver RTN",
		WireMessage:   "seq=9;sender_rtn=021000021;sender_an=12345678;receiver_rtn=hello world;receiver_an=87654321;amount=1000",
		ExpectedError: `{"error": "invalid RTN format: must be exactly 9 digits"}`,
	},
	{
		Name:          "Invalid Amount",
		WireMessage:   "seq=10;sender_rtn=021000021;sender_an=12345678;receiver_rtn=hello world;receiver_an=87654321;amount=-5",
		ExpectedError: `{"error": "invalid RTN format: must be exactly 9 digits"}`,
	},
	{
		Name:          "Invalid Amount type",
		WireMessage:   "seq=11;sender_rtn=021000021;sender_an=629385443170308;receiver_rtn=121145307;receiver_an=136657407199052;amount=hello world",
		ExpectedError: `{"error": "invalid amount format: must be numeric"}`,
	},
}
