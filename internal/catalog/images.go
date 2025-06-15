package catalog

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/google/uuid"
	"github.com/h2non/bimg"
)

type ImageSize string

const (
	ImageSizeSmall  ImageSize = "small"
	ImageSizeMedium ImageSize = "medium"
	ImageSizeLarge  ImageSize = "large"
	// imageSizeXL     ImageSize = "xl"
	ImageSizeFull ImageSize = "full"

	ImageCacheDir = "image_cache"
)

func ImageSourceSize() (size ImageSize) {
	if cfg.FullImageCacheEnabled() {
		size = ImageSizeFull
	} else {
		size = ImageSizeLarge
	}
	return
}

func ParseImageSize(size string) (ImageSize, error) {
	switch strings.ToLower(size) {
	case "small":
		return ImageSizeSmall, nil
	case "medium":
		return ImageSizeMedium, nil
	case "large":
		return ImageSizeLarge, nil
	// case "xl":
	// 	return imageSizeXL, nil
	case "full":
		return ImageSizeFull, nil
	default:
		return "", fmt.Errorf("unknown image size: %s", size)
	}
}
func GetImageSize(size ImageSize) int {
	var px int
	switch size {
	case "small":
		px = 48
	case "medium":
		px = 256
	case "large":
		px = 500
	case "xl":
		px = 1000
	}
	return px
}

func SourceImageDir() string {
	if cfg.FullImageCacheEnabled() {
		return path.Join(cfg.ConfigDir(), ImageCacheDir, "full")
	} else {
		return path.Join(cfg.ConfigDir(), ImageCacheDir, "large")
	}
}

// ValidateImageURL checks if the URL points to a valid image by performing a HEAD request.
func ValidateImageURL(url string) error {
	resp, err := http.Head(url)
	if err != nil {
		return fmt.Errorf("failed to perform HEAD request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HEAD request failed, status code: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return fmt.Errorf("URL does not point to an image, content type: %s", contentType)
	}

	return nil
}

// DownloadAndCacheImage downloads an image from the given URL, then calls CompressAndSaveImage.
func DownloadAndCacheImage(ctx context.Context, id uuid.UUID, url string, size ImageSize) error {
	l := logger.FromContext(ctx)
	err := ValidateImageURL(url)
	if err != nil {
		return err
	}
	l.Debug().Msgf("Downloading image for ID %s", id)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download image, status code: %d", resp.StatusCode)
	}

	return CompressAndSaveImage(ctx, id.String(), size, resp.Body)
}

// Compresses an image to the specified size, then saves it to the correct cache folder.
func CompressAndSaveImage(ctx context.Context, filename string, size ImageSize, body io.Reader) error {
	l := logger.FromContext(ctx)

	if size == ImageSizeFull {
		return saveImage(filename, size, body)
	}

	l.Debug().Msg("Creating resized image")
	compressed, err := compressImage(size, body)
	if err != nil {
		return err
	}

	return saveImage(filename, size, compressed)
}

// SaveImage saves an image to the image_cache/{size} folder
func saveImage(filename string, size ImageSize, data io.Reader) error {
	configDir := cfg.ConfigDir()
	cacheDir := filepath.Join(configDir, ImageCacheDir)

	// Ensure the cache directory exists
	err := os.MkdirAll(filepath.Join(cacheDir, string(size)), 0744)
	if err != nil {
		return fmt.Errorf("failed to create full image cache directory: %w", err)
	}

	// Create a file in the cache directory
	imagePath := filepath.Join(cacheDir, string(size), filename)
	file, err := os.Create(imagePath)
	if err != nil {
		return fmt.Errorf("failed to create image file: %w", err)
	}
	defer file.Close()

	// Save the image to the file
	_, err = io.Copy(file, data)
	if err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	return nil
}

func compressImage(size ImageSize, data io.Reader) (io.Reader, error) {
	imgBytes, err := io.ReadAll(data)
	if err != nil {
		return nil, err
	}
	px := GetImageSize(size)
	// Resize with bimg
	imgBytes, err = bimg.NewImage(imgBytes).Process(bimg.Options{
		Width:         px,
		Height:        px,
		Crop:          true,
		Quality:       85,
		StripMetadata: true,
		Type:          bimg.WEBP,
	})
	if err != nil {
		return nil, err
	}
	if len(imgBytes) == 0 {
		return nil, fmt.Errorf("compression failed")
	}
	return bytes.NewReader(imgBytes), nil
}

func DeleteImage(filename uuid.UUID) error {
	configDir := cfg.ConfigDir()
	cacheDir := filepath.Join(configDir, ImageCacheDir)

	// err := os.Remove(path.Join(cacheDir, "xl", filename.String()))
	// if err != nil && !os.IsNotExist(err) {
	// 	return err
	// }
	err := os.Remove(path.Join(cacheDir, "full", filename.String()))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	err = os.Remove(path.Join(cacheDir, "large", filename.String()))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	err = os.Remove(path.Join(cacheDir, "medium", filename.String()))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	err = os.Remove(path.Join(cacheDir, "small", filename.String()))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Finds any images in all image_cache folders and deletes them if they are not associated with
// an album or artist.
func PruneOrphanedImages(ctx context.Context, store db.DB) error {
	l := logger.FromContext(ctx)

	configDir := cfg.ConfigDir()
	cacheDir := filepath.Join(configDir, ImageCacheDir)

	count := 0
	// go through every folder to find orphaned images
	// store already processed images to speed up pruining
	memo := make(map[string]bool)
	for _, dir := range []string{"large", "medium", "small", "full"} {
		c, err := pruneDirImgs(ctx, store, path.Join(cacheDir, dir), memo)
		if err != nil {
			return err
		}
		count += c
	}
	l.Info().Msgf("Purged %d images", count)
	return nil
}

// returns the number of pruned images
func pruneDirImgs(ctx context.Context, store db.DB, path string, memo map[string]bool) (int, error) {
	l := logger.FromContext(ctx)
	count := 0
	files, err := os.ReadDir(path)
	if err != nil {
		l.Info().Msgf("Failed to read from directory %s; skipping for prune", path)
		files = []os.DirEntry{}
	}
	for _, file := range files {
		fn := file.Name()
		imageid, err := uuid.Parse(fn)
		if err != nil {
			l.Debug().Msgf("Filename does not appear to be UUID: %s", fn)
			continue
		}
		exists, err := store.ImageHasAssociation(ctx, imageid)
		if err != nil {
			return 0, err
		} else if exists {
			continue
		}
		// image does not have association
		l.Debug().Msgf("Deleting image: %s", imageid)
		err = DeleteImage(imageid)
		if err != nil {
			l.Err(err).Msg("Error purging orphaned images")
		}
		if memo != nil {
			memo[fn] = true
		}
		count++
	}
	return count, nil
}
