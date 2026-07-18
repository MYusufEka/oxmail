package mail

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestParseMailqOutput_Empty(t *testing.T) {
	status := parseMailqOutput("Mail queue is empty")
	if status.Total != 0 || status.Deferred != 0 || status.Active != 0 {
		t.Errorf("empty queue: got %+v, want zero status", status)
	}
	if status.OldestAge != "" {
		t.Errorf("empty queue: got oldestAge=%q, want empty", status.OldestAge)
	}
}

func TestParseMailqOutput_SingleDeferredEntry(t *testing.T) {
	// Deferred detection requires (deferred, in continuation line
	input := `-Queue ID- --Size-- ----Arrival Time---- --Sender/Recipient------
12345     1024 Sun Jan 15 11:23:45  alice@example.com
                                         (deferred, connection timeout)`

	status := parseMailqOutput(input)
	if status.Total != 1 {
		t.Errorf("Total = %d, want 1", status.Total)
	}
	if status.Deferred != 1 {
		t.Errorf("Deferred = %d, want 1", status.Deferred)
	}
	if status.Active != 0 {
		t.Errorf("Active = %d, want 0", status.Active)
	}
	if status.OldestAge == "" {
		t.Error("OldestAge should not be empty")
	}
}

func TestParseMailqOutput_ActiveEntry(t *testing.T) {
	input := `-Queue ID- --Size-- ----Arrival Time---- --Sender/Recipient------
12345*    1024 Sun Jan 15 11:23:45  alice@example.com`

	status := parseMailqOutput(input)
	if status.Total != 1 {
		t.Errorf("Total = %d, want 1", status.Total)
	}
	if status.Active != 1 {
		t.Errorf("Active = %d, want 1 (marked with *)", status.Active)
	}
	if status.Deferred != 0 {
		t.Errorf("Deferred = %d, want 0", status.Deferred)
	}
}

func TestParseMailqOutput_ActiveFallback(t *testing.T) {
	// No * flag and no deferred markers → active via Total - Deferred = 1
	input := `-Queue ID- --Size-- ----Arrival Time---- --Sender/Recipient------
12345     1024 Sun Jan 15 11:23:45  alice@example.com`

	status := parseMailqOutput(input)
	if status.Total != 1 {
		t.Errorf("Total = %d, want 1", status.Total)
	}
	if status.Active != 1 {
		t.Errorf("Active = %d, want 1 (no * so all are active)", status.Active)
	}
}

func TestParseMailqOutput_MixedEntries(t *testing.T) {
	// DEFER1 has (deferred, and DEFER2 has (deferred) = both detected as deferred
	input := `-Queue ID- --Size-- ----Arrival Time---- --Sender/Recipient------
ACTIVE1*   512  Sun Jan 15 11:23:45  bob@example.com
DEFER1    1024  Sun Jan 15 11:23:45  carol@example.com
                                        (deferred, connection timeout)
ACTIVE2*   256  Sun Jan 15 11:23:45  dave@example.com
DEFER2     768  Sun Jan 15 11:23:45  eve@example.com
                                        (deferred, HELO rejected)`

	status := parseMailqOutput(input)
	if status.Total != 4 {
		t.Errorf("Total = %d, want 4", status.Total)
	}
	if status.Active != 2 {
		t.Errorf("Active = %d, want 2", status.Active)
	}
	if status.Deferred != 2 {
		t.Errorf("Deferred = %d, want 2", status.Deferred)
	}
}

func TestParseMailqOutput_OnlyDeferred(t *testing.T) {
	input := `-Queue ID- --Size-- ----Arrival Time---- --Sender/Recipient------
DEFER1    1024 Sun Jan 15 11:23:45  carol@example.com
                                        (deferred, connect timeout)
DEFER2     768 Sun Jan 15 11:23:45  eve@example.com
                                        (deferred, mailer-daemon error)`

	status := parseMailqOutput(input)
	if status.Total != 2 {
		t.Errorf("Total = %d, want 2", status.Total)
	}
	// No * markers, Total-Deferred = active
	if status.Active != 0 {
		t.Errorf("Active = %d, want 0", status.Active)
	}
	if status.Deferred != 2 {
		t.Errorf("Deferred = %d, want 2", status.Deferred)
	}
}

func TestFormatAge(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"under 1 min", 30 * time.Second, "<1m"},
		{"1 min", 1 * time.Minute, "1m"},
		{"59 min", 59 * time.Minute, "59m"},
		{"1 hour", 1 * time.Hour, "1h"},
		{"2 hours", 2 * time.Hour, "2h"},
		{"23 hours", 23 * time.Hour, "23h"},
		{"1 day", 24 * time.Hour, "1d"},
		{"7 days", 7 * 24 * time.Hour, "7d"},
		{"30 days", 30 * 24 * time.Hour, "30d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatAge(tt.duration)
			if got != tt.want {
				t.Errorf("formatAge(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestTryParseMailqDate(t *testing.T) {
	t.Run("full datetime format", func(t *testing.T) {
		result := tryParseMailqDate("Mon Jun 15 10:30:00")
		if result.IsZero() {
			t.Fatal("got zero time, want valid")
		}
		// Year depends on system clock; just verify non-zero
	})

	t.Run("short date only", func(t *testing.T) {
		result := tryParseMailqDate("Mon Jun 15")
		if result.IsZero() {
			t.Fatal("got zero time, want valid")
		}
	})

	t.Run("no leading day name", func(t *testing.T) {
		result := tryParseMailqDate("Jun 15 10:30:00")
		if result.IsZero() {
			t.Fatal("got zero time, want valid")
		}
	})

	t.Run("short no leading day", func(t *testing.T) {
		result := tryParseMailqDate("Jun 15")
		if result.IsZero() {
			t.Fatal("got zero time, want valid")
		}
	})

	t.Run("garbage input returns zero time", func(t *testing.T) {
		result := tryParseMailqDate("not-a-date")
		if !result.IsZero() {
			t.Errorf("got %v, want zero time", result)
		}
	})

	t.Run("empty string returns zero time", func(t *testing.T) {
		result := tryParseMailqDate("")
		if !result.IsZero() {
			t.Errorf("got %v, want zero time", result)
		}
	})
}

func TestGetQueueStatus_ExecError(t *testing.T) {
	executor := &mockCommandExecutor{
		err: fmt.Errorf("mailq command not found"),
	}
	_, err := GetQueueStatus(executor)
	if err == nil {
		t.Fatal("GetQueueStatus = nil, want error")
	}
	if !strings.Contains(err.Error(), "mailq failed") {
		t.Errorf("error = %q, want 'mailq failed'", err.Error())
	}
}

func TestGetQueueStatus_Success(t *testing.T) {
	executor := &mockCommandExecutor{
		output: "Mail queue is empty",
	}
	status, err := GetQueueStatus(executor)
	if err != nil {
		t.Fatalf("GetQueueStatus = _, %v, want nil", err)
	}
	if status.Total != 0 {
		t.Errorf("Total = %d, want 0", status.Total)
	}
}
