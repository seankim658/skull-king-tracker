package images

import (
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/nfnt/resize"

	l "github.com/seankim658/skullking/internal/logger"
)

const AvatarImgSize = uint(128)
const AvatarManualKey = "manual"
const AvatarWebPrefixPath = "/static/avatars"
const MaxAvatarSizeBytes = 5 * 1024 * 1024 // 5MB
const MaxAvatarPixelDimension = 4096

const avatarComponent = "service-images"

var allowedAvatarMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

var ErrAvatarValidationFailed = errors.New("avatar pre-downlad validation failed")

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

	// 0. Pre-Download Validations
	finalValidatedURL, valErr := validateImageURLBeforeDownload(ctx, imageURL, MaxAvatarSizeBytes)
	if valErr != nil {
		logger.Warn().Err(valErr).Msg("Pre-download validation failed for avatar")
		return "", valErr
	}
	urlToDownload := finalValidatedURL.String()
	logger.Debug().Str("url_to_download", urlToDownload).Msg("URL passed pre-download validation")

	// 1. Download the image
	getHttpClient := http.Client{
		Timeout: 10 * time.Second,
	}
	getReq, err := http.NewRequestWithContext(ctx, "GET", urlToDownload, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create GET request for avatar download")
		return "", fmt.Errorf("could not prepare avatar download: %w", err)
	}
	getReq.Header.Set("User-Agent", "SkullKingTracker/1.0 (AvatarFetcher)")

	resp, err := getHttpClient.Do(getReq)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to download image")
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Warn().Int(l.StatusCodeKey, resp.StatusCode).Msg("Non-Ok status code when downloading image")
		return "", fmt.Errorf("failed to download image: status %s", resp.Status)
	}

	limitedReader := io.LimitReader(resp.Body, MaxAvatarSizeBytes)

	// 2. Decode the image
	img, formatName, err := image.Decode(limitedReader)
	if err != nil {
		contentType := resp.Header.Get("Content-Type")
		logger.Error().Err(err).Str(l.ContentTypeKey, contentType).Msg("Failed to decode image")
		var bodyPeek []byte
		if rResp, rErr := http.Get(urlToDownload); rErr == nil {
			defer rResp.Body.Close()
			peekReader := io.LimitReader(rResp.Body, 512)
			bodyPeek, _ = io.ReadAll(peekReader)
			logger.Error().Bytes("body_peek", bodyPeek).Msg("Image body peak on decode failure")
		}
		return "", fmt.Errorf("failed to decode image (format may be unsupported or file corrupted): %w", err)
	}
	logger.Debug().Str("decoded_format", formatName).Msg("Image decoded successfully")

	bounds := img.Bounds()
	if bounds.Dx() > MaxAvatarPixelDimension || bounds.Dy() > MaxAvatarPixelDimension {
		logger.Warn().
			Int("width", bounds.Dx()).
			Int("height", bounds.Dy()).
			Int("max_allowed_dimension", MaxAvatarPixelDimension).
			Msg("Image dimensions exceed allowed limits")
		return "", fmt.Errorf("%w: image dimensions (%dx%d) exceed maximum allowed size of %dpx", ErrAvatarValidationFailed, bounds.Dx(), bounds.Dy(), MaxAvatarPixelDimension)
	}
	logger.Debug().Int("width", bounds.Dx()).Int("height", bounds.Dy()).Msg("Image dimensions check passed.")

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

// Returns the potentially redirected URL and an error if validation fails
func validateImageURLBeforeDownload(
	ctx context.Context,
	originalImageURL string,
	maxSizeBytes int64,
) (*url.URL, error) {
	logger := l.WithComponentAndSource(
		l.GetLoggerFromContext(ctx),
		avatarComponent,
		"validateImageURLBeforeDownload",
	).With().Str(l.ImageURLKey, originalImageURL).Logger()

	// 1. Validate URL Scheme and Format
	parsedURL, err := url.ParseRequestURI(originalImageURL)
	if err != nil {
		logger.Warn().Err(err).Msg("Invalid image URL format")
		return nil, fmt.Errorf("%w: invalid image URL: %v", ErrAvatarValidationFailed, err)
	}
	if parsedURL.Scheme != "https" {
		logger.Warn().Str(l.SchemeKey, parsedURL.Scheme).Msg("Unsupported URL scheme, only HTTPS is allowed")
		return nil, fmt.Errorf("%w: unsupported URL scheme: %s", ErrAvatarValidationFailed, parsedURL.Scheme)
	}
	if parsedURL.Hostname() == "" {
		logger.Warn().Msg("URL has no hostname")
		return nil, fmt.Errorf("%w: URL missing hostname", ErrAvatarValidationFailed)
	}

	// 2. SSRF Prevention: Resolve and check IP
	ips, err := net.LookupIP(parsedURL.Hostname())
	if err != nil {
		logger.Warn().Err(err).Str(l.HostnameKey, parsedURL.Hostname()).Msg("Failed to resolve hostname")
		return nil, fmt.Errorf("%w: could not resolve image host: %v", ErrAvatarValidationFailed, err)
	}
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			logger.Warn().Str(l.IPKey, ip.String()).Msg("URL resolves to a restricted IP address")
			return nil, fmt.Errorf("%w: URL points to a restricted network address", ErrAvatarValidationFailed)
		}
	}
	logger.Debug().Str(l.HostnameKey, parsedURL.Hostname()).Msg("Initial hostname IP check passed.")

	// 3. HEAD Request for Content-Length and Content-Type
	httpClient := http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 { // Max 3 redirects
				logger.Warn().Int("redirect_count", len(via)).Str("url", req.URL.String()).Msg("Exceeded redirect limit")
				return fmt.Errorf("too many redirects (%d)", len(via))
			}
			if req.URL.Scheme != "https" {
				logger.Warn().Str("redirect_scheme", req.URL.Scheme).Msg("Redirect to non-HTTPS URL blocked")
				return fmt.Errorf("unsafe redirect: scheme %s is not HTTPS", req.URL.Scheme)
			}
			if req.URL.Hostname() != "" {
				redirectIPs, lookupErr := net.LookupIP(req.URL.Hostname())
				if lookupErr != nil {
					logger.Warn().Err(lookupErr).Str("redirect_host", req.URL.Hostname()).Msg("Failed to resolve redirect hostname")
					return fmt.Errorf("unsafe redirect: could not resolve host %s", req.URL.Hostname())
				}
				for _, rIP := range redirectIPs {
					if rIP.IsLoopback() || rIP.IsPrivate() || rIP.IsLinkLocalUnicast() || rIP.IsLinkLocalMulticast() || rIP.IsUnspecified() {
						logger.Warn().Str("redirect_ip", rIP.String()).Msg("Redirect URL resolves to a restricted IP address")
						return fmt.Errorf("unsafe redirect to restricted IP: %s", rIP.String())
					}
				}
			}
			logger.Debug().Str("redirecting_to", req.URL.String()).Int("redirect_hop", len(via)+1).Msg("Following redirect")
			return nil
		},
	}

	headReq, err := http.NewRequestWithContext(ctx, "HEAD", originalImageURL, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create HEAD request")
		return nil, fmt.Errorf("internal error preparing validation: %w", err)
	}
	headReq.Header.Set("User-Agent", "SkullKingTracker/1.0 (AvatarValidator)")

	headResp, err := httpClient.Do(headReq)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to execute HEAD request")
		return nil, fmt.Errorf("%w: could not connect to validate URL metadata: %v", ErrAvatarValidationFailed, err)
	}
	defer headResp.Body.Close()

	finalURLAfterRedirects := headResp.Request.URL

	if headResp.StatusCode != http.StatusOK {
		logger.Warn().
			Int(l.StatusCodeKey, headResp.StatusCode).
			Str("final_url", finalURLAfterRedirects.String()).
			Msg("Non-OK status from HEAD request")
		return finalURLAfterRedirects, fmt.Errorf("%w: metadata check failed with status %s for %s", ErrAvatarValidationFailed, headResp.Status, finalURLAfterRedirects.String())
	}

	// 4. Check Content-Length
	contentLengthStr := headResp.Header.Get("Content-Length")
	if contentLengthStr != "" {
		contentLength, parseErr := strconv.ParseInt(contentLengthStr, 10, 64)
		if parseErr != nil {
			logger.Warn().Err(parseErr).Str("content_length_header", contentLengthStr).Msg("Failed to parse Content-Length")
			return finalURLAfterRedirects, fmt.Errorf("%w: invalid content length header: %v", ErrAvatarValidationFailed, parseErr)
		}
		if contentLength > maxSizeBytes {
			logger.Warn().Int64(l.SizeBytesKey, contentLength).Msg("Content exceeds maximum allowed size")
			return finalURLAfterRedirects, fmt.Errorf("%w: image is too large (max %dMB)", ErrAvatarValidationFailed, maxSizeBytes/(1024*1024))
		}
		logger.Debug().Int64(l.SizeBytesKey, contentLength).Msg("Size check passed")
	} else {
		logger.Warn().Msg("Content-Length header missing. Download will be limited by MaxAvatarSizeBytes during GET")
	}

	// 5. Check Content-Type
	contentType := headResp.Header.Get("Content-Type")
	if contentType == "" {
		logger.Warn().Msg("Content-Type header is missing")
		return finalURLAfterRedirects, fmt.Errorf("%w: content type unknown", ErrAvatarValidationFailed)
	}
	mimeType := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))

	if !allowedAvatarMimeTypes[mimeType] {
		logger.Warn().Str(l.ContentTypeKey, contentType).Str("parsed_mime_type", mimeType).Msg("Unsupported Content-Type (Strict Policy)")
		return finalURLAfterRedirects, fmt.Errorf("%w: unsupported avatar content type: %s. Only JPEG and PNG are allowed", ErrAvatarValidationFailed, mimeType)
	}
	logger.Debug().Str(l.ContentTypeKey, contentType).Msg("Content-Type check passed (Strict Policy)")

	logger.Info().Str("final_url_after_redirects", finalURLAfterRedirects.String()).Msg("Pre-download validation successful")
	return finalURLAfterRedirects, nil
}
