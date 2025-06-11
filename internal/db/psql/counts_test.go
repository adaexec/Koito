package psql_test

import (
	"context"
	"testing"

	"github.com/gabehf/koito/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCountListens(t *testing.T) {
	ctx := context.Background()
	testDataForTopItems(t)

	// Test CountListens
	period := db.PeriodWeek
	count, err := store.CountListens(ctx, period)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count, "expected listens count to match inserted data")

	truncateTestData(t)
}

func TestCountTracks(t *testing.T) {
	ctx := context.Background()
	testDataForTopItems(t)

	// Test CountTracks
	period := db.PeriodMonth
	count, err := store.CountTracks(ctx, period)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count, "expected tracks count to match inserted data")

	truncateTestData(t)
}

func TestCountAlbums(t *testing.T) {
	ctx := context.Background()
	testDataForTopItems(t)

	// Test CountAlbums
	period := db.PeriodYear
	count, err := store.CountAlbums(ctx, period)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count, "expected albums count to match inserted data")

	truncateTestData(t)
}

func TestCountArtists(t *testing.T) {
	ctx := context.Background()
	testDataForTopItems(t)

	// Test CountArtists
	period := db.PeriodAllTime
	count, err := store.CountArtists(ctx, period)
	require.NoError(t, err)
	assert.Equal(t, int64(4), count, "expected artists count to match inserted data")

	truncateTestData(t)
}

func TestCountTimeListened(t *testing.T) {
	ctx := context.Background()
	testDataForTopItems(t)

	// Test CountTimeListened
	period := db.PeriodMonth
	count, err := store.CountTimeListened(ctx, period)
	require.NoError(t, err)
	// 3 listens in past month, each 100 seconds
	assert.Equal(t, int64(300), count, "expected total time listened to match inserted data")

	truncateTestData(t)
}
