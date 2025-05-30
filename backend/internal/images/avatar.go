package images

import (
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"

	l "github.com/seankim658/skullking/internal/logger"
)

const AvatarImgSize = uint(128)
const AvatarManualKey = "manual"
const AvatarWebPrefixPath = "/static/avatars"
const avatarComponent = "service-images"

// Processes and downloads an image
func ProcessAndStoreAvatar(
	ctx context.Context,
	imageURL string,
	userID, storageBasePath, webPathPrefix string,
	targetSize uint,
) (string, error) {
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		avatarComponent,
		"ProcessAndStoreAvatar",
	).With().Str(l.ImageURLKey, imageURL).Str(l.UserIDKey, userID).Logger()

	if imageURL == "" {
		return "", fmt.Errorf("image URL cannot be empty")
	}

	// 1. Download the image
	resp, err := http.Get(imageURL)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to download image")
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Warn().Int(l.StatusCodeKey, resp.StatusCode).Msg("Non-Ok status code when downloading image")
		return "", fmt.Errorf("failed to download image: status %s", resp.Status)
	}

	// 2. Decode the image
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		contentType := resp.Header.Get("Content-Type")
		logger.Error().Err(err).Str(l.ContentTypeKey, contentType).Msg("Failed to decode image")
		var bodyPeek []byte
		if rResp, rErr := http.Get(imageURL); rErr == nil {
			defer rResp.Body.Close()
			limitedReader := io.LimitReader(rResp.Body, 512)
			bodyPeek, _ = io.ReadAll(limitedReader)
			logger.Error().Bytes("body_peek", bodyPeek).Msg("Image body peak on decode failure")
		}
		return "", fmt.Errorf("failed to decode image (format may be unsupported or file corrupted): %w", err)
	}

	// 3. Resize the image
	resizedImg := resize.Resize(targetSize, targetSize, img, resize.Lanczos2)

	// 4. Ensure storage directory exists
	if err := os.MkdirAll(storageBasePath, 0750); err != nil {
		logger.Error().Err(err).Str(l.PathKey, storageBasePath).Msg("Failed to create avatar storage directory")
		return "", fmt.Errorf("failed to create avatar storage directory: %w", err)
	}

	// 5. Save the image as PNG
	avatarFilename := fmt.Sprintf("%s.png", userID)
	fullDiskPath := filepath.Join(storageBasePath, avatarFilename)
	logger = logger.With().Str(l.PathKey, fullDiskPath).Logger()

	outFile, err := os.Create(fullDiskPath)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create avatar file")
		return "", fmt.Errorf("failed to create avatar file: %w", err)
	}
	defer outFile.Close()

	err = png.Encode(outFile, resizedImg)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to encode image to PNG")
		return "", fmt.Errorf("failed to encode image to PNG: %w", err)
	}

	logger.Info().Msg("Avatar processed and stored successfully")

	// 6. Return the relative web path
	relativeWebPath := webPathPrefix
	if relativeWebPath != "" && !strings.HasSuffix(relativeWebPath, "/") {
		relativeWebPath += "/"
	}
	return relativeWebPath + avatarFilename, nil
}
