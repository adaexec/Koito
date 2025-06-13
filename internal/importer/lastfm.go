package importer

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/gabehf/koito/internal/catalog"
	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/mbz"
	"github.com/google/uuid"
)

type LastFMExportPage struct {
	Track []LastFMTrack `json:"track"`
}
type LastFMTrack struct {
	Artist LastFMItem    `json:"artist"`
	Images []LastFMImage `json:"image"`
	MBID   string        `json:"mbid"`
	Album  LastFMItem    `json:"album"`
	Name   string        `json:"name"`
	Date   LastFMDate    `json:"date"`
}
type LastFMItem struct {
	MBID string `json:"mbid"`
	Text string `json:"#text"`
}
type LastFMDate struct {
	Unix string `json:"uts"`
	Text string `json:"#text"`
}
type LastFMImage struct {
	Size string `json:"size"`
	Url  string `json:"#text"`
}

func ImportLastFMFile(ctx context.Context, store db.DB, mbzc mbz.MusicBrainzCaller, filename string) error {
	l := logger.FromContext(ctx)
	l.Info().Msgf("Beginning LastFM import on file: %s", filename)
	file, err := os.Open(path.Join(cfg.ConfigDir(), "import", filename))
	if err != nil {
		l.Err(err).Msgf("Failed to read import file: %s", filename)
		return err
	}
	var throttleFunc = func() {}
	if ms := cfg.ThrottleImportMs(); ms > 0 {
		throttleFunc = func() {
			time.Sleep(time.Duration(ms) * time.Millisecond)
		}
	}
	export := make([]LastFMExportPage, 0)
	err = json.NewDecoder(file).Decode(&export)
	if err != nil {
		return err
	}
	count := 0
	for _, item := range export {
		for _, track := range item.Track {
			album := track.Album.Text
			if album == "" {
				album = track.Name
			}
			if track.Name == "" || track.Artist.Text == "" {
				l.Debug().Msg("Skipping invalid LastFM import item")
				continue
			}
			albumMbzID, err := uuid.Parse(track.Album.MBID)
			if err != nil {
				albumMbzID = uuid.Nil
			}
			artistMbzID, err := uuid.Parse(track.Artist.MBID)
			if err != nil {
				artistMbzID = uuid.Nil
			}
			trackMbzID, err := uuid.Parse(track.MBID)
			if err != nil {
				trackMbzID = uuid.Nil
			}
			var ts time.Time
			unix, err := strconv.ParseInt(track.Date.Unix, 10, 64)
			if err != nil {
				ts, err = time.Parse("02 Jan 2006, 15:04", track.Date.Text)
				if err != nil {
					ts = time.Now()
				}
			} else {
				ts = time.Unix(unix, 0).UTC()
			}
			opts := catalog.SubmitListenOpts{
				MbzCaller:      mbzc,
				Artist:         track.Artist.Text,
				ArtistMbzIDs:   []uuid.UUID{artistMbzID},
				TrackTitle:     track.Name,
				RecordingMbzID: trackMbzID,
				ReleaseTitle:   album,
				ReleaseMbzID:   albumMbzID,
				Time:           ts,
				UserID:         1,
			}
			err = catalog.SubmitListen(ctx, store, opts)
			if err != nil {
				l.Err(err).Msg("Failed to import LastFM playback item")
				return err
			}
			count++
			throttleFunc()
		}
	}
	return finishImport(ctx, filename, count)
}
