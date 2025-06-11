package psql_test

import (
	"context"
	"testing"

	"github.com/gabehf/koito/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTopAlbumsPaginated(t *testing.T) {
	testDataForTopItems(t)
	ctx := context.Background()

	// Test valid
	resp, err := store.GetTopAlbumsPaginated(ctx, db.GetItemsOpts{Period: db.PeriodAllTime})
	require.NoError(t, err)
	require.Len(t, resp.Items, 4)
	assert.Equal(t, int64(4), resp.TotalCount)
	assert.Equal(t, "Release One", resp.Items[0].Title)
	assert.Equal(t, "Release Two", resp.Items[1].Title)
	assert.Equal(t, "Release Three", resp.Items[2].Title)
	assert.Equal(t, "Release Four", resp.Items[3].Title)

	// Test pagination
	resp, err = store.GetTopAlbumsPaginated(ctx, db.GetItemsOpts{Limit: 1, Page: 2, Period: db.PeriodAllTime})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, "Release Two", resp.Items[0].Title)

	// Test page out of range
	resp, err = store.GetTopAlbumsPaginated(ctx, db.GetItemsOpts{Limit: 1, Page: 10, Period: db.PeriodAllTime})
	require.NoError(t, err)
	require.Empty(t, resp.Items)
	assert.False(t, resp.HasNextPage)

	// Test invalid inputs
	_, err = store.GetTopAlbumsPaginated(ctx, db.GetItemsOpts{Limit: -1, Page: 0})
	assert.Error(t, err)

	_, err = store.GetTopAlbumsPaginated(ctx, db.GetItemsOpts{Limit: 1, Page: -1})
	assert.Error(t, err)

	// Test specify period
	resp, err = store.GetTopAlbumsPaginated(ctx, db.GetItemsOpts{Period: db.PeriodDay})
	require.NoError(t, err)
	require.Len(t, resp.Items, 0) // empty
	assert.Equal(t, int64(0), resp.TotalCount)
	// should default to PeriodDay
	resp, err = store.GetTopAlbumsPaginated(ctx, db.GetItemsOpts{})
	require.NoError(t, err)
	require.Len(t, resp.Items, 0) // empty
	assert.Equal(t, int64(0), resp.TotalCount)

	resp, err = store.GetTopAlbumsPaginated(ctx, db.GetItemsOpts{Period: db.PeriodWeek})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, int64(1), resp.TotalCount)
	assert.Equal(t, "Release Four", resp.Items[0].Title)

	resp, err = store.GetTopAlbumsPaginated(ctx, db.GetItemsOpts{Period: db.PeriodMonth})
	require.NoError(t, err)
	require.Len(t, resp.Items, 2)
	assert.Equal(t, int64(2), resp.TotalCount)
	assert.Equal(t, "Release Three", resp.Items[0].Title)
	assert.Equal(t, "Release Four", resp.Items[1].Title)

	resp, err = store.GetTopAlbumsPaginated(ctx, db.GetItemsOpts{Period: db.PeriodYear})
	require.NoError(t, err)
	require.Len(t, resp.Items, 3)
	assert.Equal(t, int64(3), resp.TotalCount)
	assert.Equal(t, "Release Two", resp.Items[0].Title)
	assert.Equal(t, "Release Three", resp.Items[1].Title)
	assert.Equal(t, "Release Four", resp.Items[2].Title)

	// test specific artist
	resp, err = store.GetTopAlbumsPaginated(ctx, db.GetItemsOpts{Period: db.PeriodYear, ArtistID: 2})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, int64(1), resp.TotalCount)
	assert.Equal(t, "Release Two", resp.Items[0].Title)

	// Test specify dates

	testDataAbsoluteListenTimes(t)

	resp, err = store.GetTopAlbumsPaginated(ctx, db.GetItemsOpts{Year: 2023})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, int64(1), resp.TotalCount)
	assert.Equal(t, "Release One", resp.Items[0].Title)

	resp, err = store.GetTopAlbumsPaginated(ctx, db.GetItemsOpts{Month: 6, Year: 2024})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, int64(1), resp.TotalCount)
	assert.Equal(t, "Release Two", resp.Items[0].Title)

	// invalid, year required with month
	_, err = store.GetTopAlbumsPaginated(ctx, db.GetItemsOpts{Month: 10})
	require.Error(t, err)
}
