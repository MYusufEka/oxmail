package domain_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/MYusufEka/oxmail/internal/database"
	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupBounceTestDB(t *testing.T) *database.DB {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	db, err := database.Open(":memory:", logger)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestBounceService_RecordBounce(t *testing.T) {
	testDB := setupBounceTestDB(t)
	svc := domain.NewBounceService(testDB.Conn)
	ctx := context.Background()

	tests := []struct {
		name         string
		recipient    string
		sender       string
		subject      string
		bounceType   string
		errorMessage string
		wantErr      bool
	}{
		{
			name:         "hard bounce 5xx",
			recipient:    "bad@example.com",
			sender:       "alice@local.test",
			subject:      "Hello",
			bounceType:   "hard",
			errorMessage: "550 5.1.1 user unknown",
		},
		{
			name:         "soft bounce 4xx",
			recipient:    "busy@example.com",
			sender:       "alice@local.test",
			subject:      "Hi",
			bounceType:   "soft",
			errorMessage: "421 4.2.1 mailbox full",
		},
		{
			name:         "empty recipient fails",
			recipient:    "",
			sender:       "alice@local.test",
			bounceType:   "hard",
			errorMessage: "550 user unknown",
			wantErr:      true,
		},
		{
			name:         "invalid bounce type fails",
			recipient:    "x@example.com",
			sender:       "alice@local.test",
			bounceType:   "unknown",
			errorMessage: "some error",
			wantErr:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bounce, err := svc.RecordBounce(ctx, tc.recipient, tc.sender, tc.subject, tc.bounceType, tc.errorMessage)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, bounce)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, bounce)
			assert.Greater(t, bounce.ID, int64(0))
			assert.Equal(t, tc.recipient, bounce.Recipient)
			assert.Equal(t, tc.sender, bounce.Sender)
			assert.Equal(t, tc.bounceType, bounce.BounceType)
			assert.Equal(t, tc.errorMessage, bounce.ErrorMessage)
			assert.WithinDuration(t, time.Now(), bounce.BouncedAt, 5*time.Second)
		})
	}
}

func TestBounceService_ListBounces(t *testing.T) {
	testDB := setupBounceTestDB(t)
	svc := domain.NewBounceService(testDB.Conn)
	ctx := context.Background()

	_, err := svc.RecordBounce(ctx, "hard@example.com", "alice@local.test", "Subj1", "hard", "550 user unknown")
	require.NoError(t, err)
	_, err = svc.RecordBounce(ctx, "soft@example.com", "alice@local.test", "Subj2", "soft", "421 mailbox full")
	require.NoError(t, err)
	_, err = svc.RecordBounce(ctx, "hard@example.com", "bob@local.test", "Subj3", "hard", "550 user unknown")
	require.NoError(t, err)

	t.Run("list all", func(t *testing.T) {
		bounces, err := svc.ListBounces(ctx, domain.BounceFilter{})
		require.NoError(t, err)
		assert.Len(t, bounces, 3)
	})

	t.Run("filter by recipient", func(t *testing.T) {
		bounces, err := svc.ListBounces(ctx, domain.BounceFilter{Recipient: "hard@example.com"})
		require.NoError(t, err)
		assert.Len(t, bounces, 2)
		for _, b := range bounces {
			assert.Equal(t, "hard@example.com", b.Recipient)
		}
	})

	t.Run("filter by bounce type soft", func(t *testing.T) {
		bounces, err := svc.ListBounces(ctx, domain.BounceFilter{BounceType: "soft"})
		require.NoError(t, err)
		assert.Len(t, bounces, 1)
		assert.Equal(t, "soft", bounces[0].BounceType)
	})

	t.Run("limit", func(t *testing.T) {
		bounces, err := svc.ListBounces(ctx, domain.BounceFilter{Limit: 2})
		require.NoError(t, err)
		assert.Len(t, bounces, 2)
	})

	t.Run("empty result", func(t *testing.T) {
		bounces, err := svc.ListBounces(ctx, domain.BounceFilter{Recipient: "nobody@example.com"})
		require.NoError(t, err)
		assert.Empty(t, bounces)
	})
}
