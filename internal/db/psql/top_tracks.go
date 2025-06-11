package psql

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/repository"
	"github.com/gabehf/koito/internal/utils"
)

func (d *Psql) GetTopTracksPaginated(ctx context.Context, opts db.GetItemsOpts) (*db.PaginatedResponse[*models.Track], error) {
	l := logger.FromContext(ctx)
	offset := (opts.Page - 1) * opts.Limit
	t1, t2, err := utils.DateRange(opts.Week, opts.Month, opts.Year)
	if err != nil {
		return nil, err
	}
	if opts.Month == 0 && opts.Year == 0 {
		// use period, not date range
		t2 = time.Now()
		t1 = db.StartTimeFromPeriod(opts.Period)
	}
	if opts.Limit == 0 {
		opts.Limit = DefaultItemsPerPage
	}
	var tracks []*models.Track
	var count int64
	if opts.AlbumID > 0 {
		l.Debug().Msgf("Fetching top %d tracks with period %s on page %d from range %v to %v",
			opts.Limit, opts.Period, opts.Page, t1.Format("Jan 02, 2006"), t2.Format("Jan 02, 2006"))
		rows, err := d.q.GetTopTracksInReleasePaginated(ctx, repository.GetTopTracksInReleasePaginatedParams{
			ListenedAt:   t1,
			ListenedAt_2: t2,
			Limit:        int32(opts.Limit),
			Offset:       int32(offset),
			ReleaseID:    int32(opts.AlbumID),
		})
		if err != nil {
			return nil, err
		}
		tracks = make([]*models.Track, len(rows))
		for i, row := range rows {
			artists := make([]models.SimpleArtist, 0)
			err = json.Unmarshal(row.Artists, &artists)
			if err != nil {
				l.Err(err).Msgf("Error unmarshalling artists for track with id %d", row.ID)
				artists = nil
			}
			t := &models.Track{
				Title:       row.Title,
				MbzID:       row.MusicBrainzID,
				ID:          row.ID,
				ListenCount: row.ListenCount,
				Image:       row.Image,
				AlbumID:     row.ReleaseID,
				Artists:     artists,
			}
			tracks[i] = t
		}
		count, err = d.q.CountTopTracksByRelease(ctx, repository.CountTopTracksByReleaseParams{
			ListenedAt:   t1,
			ListenedAt_2: t2,
			ReleaseID:    int32(opts.AlbumID),
		})
		if err != nil {
			return nil, err
		}
	} else if opts.ArtistID > 0 {
		l.Debug().Msgf("Fetching top %d tracks with period %s on page %d from range %v to %v",
			opts.Limit, opts.Period, opts.Page, t1.Format("Jan 02, 2006"), t2.Format("Jan 02, 2006"))
		rows, err := d.q.GetTopTracksByArtistPaginated(ctx, repository.GetTopTracksByArtistPaginatedParams{
			ListenedAt:   t1,
			ListenedAt_2: t2,
			Limit:        int32(opts.Limit),
			Offset:       int32(offset),
			ArtistID:     int32(opts.ArtistID),
		})
		if err != nil {
			return nil, err
		}
		tracks = make([]*models.Track, len(rows))
		for i, row := range rows {
			artists := make([]models.SimpleArtist, 0)
			err = json.Unmarshal(row.Artists, &artists)
			if err != nil {
				l.Err(err).Msgf("Error unmarshalling artists for track with id %d", row.ID)
				artists = nil
			}
			t := &models.Track{
				Title:       row.Title,
				MbzID:       row.MusicBrainzID,
				ID:          row.ID,
				Image:       row.Image,
				ListenCount: row.ListenCount,
				AlbumID:     row.ReleaseID,
				Artists:     artists,
			}
			tracks[i] = t
		}
		count, err = d.q.CountTopTracksByArtist(ctx, repository.CountTopTracksByArtistParams{
			ListenedAt:   t1,
			ListenedAt_2: t2,
			ArtistID:     int32(opts.ArtistID),
		})
		if err != nil {
			return nil, err
		}
	} else {
		l.Debug().Msgf("Fetching top %d tracks with period %s on page %d from range %v to %v",
			opts.Limit, opts.Period, opts.Page, t1.Format("Jan 02, 2006"), t2.Format("Jan 02, 2006"))
		rows, err := d.q.GetTopTracksPaginated(ctx, repository.GetTopTracksPaginatedParams{
			ListenedAt:   t1,
			ListenedAt_2: t2,
			Limit:        int32(opts.Limit),
			Offset:       int32(offset),
		})
		if err != nil {
			return nil, err
		}
		tracks = make([]*models.Track, len(rows))
		for i, row := range rows {
			artists := make([]models.SimpleArtist, 0)
			err = json.Unmarshal(row.Artists, &artists)
			if err != nil {
				l.Err(err).Msgf("Error unmarshalling artists for track with id %d", row.ID)
				artists = nil
			}
			t := &models.Track{
				Title:       row.Title,
				MbzID:       row.MusicBrainzID,
				ID:          row.ID,
				Image:       row.Image,
				ListenCount: row.ListenCount,
				AlbumID:     row.ReleaseID,
				Artists:     artists,
			}
			tracks[i] = t
		}
		count, err = d.q.CountTopTracks(ctx, repository.CountTopTracksParams{
			ListenedAt:   t1,
			ListenedAt_2: t2,
		})
		if err != nil {
			return nil, err
		}
		l.Debug().Msgf("Database responded with %d tracks out of a total %d", len(rows), count)
	}

	return &db.PaginatedResponse[*models.Track]{
		Items:        tracks,
		TotalCount:   count,
		ItemsPerPage: int32(opts.Limit),
		HasNextPage:  int64(offset+len(tracks)) < count,
		CurrentPage:  int32(opts.Page),
	}, nil
}
