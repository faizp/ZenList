package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNormalizeStatus(t *testing.T) {
	tests := []struct {
		in      string
		expects string
		wantErr bool
	}{
		{in: "", expects: "TODO"},
		{in: "todo", expects: "TODO"},
		{in: "IN_PROGRESS", expects: "IN_PROGRESS"},
		{in: "BAD", wantErr: true},
	}

	for _, tc := range tests {
		got, err := normalizeStatus(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("normalizeStatus(%q): expected error", tc.in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("normalizeStatus(%q): unexpected error: %v", tc.in, err)
		}
		if got != tc.expects {
			t.Fatalf("normalizeStatus(%q): got %q want %q", tc.in, got, tc.expects)
		}
	}
}

func TestValidateSchedule(t *testing.T) {
	start := time.Date(2026, 2, 17, 10, 0, 0, 0, time.UTC)
	dueValid := start.Add(2 * time.Hour)
	dueInvalid := start.Add(-2 * time.Hour)

	if err := validateSchedule(&start, &dueValid); err != nil {
		t.Fatalf("expected valid schedule, got error: %v", err)
	}
	if err := validateSchedule(&start, &dueInvalid); err == nil {
		t.Fatal("expected invalid schedule error")
	}
}

func TestCursorRoundTrip(t *testing.T) {
	now := time.Date(2026, 2, 17, 11, 45, 0, 0, time.UTC)
	id := uuid.New()

	encoded := encodeCursor(now, id)
	decoded, err := decodeCursor(encoded)
	if err != nil {
		t.Fatalf("decodeCursor returned error: %v", err)
	}
	if !decoded.CreatedAt.Equal(now) {
		t.Fatalf("decoded.CreatedAt = %s, expected %s", decoded.CreatedAt, now)
	}
	if decoded.ID != id {
		t.Fatalf("decoded.ID = %s, expected %s", decoded.ID, id)
	}
}
