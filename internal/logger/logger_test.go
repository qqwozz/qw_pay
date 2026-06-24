package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestSetup(t *testing.T) {
	var buf bytes.Buffer
	Setup(&buf)

	if L == nil {
		t.Fatal("logger should not be nil after setup")
	}

	Info("test message", "key", "value")

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	if entry["msg"] != "test message" {
		t.Errorf("expected msg 'test message', got %v", entry["msg"])
	}
	if entry["key"] != "value" {
		t.Errorf("expected key 'value', got %v", entry["key"])
	}
}

func TestInfo(t *testing.T) {
	var buf bytes.Buffer
	Setup(&buf)

	Info("info test", "foo", "bar")

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	if entry["msg"] != "info test" {
		t.Errorf("expected msg 'info test', got %v", entry["msg"])
	}
	if entry["foo"] != "bar" {
		t.Errorf("expected foo 'bar', got %v", entry["foo"])
	}
}

func TestError(t *testing.T) {
	var buf bytes.Buffer
	Setup(&buf)

	Error("error test", "reason", "fail")

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	if entry["level"] != "ERROR" {
		t.Errorf("expected level ERROR, got %v", entry["level"])
	}
}

func TestWarn(t *testing.T) {
	var buf bytes.Buffer
	Setup(&buf)

	Warn("warn test", "detail", "something")

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	if entry["level"] != "WARN" {
		t.Errorf("expected level WARN, got %v", entry["level"])
	}
}

func TestDebug(t *testing.T) {
	var buf bytes.Buffer
	Setup(&buf)

	L = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(L)

	Debug("debug test", "x", 1)

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	if entry["level"] != "DEBUG" {
		t.Errorf("expected level DEBUG, got %v", entry["level"])
	}
}
