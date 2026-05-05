package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func captureRootStderr(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stderr = w
	defer func() { os.Stderr = orig }()

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()
	_ = w.Close()
	return <-done
}

func TestWriteRootErrorEventsUsesNDJSON(t *testing.T) {
	raw := captureRootStderr(t, func() {
		writeRootError(rootFlags{events: true}, errors.New("boom"))
	})

	var evt struct {
		Event string         `json:"event"`
		Data  map[string]any `json:"data"`
	}
	if err := json.Unmarshal([]byte(raw), &evt); err != nil {
		t.Fatalf("root error was not NDJSON: %q: %v", raw, err)
	}
	if evt.Event != "error" {
		t.Fatalf("event = %q, want error", evt.Event)
	}
	if evt.Data["message"] != "boom" {
		t.Fatalf("message = %#v, want boom", evt.Data["message"])
	}
}

func TestRootFlagsReadOnlyFlag(t *testing.T) {
	flags := &rootFlags{readOnly: true}

	if !flags.isReadOnly() {
		t.Fatal("isReadOnly = false, want true")
	}
	err := flags.requireWritable()
	if err == nil || !strings.Contains(err.Error(), "read-only mode") {
		t.Fatalf("requireWritable error = %v", err)
	}
}

func TestRootFlagsReadOnlyEnv(t *testing.T) {
	t.Setenv("WACLI_READONLY", "yes")

	if !(&rootFlags{}).isReadOnly() {
		t.Fatal("isReadOnly = false, want true")
	}
}
