// package imagesrc defines interfaces for album and artist image providers
package images

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/gabehf/koito/internal/logger"
	"github.com/google/uuid"
)

type ImageSource struct {
	deezerEnabled bool
	deezerC       *DeezerClient
	caaEnabled    bool
}
type ImageSourceOpts struct {
	UserAgent    string
	EnableCAA    bool
	EnableDeezer bool
}

var once sync.Once
var imgsrc ImageSource

type ArtistImageOpts struct {
	Aliases []string
}

type AlbumImageOpts struct {
	Artists           []string
	Album             string
	ReleaseMbzID      *uuid.UUID
	ReleaseGroupMbzID *uuid.UUID
}

const caaBaseUrl = "https://coverartarchive.org"

// all functions are no-op if no providers are enabled
func Initialize(opts ImageSourceOpts) {
	once.Do(func() {
		if opts.EnableCAA {
			imgsrc.caaEnabled = true
		}
		if opts.EnableDeezer {
			imgsrc.deezerEnabled = true
			imgsrc.deezerC = NewDeezerClient(opts.UserAgent)
		}
	})
}

func GetArtistImage(ctx context.Context, opts ArtistImageOpts) (string, error) {
	l := logger.FromContext(ctx)
	if imgsrc.deezerC != nil {
		img, err := imgsrc.deezerC.GetArtistImages(ctx, opts.Aliases)
		if err != nil {
			return "", err
		}
		return img, nil
	}
	l.Warn().Msg("No image providers are enabled")
	return "", nil
}
func GetAlbumImage(ctx context.Context, opts AlbumImageOpts) (string, error) {
	l := logger.FromContext(ctx)
	if imgsrc.caaEnabled {
		l.Debug().Msg("Attempting to find album image from CoverArtArchive")
		if opts.ReleaseMbzID != nil && *opts.ReleaseMbzID != uuid.Nil {
			url := fmt.Sprintf(caaBaseUrl+"/release/%s/front", opts.ReleaseMbzID.String())
			resp, err := http.DefaultClient.Head(url)
			if err != nil {
				return "", err
			}
			if resp.StatusCode == 200 {
				return url, nil
			}
			l.Debug().Str("url", url).Str("status", resp.Status).Msg("Could not find album cover from CoverArtArchive with MusicBrainz release ID")
		}
		if opts.ReleaseGroupMbzID != nil && *opts.ReleaseGroupMbzID != uuid.Nil {
			url := fmt.Sprintf(caaBaseUrl+"/release-group/%s/front", opts.ReleaseGroupMbzID.String())
			resp, err := http.DefaultClient.Head(url)
			if err != nil {
				return "", err
			}
			if resp.StatusCode == 200 {
				return url, nil
			}
			l.Debug().Str("url", url).Str("status", resp.Status).Msg("Could not find album cover from CoverArtArchive with MusicBrainz release group ID")
		}
	}
	if imgsrc.deezerEnabled {
		l.Debug().Msg("Attempting to find album image from Deezer")
		img, err := imgsrc.deezerC.GetAlbumImages(ctx, opts.Artists, opts.Album)
		if err != nil {
			return "", err
		}
		return img, nil
	}
	l.Warn().Msg("No image providers are enabled")
	return "", nil
}
