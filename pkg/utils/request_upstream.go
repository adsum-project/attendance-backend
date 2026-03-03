package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type UpstreamResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// RequestUpstream performs an HTTP request and returns status, headers, and body.
func RequestUpstream(
	ctx context.Context,
	client *http.Client,
	method string,
	baseURL string,
	upstreamPath string,
	query url.Values,
	headers http.Header,
	body io.Reader,
) (*UpstreamResponse, error) {
	if client == nil {
		return nil, fmt.Errorf("http client is required")
	}

	targetURL := strings.TrimRight(baseURL, "/") + upstreamPath
	if query != nil {
		encodedQuery := query.Encode()
		if encodedQuery != "" {
			targetURL += "?" + encodedQuery
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, targetURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to build upstream request: %w", err)
	}

	for key, values := range headers {
		if strings.EqualFold(key, "Content-Length") {
			continue
		}
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call upstream service: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response body: %w", err)
	}

	return &UpstreamResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header.Clone(),
		Body:       respBody,
	}, nil
}
