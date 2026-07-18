package domain

import (
	"testing"
	"time"
)

// TestModelTypes verify models can be created with expected zero/zero-ish values.
// These are minimal — models.go is mostly type definitions without behavior.

func TestDomainModelDefaults(t *testing.T) {
	d := Domain{}
	if d.ID != 0 {
		t.Errorf("default Domain.ID = %d, want 0", d.ID)
	}
	if d.Active {
		t.Error("default Domain.Active should be false")
	}
}

func TestUserModelDefaults(t *testing.T) {
	u := User{}
	if u.ID != 0 {
		t.Errorf("default User.ID = %d, want 0", u.ID)
	}
	if u.Quota != 0 {
		t.Errorf("default User.Quota = %d, want 0", u.Quota)
	}
}

func TestMailMessageModel(t *testing.T) {
	m := MailMessage{ReceivedAt: time.Now()}
	if m.ReceivedAt.IsZero() {
		t.Error("MailMessage.ReceivedAt should not be zero after setting")
	}
	if m.To != nil {
		t.Error("MailMessage.To should be nil slice (omitempty in JSON) at zero value")
	}
}

func TestErrorResponseModel(t *testing.T) {
	e := ErrorResponse{
		Error: ErrorDetail{
			Code:    "INVALID_INPUT",
			Message: "Bad request",
		},
	}
	if e.Error.Code != "INVALID_INPUT" {
		t.Errorf("ErrorDetail.Code = %q, want INVALID_INPUT", e.Error.Code)
	}
	if e.Error.Message != "Bad request" {
		t.Errorf("ErrorDetail.Message = %q, want Bad request", e.Error.Message)
	}
}
