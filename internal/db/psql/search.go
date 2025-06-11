package psql

import (
	"context"
	"encoding/json"

	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

const searchItemLimit = 5
const substringSearchLength = 6

func (d *Psql) SearchArtists(ctx context.Context, q string) ([]*models.Artist, error) {
	if len(q) < substringSearchLength {
		rows, err := d.q.SearchArtistsBySubstring(ctx, repository.SearchArtistsBySubstringParams{
			Column1: pgtype.Text{String: q, Valid: true},
			Limit:   searchItemLimit,
		})
		if err != nil {
			return nil, err
		}
		ret := make([]*models.Artist, len(rows))
		for i, row := range rows {
			ret[i] = &models.Artist{
				ID:    row.ID,
				MbzID: row.MusicBrainzID,
				Name:  row.Name,
				Image: row.Image,
			}
		}
		return ret, nil
	} else {
		rows, err := d.q.SearchArtists(ctx, repository.SearchArtistsParams{
			Similarity: q,
			Limit:      searchItemLimit,
		})
		if err != nil {
			return nil, err
		}
		ret := make([]*models.Artist, len(rows))
		for i, row := range rows {
			ret[i] = &models.Artist{
				ID:    row.ID,
				MbzID: row.MusicBrainzID,
				Name:  row.Name,
				Image: row.Image,
			}
		}
		return ret, nil
	}
}

func (d *Psql) SearchAlbums(ctx context.Context, q string) ([]*models.Album, error) {
	if len(q) < substringSearchLength {
		rows, err := d.q.SearchReleasesBySubstring(ctx, repository.SearchReleasesBySubstringParams{
			Column1: pgtype.Text{String: q, Valid: true},
			Limit:   searchItemLimit,
		})
		if err != nil {
			return nil, err
		}
		ret := make([]*models.Album, len(rows))
		for i, row := range rows {
			ret[i] = &models.Album{
				ID:             row.ID,
				MbzID:          row.MusicBrainzID,
				Title:          row.Title,
				VariousArtists: row.VariousArtists,
				Image:          row.Image,
			}
			err = json.Unmarshal(row.Artists, &ret[i].Artists)
			if err != nil {
				return nil, err
			}
		}
		return ret, nil
	} else {
		rows, err := d.q.SearchReleases(ctx, repository.SearchReleasesParams{
			Similarity: q,
			Limit:      searchItemLimit,
		})
		if err != nil {
			return nil, err
		}
		ret := make([]*models.Album, len(rows))
		for i, row := range rows {
			ret[i] = &models.Album{
				ID:             row.ID,
				MbzID:          row.MusicBrainzID,
				Title:          row.Title,
				VariousArtists: row.VariousArtists,
				Image:          row.Image,
			}
			err = json.Unmarshal(row.Artists, &ret[i].Artists)
			if err != nil {
				return nil, err
			}
		}
		return ret, nil
	}
}

func (d *Psql) SearchTracks(ctx context.Context, q string) ([]*models.Track, error) {
	if len(q) < substringSearchLength {
		rows, err := d.q.SearchTracksBySubstring(ctx, repository.SearchTracksBySubstringParams{
			Column1: pgtype.Text{String: q, Valid: true},
			Limit:   searchItemLimit,
		})
		if err != nil {
			return nil, err
		}
		ret := make([]*models.Track, len(rows))
		for i, row := range rows {
			ret[i] = &models.Track{
				ID:    row.ID,
				MbzID: row.MusicBrainzID,
				Title: row.Title,
				Image: row.Image,
			}
			err = json.Unmarshal(row.Artists, &ret[i].Artists)
			if err != nil {
				return nil, err
			}
		}
		return ret, nil
	} else {
		rows, err := d.q.SearchTracks(ctx, repository.SearchTracksParams{
			Similarity: q,
			Limit:      searchItemLimit,
		})
		if err != nil {
			return nil, err
		}
		ret := make([]*models.Track, len(rows))
		for i, row := range rows {
			ret[i] = &models.Track{
				ID:    row.ID,
				MbzID: row.MusicBrainzID,
				Title: row.Title,
				Image: row.Image,
			}
			err = json.Unmarshal(row.Artists, &ret[i].Artists)
			if err != nil {
				return nil, err
			}
		}
		return ret, nil
	}
}
