package importer

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gabehf/koito/internal/catalog"
	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/mbz"
	"github.com/gabehf/koito/internal/utils"
)

type MalojaExport struct {
	Scrobbles []MalojaExportItem `json:"scrobbles"`
}
type MalojaExportItem struct {
	Time  int64       `json:"time"`
	Track MalojaTrack `json:"track"`
}
type MalojaTrack struct {
	Artists []string `json:"artists"`
	Title   string   `json:"title"`
	Album   struct {
		Title string `json:"albumtitle"`
	} `json:"album"`
}

func ImportMalojaFile(ctx context.Context, store db.DB, filename string) error {
	l := logger.FromContext(ctx)
	l.Info().Msgf("Beginning maloja import on file: %s", filename)
	file, err := os.Open(path.Join(cfg.ConfigDir(), "import", filename))
	if err != nil {
		l.Err(err).Msgf("Failed to read import file: %s", filename)
		return err
	}
	defer file.Close()
	var throttleFunc = func() {}
	if ms := cfg.ThrottleImportMs(); ms > 0 {
		throttleFunc = func() {
			time.Sleep(time.Duration(ms) * time.Millisecond)
		}
	}
	export := new(MalojaExport)
	err = json.NewDecoder(file).Decode(&export)
	if err != nil {
		return err
	}
	for _, item := range export.Scrobbles {
		martists := make([]string, 0)
		// Maloja has a tendency to have the the artist order ['feature', 'main \u2022 feature'], so
		// here we try to turn that artist array into ['main', 'feature']
		item.Track.Artists = utils.MoveFirstMatchToFront(item.Track.Artists, " \u2022 ")
		for _, an := range item.Track.Artists {
			ans := strings.Split(an, " \u2022 ")
			martists = append(martists, ans...)
		}
		artists := utils.UniqueIgnoringCase(martists)
		if len(item.Track.Artists) < 1 || item.Track.Title == "" {
			l.Debug().Msg("Skipping invalid maloja import item")
			continue
		}
		ts := time.Unix(item.Time, 0)
		if !inImportTimeWindow(ts) {
			l.Debug().Msgf("Skipping import due to import time rules")
			continue
		}
		opts := catalog.SubmitListenOpts{
			MbzCaller:    &mbz.MusicBrainzClient{},
			Artist:       item.Track.Artists[0],
			ArtistNames:  artists,
			TrackTitle:   item.Track.Title,
			ReleaseTitle: item.Track.Album.Title,
			Time:         ts.Local(),
			Client:       "maloja",
			UserID:       1,
		}
		err = catalog.SubmitListen(ctx, store, opts)
		if err != nil {
			l.Err(err).Msg("Failed to import maloja playback item")
			return err
		}
		throttleFunc()
	}
	return finishImport(ctx, filename, len(export.Scrobbles))
}
