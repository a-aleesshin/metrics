package audit

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	serveraudit "github.com/a-aleesshin/metrics/internal/server/audit"
)

func TestFileObserver_Notify_AppendsJSONLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")

	observer, err := NewFileObserver(path)
	if err != nil {
		t.Fatalf("create file observer: %v", err)
	}
	defer observer.Close()

	events := []serveraudit.Event{
		{TS: 1, Metrics: []string{"Alloc"}, IPAddress: "10.0.0.1"},
		{TS: 2, Metrics: []string{"Frees", "Alloc"}, IPAddress: "10.0.0.2"},
	}

	for _, event := range events {
		if err := observer.Notify(context.Background(), event); err != nil {
			t.Fatalf("notify: %v", err)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read audit file: %v", err)
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), string(data))
	}

	for i, line := range lines {
		var got serveraudit.Event
		if err := json.Unmarshal([]byte(line), &got); err != nil {
			t.Fatalf("line %d is not valid json: %v", i, err)
		}

		if got.TS != events[i].TS || got.IPAddress != events[i].IPAddress {
			t.Fatalf("line %d: expected %+v, got %+v", i, events[i], got)
		}
	}
}

func TestFileObserver_Notify_AppendsToExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")

	if err := os.WriteFile(path, []byte("existing line\n"), 0o644); err != nil {
		t.Fatalf("prepare file: %v", err)
	}

	observer, err := NewFileObserver(path)
	if err != nil {
		t.Fatalf("create file observer: %v", err)
	}
	defer observer.Close()

	if err := observer.Notify(context.Background(), serveraudit.Event{TS: 1}); err != nil {
		t.Fatalf("notify: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read audit file: %v", err)
	}

	if !strings.HasPrefix(string(data), "existing line\n") {
		t.Fatalf("expected existing content to be preserved, got %q", string(data))
	}
}
