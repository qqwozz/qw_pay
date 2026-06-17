package antifraud

import (
	"encoding/json"
	"testing"
)

func TestVerdict(t *testing.T) {
	v := Verdict{
		ID:        "test-id",
		Approved:  true,
		Reason:    "low risk",
		RiskScore: 10,
		Engine:    "cpp",
	}

	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal verdict: %v", err)
	}

	var decoded Verdict
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal verdict: %v", err)
	}

	if decoded.ID != v.ID {
		t.Error("ID mismatch")
	}
	if decoded.Approved != v.Approved {
		t.Error("Approved mismatch")
	}
	if decoded.Reason != v.Reason {
		t.Error("Reason mismatch")
	}
	if decoded.RiskScore != v.RiskScore {
		t.Error("RiskScore mismatch")
	}
	if decoded.Engine != v.Engine {
		t.Error("Engine mismatch")
	}
}

func TestTransferRequest(t *testing.T) {
	req := TransferRequest{
		ID:          "test-id",
		FromAccount: "from-acc",
		ToAccount:   "to-acc",
		Amount:      100.0,
		Currency:    "RUB",
		UserID:      "user-id",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	var decoded TransferRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal request: %v", err)
	}

	if decoded.ID != req.ID {
		t.Error("ID mismatch")
	}
	if decoded.FromAccount != req.FromAccount {
		t.Error("FromAccount mismatch")
	}
	if decoded.ToAccount != req.ToAccount {
		t.Error("ToAccount mismatch")
	}
	if decoded.Amount != req.Amount {
		t.Error("Amount mismatch")
	}
	if decoded.Currency != req.Currency {
		t.Error("Currency mismatch")
	}
	if decoded.UserID != req.UserID {
		t.Error("UserID mismatch")
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient("localhost:6379")
	if client == nil {
		t.Fatal("client should not be nil")
	}
	if client.rdb == nil {
		t.Error("redis client should not be nil")
	}
}
