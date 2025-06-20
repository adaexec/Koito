// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: search.sql

package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const searchArtists = `-- name: SearchArtists :many
SELECT id, name, musicbrainz_id, image, score
FROM (
    SELECT
        a.id,
        a.name,
        a.musicbrainz_id,
        a.image,
        similarity(aa.alias, $1) AS score,
        ROW_NUMBER() OVER (PARTITION BY a.id ORDER BY similarity(aa.alias, $1) DESC) AS rn
    FROM artist_aliases aa
    JOIN artists_with_name a ON aa.artist_id = a.id
    WHERE similarity(aa.alias, $1) > 0.22
) ranked
WHERE rn = 1
ORDER BY score DESC
LIMIT $2
`

type SearchArtistsParams struct {
	Similarity string
	Limit      int32
}

type SearchArtistsRow struct {
	ID            int32
	Name          string
	MusicBrainzID *uuid.UUID
	Image         *uuid.UUID
	Score         float32
}

func (q *Queries) SearchArtists(ctx context.Context, arg SearchArtistsParams) ([]SearchArtistsRow, error) {
	rows, err := q.db.Query(ctx, searchArtists, arg.Similarity, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchArtistsRow
	for rows.Next() {
		var i SearchArtistsRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.MusicBrainzID,
			&i.Image,
			&i.Score,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const searchArtistsBySubstring = `-- name: SearchArtistsBySubstring :many
SELECT id, name, musicbrainz_id, image, score
FROM (
    SELECT
        a.id,
        a.name,
        a.musicbrainz_id,
        a.image,
        1.0 AS score, -- why
        ROW_NUMBER() OVER (PARTITION BY a.id ORDER BY aa.alias) AS rn
    FROM artist_aliases aa
    JOIN artists_with_name a ON aa.artist_id = a.id
    WHERE aa.alias ILIKE $1 || '%'
) ranked
WHERE rn = 1
ORDER BY score DESC
LIMIT $2
`

type SearchArtistsBySubstringParams struct {
	Column1 pgtype.Text
	Limit   int32
}

type SearchArtistsBySubstringRow struct {
	ID            int32
	Name          string
	MusicBrainzID *uuid.UUID
	Image         *uuid.UUID
	Score         float64
}

func (q *Queries) SearchArtistsBySubstring(ctx context.Context, arg SearchArtistsBySubstringParams) ([]SearchArtistsBySubstringRow, error) {
	rows, err := q.db.Query(ctx, searchArtistsBySubstring, arg.Column1, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchArtistsBySubstringRow
	for rows.Next() {
		var i SearchArtistsBySubstringRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.MusicBrainzID,
			&i.Image,
			&i.Score,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const searchReleases = `-- name: SearchReleases :many
SELECT
    ranked.id,
    ranked.title,
    ranked.musicbrainz_id,
    ranked.image,
    ranked.various_artists,
    ranked.score,
    get_artists_for_release(ranked.id) AS artists
FROM (
    SELECT
        r.id,
        r.title,
        r.musicbrainz_id,
        r.image,
        r.various_artists,
        similarity(ra.alias, $1) AS score,
        ROW_NUMBER() OVER (PARTITION BY r.id ORDER BY similarity(ra.alias, $1) DESC) AS rn
    FROM release_aliases ra
    JOIN releases_with_title r ON ra.release_id = r.id
    WHERE similarity(ra.alias, $1) > 0.22
) ranked
WHERE rn = 1
ORDER BY score DESC, title
LIMIT $2
`

type SearchReleasesParams struct {
	Similarity string
	Limit      int32
}

type SearchReleasesRow struct {
	ID             int32
	Title          string
	MusicBrainzID  *uuid.UUID
	Image          *uuid.UUID
	VariousArtists bool
	Score          float32
	Artists        []byte
}

func (q *Queries) SearchReleases(ctx context.Context, arg SearchReleasesParams) ([]SearchReleasesRow, error) {
	rows, err := q.db.Query(ctx, searchReleases, arg.Similarity, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchReleasesRow
	for rows.Next() {
		var i SearchReleasesRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.MusicBrainzID,
			&i.Image,
			&i.VariousArtists,
			&i.Score,
			&i.Artists,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const searchReleasesBySubstring = `-- name: SearchReleasesBySubstring :many
SELECT
    ranked.id,
    ranked.title,
    ranked.musicbrainz_id,
    ranked.image,
    ranked.various_artists,
    ranked.score,
    get_artists_for_release(ranked.id) AS artists
FROM (
    SELECT
        r.id,
        r.title,
        r.musicbrainz_id,
        r.image,
        r.various_artists,
        1.0 AS score, -- idk why
        ROW_NUMBER() OVER (PARTITION BY r.id ORDER BY ra.alias) AS rn
    FROM release_aliases ra
    JOIN releases_with_title r ON ra.release_id = r.id
    WHERE ra.alias ILIKE $1 || '%'
) ranked
WHERE rn = 1
ORDER BY score DESC, title
LIMIT $2
`

type SearchReleasesBySubstringParams struct {
	Column1 pgtype.Text
	Limit   int32
}

type SearchReleasesBySubstringRow struct {
	ID             int32
	Title          string
	MusicBrainzID  *uuid.UUID
	Image          *uuid.UUID
	VariousArtists bool
	Score          float64
	Artists        []byte
}

func (q *Queries) SearchReleasesBySubstring(ctx context.Context, arg SearchReleasesBySubstringParams) ([]SearchReleasesBySubstringRow, error) {
	rows, err := q.db.Query(ctx, searchReleasesBySubstring, arg.Column1, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchReleasesBySubstringRow
	for rows.Next() {
		var i SearchReleasesBySubstringRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.MusicBrainzID,
			&i.Image,
			&i.VariousArtists,
			&i.Score,
			&i.Artists,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const searchTracks = `-- name: SearchTracks :many
SELECT
    ranked.id,
    ranked.title,
    ranked.musicbrainz_id,
    ranked.release_id,
    ranked.image,
    ranked.score,
    get_artists_for_track(ranked.id) AS artists
FROM (
    SELECT
        t.id,
        t.title,
        t.musicbrainz_id,
        t.release_id,
        r.image,
        similarity(ta.alias, $1) AS score,
        ROW_NUMBER() OVER (PARTITION BY t.id ORDER BY similarity(ta.alias, $1) DESC) AS rn
    FROM track_aliases ta
    JOIN tracks_with_title t ON ta.track_id = t.id
    JOIN releases r ON t.release_id = r.id
    WHERE similarity(ta.alias, $1) > 0.22
) ranked
WHERE rn = 1
ORDER BY score DESC, title
LIMIT $2
`

type SearchTracksParams struct {
	Similarity string
	Limit      int32
}

type SearchTracksRow struct {
	ID            int32
	Title         string
	MusicBrainzID *uuid.UUID
	ReleaseID     int32
	Image         *uuid.UUID
	Score         float32
	Artists       []byte
}

func (q *Queries) SearchTracks(ctx context.Context, arg SearchTracksParams) ([]SearchTracksRow, error) {
	rows, err := q.db.Query(ctx, searchTracks, arg.Similarity, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchTracksRow
	for rows.Next() {
		var i SearchTracksRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.MusicBrainzID,
			&i.ReleaseID,
			&i.Image,
			&i.Score,
			&i.Artists,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const searchTracksBySubstring = `-- name: SearchTracksBySubstring :many
SELECT
    ranked.id,
    ranked.title,
    ranked.musicbrainz_id,
    ranked.release_id,
    ranked.image,
    ranked.score,
    get_artists_for_track(ranked.id) AS artists
FROM (
    SELECT
        t.id,
        t.title,
        t.musicbrainz_id,
        t.release_id,
        r.image,
        1.0 AS score,
        ROW_NUMBER() OVER (PARTITION BY t.id ORDER BY ta.alias) AS rn
    FROM track_aliases ta
    JOIN tracks_with_title t ON ta.track_id = t.id
    JOIN releases r ON t.release_id = r.id
    WHERE ta.alias ILIKE $1 || '%'
) ranked
WHERE rn = 1
ORDER BY score DESC, title
LIMIT $2
`

type SearchTracksBySubstringParams struct {
	Column1 pgtype.Text
	Limit   int32
}

type SearchTracksBySubstringRow struct {
	ID            int32
	Title         string
	MusicBrainzID *uuid.UUID
	ReleaseID     int32
	Image         *uuid.UUID
	Score         float64
	Artists       []byte
}

func (q *Queries) SearchTracksBySubstring(ctx context.Context, arg SearchTracksBySubstringParams) ([]SearchTracksBySubstringRow, error) {
	rows, err := q.db.Query(ctx, searchTracksBySubstring, arg.Column1, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchTracksBySubstringRow
	for rows.Next() {
		var i SearchTracksBySubstringRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.MusicBrainzID,
			&i.ReleaseID,
			&i.Image,
			&i.Score,
			&i.Artists,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
