package repository

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"rasanta/internal/service"
)

type GoogleMapsRepository struct {
	client        *http.Client
	userAgent     string
	defaultCookie string
}

func NewGoogleMapsRepository(timeout time.Duration, userAgent, defaultCookie string) *GoogleMapsRepository {
	return &GoogleMapsRepository{
		client:        &http.Client{Timeout: timeout},
		userAgent:     strings.TrimSpace(userAgent),
		defaultCookie: strings.TrimSpace(defaultCookie),
	}
}

func (r *GoogleMapsRepository) FetchSearch(ctx context.Context, params service.SearchParams) (*service.FetchResult, error) {
	values := url.Values{}
	values.Set("tbm", "map")
	values.Set("authuser", defaultString(params.AuthUser, "0"))
	values.Set("hl", defaultString(params.HL, "id"))
	values.Set("gl", defaultString(params.GL, "id"))
	values.Set("pb", buildSearchPB(params.Lat, params.Lng, 13.1, 56329.51, 732, 683))
	values.Set("q", params.Query)
	values.Set("nfpr", "1")
	values.Set("tch", "1")
	values.Set("ech", "10")

	return r.doRequest(ctx, "https://www.google.com/search", values, nil)
}

func (r *GoogleMapsRepository) FetchDetail(ctx context.Context, params service.DetailParams) (*service.FetchResult, error) {
	values := url.Values{}
	values.Set("authuser", defaultString(params.AuthUser, "0"))
	values.Set("hl", defaultString(params.HL, "id"))
	values.Set("gl", defaultString(params.GL, "id"))
	values.Set("pb", params.PB)
	if params.Q != "" {
		values.Set("q", params.Q)
	}

	return r.doRequest(ctx, "https://www.google.com/maps/preview/place", values, params.ForwardHeaders)
}

func (r *GoogleMapsRepository) doRequest(ctx context.Context, endpoint string, query url.Values, forwardHeaders map[string]string) (*service.FetchResult, error) {
	requestURL := endpoint + "?" + query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "en-US,en;q=0.9,id;q=0.8")
	req.Header.Set("referer", "https://www.google.com/")
	req.Header.Set("user-agent", defaultString(r.userAgent, "Mozilla/5.0"))
	if r.defaultCookie != "" {
		req.Header.Set("cookie", r.defaultCookie)
	}
	for key, value := range forwardHeaders {
		if value == "" {
			continue
		}
		req.Header.Set(key, value)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request google maps: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	result := &service.FetchResult{StatusCode: resp.StatusCode, Body: string(body)}
	if resp.StatusCode >= http.StatusBadRequest {
		snippet := result.Body
		if len(snippet) > 300 {
			snippet = snippet[:300]
		}
		return result, fmt.Errorf("http error %d: %s", resp.StatusCode, snippet)
	}

	return result, nil
}

func buildSearchPB(lat, lng, zoom, radius float64, width, height int) string {
	// Format PB untuk pencarian tempat di Google Maps.
	return fmt.Sprintf(
		"!4m12!1m3!1d%v!2d%v!3d%v!2m3!1f0!2f0!3f0!3m2!1i%d!2i%d!4f%v!7i20!10b1",
		radius,
		lng,
		lat,
		width,
		height,
		zoom,
	)
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
