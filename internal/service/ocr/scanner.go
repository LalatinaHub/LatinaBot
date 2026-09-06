package ocr

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	vision "cloud.google.com/go/vision/v2/apiv1"
	"cloud.google.com/go/vision/v2/apiv1/visionpb"
	"github.com/LalatinaHub/LatinaBot/pkg/logger"
	"google.golang.org/api/option"
)

var uuidRegex = regexp.MustCompile(`(?i)[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)

// Scanner interface for OCR operations.
type Scanner interface {
	ScanImage(ctx context.Context, imgBytes []byte) (string, error)
	ScanImageURL(ctx context.Context, imgURL string) (string, error)
	ExtractOrderID(text string) string
}

// VisionScanner implements Scanner using Google Cloud Vision API.
type VisionScanner struct {
	client     *vision.ImageAnnotatorClient
	httpClient *http.Client
}

// NewVisionScanner creates a Vision OCR scanner.
// It tries to initialize using serviceAccountURL, local credentials file, or default credentials.
func NewVisionScanner(ctx context.Context, serviceAccountURL string) (*VisionScanner, error) {
	var opts []option.ClientOption

	httpClient := &http.Client{Timeout: 15 * time.Second}

	// 1. Check if SERVICE_ACCOUNT_URL is provided to download credentials
	if serviceAccountURL != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, serviceAccountURL, nil)
		if err == nil {
			resp, err := httpClient.Do(req)
			if err == nil && resp.StatusCode == http.StatusOK {
				data, err := io.ReadAll(resp.Body)
				_ = resp.Body.Close()
				if err == nil && len(data) > 0 {
					_ = os.WriteFile("./gcloud-cred.json", data, 0600)
					opts = append(opts, option.WithCredentialsJSON(data))
					logger.Info().Msg("Loaded Google Cloud credentials from SERVICE_ACCOUNT_URL")
				}
			}
		}
	}

	// 2. Fallback to local file if not loaded
	if len(opts) == 0 {
		if _, err := os.Stat("./gcloud-cred.json"); err == nil {
			opts = append(opts, option.WithCredentialsFile("./gcloud-cred.json"))
			logger.Info().Msg("Loaded Google Cloud credentials from local gcloud-cred.json")
		}
	}

	client, err := vision.NewImageAnnotatorClient(ctx, opts...)
	if err != nil {
		logger.Warn().Err(err).Msg("Google Cloud Vision client could not be initialized")
		return &VisionScanner{httpClient: httpClient}, nil
	}

	return &VisionScanner{
		client:     client,
		httpClient: httpClient,
	}, nil
}

// ScanImage runs text detection on raw image bytes.
func (s *VisionScanner) ScanImage(ctx context.Context, imgBytes []byte) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("Google Cloud Vision client not initialized")
	}

	req := &visionpb.BatchAnnotateImagesRequest{
		Requests: []*visionpb.AnnotateImageRequest{
			{
				Image: &visionpb.Image{
					Content: imgBytes,
				},
				Features: []*visionpb.Feature{
					{Type: visionpb.Feature_TEXT_DETECTION},
				},
			},
		},
	}

	res, err := s.client.BatchAnnotateImages(ctx, req)
	if err != nil {
		return "", fmt.Errorf("OCR batch annotation failed: %w", err)
	}

	if len(res.Responses) > 0 {
		resp := res.Responses[0]
		if resp.Error != nil && resp.Error.Code != 0 {
			return "", fmt.Errorf("OCR error (%d): %s", resp.Error.Code, resp.Error.Message)
		}
		if resp.FullTextAnnotation != nil {
			return resp.FullTextAnnotation.Text, nil
		}
		if len(resp.TextAnnotations) > 0 {
			return resp.TextAnnotations[0].Description, nil
		}
	}

	return "", nil
}

// ScanImageURL downloads image from URL and runs text detection.
func (s *VisionScanner) ScanImageURL(ctx context.Context, imgURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imgURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image, status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return s.ScanImage(ctx, data)
}

// ExtractOrderID finds the last UUID occurrence in the given text.
func (s *VisionScanner) ExtractOrderID(text string) string {
	matches := uuidRegex.FindAllString(text, -1)
	if len(matches) == 0 {
		return ""
	}
	return strings.ToLower(matches[len(matches)-1])
}
