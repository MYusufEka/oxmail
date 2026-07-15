package domain_test

import (
	"context"
	"testing"

	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatsService_IncrementSent(t *testing.T) {
	db := setupTestDB(t)
	svc := domain.NewStatsService(db.Conn)
	ctx := context.Background()

	require.NoError(t, svc.IncrementSent(ctx))
	require.NoError(t, svc.IncrementSent(ctx))
	require.NoError(t, svc.IncrementSent(ctx))

	stats, err := svc.GetStats(ctx, 1)
	require.NoError(t, err)
	require.Len(t, stats, 1)
	assert.Equal(t, int64(3), stats[0].Sent)
}

func TestStatsService_IncrementReceived(t *testing.T) {
	db := setupTestDB(t)
	svc := domain.NewStatsService(db.Conn)
	ctx := context.Background()

	require.NoError(t, svc.IncrementReceived(ctx))
	require.NoError(t, svc.IncrementReceived(ctx))

	stats, err := svc.GetStats(ctx, 1)
	require.NoError(t, err)
	require.Len(t, stats, 1)
	assert.Equal(t, int64(2), stats[0].Received)
}

func TestStatsService_IncrementBounced(t *testing.T) {
	db := setupTestDB(t)
	svc := domain.NewStatsService(db.Conn)
	ctx := context.Background()

	require.NoError(t, svc.IncrementBounced(ctx))

	stats, err := svc.GetStats(ctx, 1)
	require.NoError(t, err)
	require.Len(t, stats, 1)
	assert.Equal(t, int64(1), stats[0].Bounced)
}

func TestStatsService_IncrementSpamCaught(t *testing.T) {
	db := setupTestDB(t)
	svc := domain.NewStatsService(db.Conn)
	ctx := context.Background()

	require.NoError(t, svc.IncrementSpamCaught(ctx))

	stats, err := svc.GetStats(ctx, 1)
	require.NoError(t, err)
	require.Len(t, stats, 1)
	assert.Equal(t, int64(1), stats[0].SpamCaught)
}

func TestStatsService_GetStats_Empty(t *testing.T) {
	db := setupTestDB(t)
	svc := domain.NewStatsService(db.Conn)
	ctx := context.Background()

	stats, err := svc.GetStats(ctx, 7)
	require.NoError(t, err)
	assert.Empty(t, stats)
}

func TestStatsService_GetStats_MultipleCounters(t *testing.T) {
	db := setupTestDB(t)
	svc := domain.NewStatsService(db.Conn)
	ctx := context.Background()

	for range 5 {
		require.NoError(t, svc.IncrementSent(ctx))
	}
	for range 3 {
		require.NoError(t, svc.IncrementReceived(ctx))
	}
	require.NoError(t, svc.IncrementBounced(ctx))
	require.NoError(t, svc.IncrementSpamCaught(ctx))
	require.NoError(t, svc.IncrementSpamCaught(ctx))

	stats, err := svc.GetStats(ctx, 7)
	require.NoError(t, err)
	require.Len(t, stats, 1)
	assert.Equal(t, int64(5), stats[0].Sent)
	assert.Equal(t, int64(3), stats[0].Received)
	assert.Equal(t, int64(1), stats[0].Bounced)
	assert.Equal(t, int64(2), stats[0].SpamCaught)
}
