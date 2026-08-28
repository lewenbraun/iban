// Package elevenlabs transcribes audio recordings through the ElevenLabs
// Scribe v2 speech-to-text HTTP API.
package elevenlabs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"strings"
	"time"
)

const (
	scribeURL        = "https://api.elevenlabs.io/v1/speech-to-text"
	scribeModelID    = "scribe_v2"
	apiTimeout       = 90 * time.Second
	maxResponseBytes = 1 << 20
)

// Client talks to the ElevenLabs speech-to-text API.
type Client struct {
	apiKey string
	client *http.Client
}

type scribeResult struct {
	Text         string  `json:"text"`
	LanguageCode string  `json:"language_code"`
	LanguageProb float64 `json:"language_probability"`
}

type apiErrorDetail struct {
	Detail struct {
		Message string `json:"message"`
	} `json:"detail"`
}

// New creates a client authenticating with the given API key.
func New(apiKey string) *Client {
	return &Client{apiKey: apiKey, client: &http.Client{Timeout: apiTimeout}}
}

// Transcribe uploads the audio file and returns its transcript.
func (c *Client) Transcribe(ctx context.Context, audioPath, lang string) (string, error) {
	body, contentType, err := buildMultipart(audioPath, lang)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, scribeURL, body)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("Content-Type", contentType)
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("call api: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	return decodeResponse(resp)
}

func buildMultipart(audioPath, lang string) (*bytes.Buffer, string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fields := map[string]string{
		"model_id":               scribeModelID,
		"timestamps_granularity": "none",
		"no_verbatim":            "true",
	}
	if lang != "" {
		fields["language_code"] = lang
	}
	for name, value := range fields {
		if err := w.WriteField(name, value); err != nil {
			return nil, "", err
		}
	}
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="recording.wav"`)
	h.Set("Content-Type", "audio/wav")
	part, err := w.CreatePart(h)
	if err != nil {
		return nil, "", err
	}
	if err := copyAudio(part, audioPath); err != nil {
		return nil, "", err
	}
	if err := w.Close(); err != nil {
		return nil, "", err
	}
	return &buf, w.FormDataContentType(), nil
}

func copyAudio(part io.Writer, audioPath string) error {
	f, err := os.Open(audioPath)
	if err != nil {
		return fmt.Errorf("open recording: %w", err)
	}
	defer func() { _ = f.Close() }()
	if _, err := io.Copy(part, f); err != nil {
		return fmt.Errorf("read recording: %w", err)
	}
	return nil
}

func decodeResponse(resp *http.Response) (string, error) {
	payload, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", decodeAPIError(resp.Status, payload)
	}
	var result scribeResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	return strings.TrimSpace(result.Text), nil
}

func decodeAPIError(status string, payload []byte) error {
	var ae apiErrorDetail
	if json.Unmarshal(payload, &ae) == nil && ae.Detail.Message != "" {
		return fmt.Errorf("api %s: %s", status, ae.Detail.Message)
	}
	return fmt.Errorf("api %s: %s", status, payload)
}
