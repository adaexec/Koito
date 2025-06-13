package importer

import (
	"archive/zip"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path"
	"strings"
	"time"

	"github.com/gabehf/koito/engine/handlers"
	"github.com/gabehf/koito/internal/catalog"
	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/mbz"
	"github.com/gabehf/koito/internal/utils"
	"github.com/google/uuid"
)

func ImportListenBrainzExport(ctx context.Context, store db.DB, mbzc mbz.MusicBrainzCaller, filename string) error {
	l := logger.FromContext(ctx)

	r, err := zip.OpenReader(path.Join(path.Join(cfg.ConfigDir(), "import", filename)))
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {

		if f.FileInfo().IsDir() {
			continue
		}

		if strings.HasPrefix(f.Name, "listens/") && strings.HasSuffix(f.Name, ".jsonl") {
			fmt.Println("Found:", f.Name)

			rc, err := f.Open()
			if err != nil {
				log.Printf("Failed to open %s: %v\n", f.Name, err)
				continue
			}

			err = ImportListenBrainzFile(ctx, store, mbzc, rc, f.Name)
			if err != nil {
				l.Err(err).Msgf("Failed to import listens from file: %s", f.Name)
			}

			rc.Close()
		}
	}
	return finishImport(ctx, filename, 0)
}

func ImportListenBrainzFile(ctx context.Context, store db.DB, mbzc mbz.MusicBrainzCaller, r io.Reader, filename string) error {
	l := logger.FromContext(ctx)
	l.Info().Msgf("Beginning ListenBrainz import on file: %s", filename)

	scanner := bufio.NewScanner(r)

	var throttleFunc = func() {}
	if ms := cfg.ThrottleImportMs(); ms > 0 {
		throttleFunc = func() {
			time.Sleep(time.Duration(ms) * time.Millisecond)
		}
	}
	count := 0
	for scanner.Scan() {
		line := scanner.Bytes()
		payload := new(handlers.LbzSubmitListenPayload)
		err := json.Unmarshal(line, payload)
		if err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			continue
		}
		artistMbzIDs, err := utils.ParseUUIDSlice(payload.TrackMeta.AdditionalInfo.ArtistMBIDs)
		if err != nil {
			l.Debug().Err(err).Msg("Failed to parse one or more uuids")
		}
		rgMbzID, err := uuid.Parse(payload.TrackMeta.AdditionalInfo.ReleaseGroupMBID)
		if err != nil {
			rgMbzID = uuid.Nil
		}
		releaseMbzID, err := uuid.Parse(payload.TrackMeta.AdditionalInfo.ReleaseMBID)
		if err != nil {
			releaseMbzID = uuid.Nil
		}
		recordingMbzID, err := uuid.Parse(payload.TrackMeta.AdditionalInfo.RecordingMBID)
		if err != nil {
			recordingMbzID = uuid.Nil
		}

		var client string
		if payload.TrackMeta.AdditionalInfo.MediaPlayer != "" {
			client = payload.TrackMeta.AdditionalInfo.MediaPlayer
		} else if payload.TrackMeta.AdditionalInfo.SubmissionClient != "" {
			client = payload.TrackMeta.AdditionalInfo.SubmissionClient
		}

		var duration int32
		if payload.TrackMeta.AdditionalInfo.Duration != 0 {
			duration = payload.TrackMeta.AdditionalInfo.Duration
		} else if payload.TrackMeta.AdditionalInfo.DurationMs != 0 {
			duration = payload.TrackMeta.AdditionalInfo.DurationMs / 1000
		}
		opts := catalog.SubmitListenOpts{
			MbzCaller:         mbzc,
			ArtistNames:       payload.TrackMeta.AdditionalInfo.ArtistNames,
			Artist:            payload.TrackMeta.ArtistName,
			ArtistMbzIDs:      artistMbzIDs,
			TrackTitle:        payload.TrackMeta.TrackName,
			RecordingMbzID:    recordingMbzID,
			ReleaseTitle:      payload.TrackMeta.ReleaseName,
			ReleaseMbzID:      releaseMbzID,
			ReleaseGroupMbzID: rgMbzID,
			Duration:          duration,
			Time:              time.Unix(payload.ListenedAt, 0),
			UserID:            1,
			Client:            client,
		}
		err = catalog.SubmitListen(ctx, store, opts)
		if err != nil {
			l.Err(err).Msg("Failed to import LastFM playback item")
			return err
		}
		count++
		throttleFunc()
	}
	l.Info().Msgf("Finished importing %s; imported %d items", filename, count)
	return nil
}
