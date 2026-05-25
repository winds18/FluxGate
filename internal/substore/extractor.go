package substore

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/winds18/FluxGate/internal/upstream"
)

const maxExtractBytes = 5 << 20

type Extractor struct {
	URLTemplate string
	Client      *http.Client
	Timeout     time.Duration
}

func (e Extractor) Enabled() bool {
	return strings.TrimSpace(e.URLTemplate) != ""
}

func (e Extractor) Extract(ctx context.Context, sourceURL string) (string, error) {
	sourceURL = strings.TrimSpace(sourceURL)
	if err := validateSubscriptionURL(sourceURL); err != nil {
		return "", err
	}
	endpoint, err := e.endpoint(sourceURL)
	if err != nil {
		return "", err
	}

	timeout := e.Timeout
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	client := e.Client
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("user-agent", "FluxGate")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("sub-store extraction returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxExtractBytes+1))
	if err != nil {
		return "", err
	}
	if len(body) > maxExtractBytes {
		return "", fmt.Errorf("sub-store extraction response exceeds %d bytes", maxExtractBytes)
	}
	normalized, err := upstream.NormalizeContent(string(body))
	if err != nil {
		return "", fmt.Errorf("sub-store extraction response: %w", err)
	}
	return normalized, nil
}

func (e Extractor) endpoint(sourceURL string) (string, error) {
	template := strings.TrimSpace(e.URLTemplate)
	if template == "" {
		return "", fmt.Errorf("sub-store extraction template is empty")
	}
	if !strings.Contains(template, "{url}") {
		return "", fmt.Errorf("sub-store extraction template must contain {url}")
	}
	rendered := strings.ReplaceAll(template, "{url}", url.QueryEscape(sourceURL))
	parsed, err := url.Parse(rendered)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid sub-store extraction endpoint")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("unsupported sub-store extraction endpoint scheme")
	}
	return parsed.String(), nil
}

func validateSubscriptionURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("invalid subscription url")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("unsupported subscription url scheme")
	}
	return nil
}
