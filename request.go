package fortimgr

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
)

// forward sends a FlatUI forward request and returns the data payload.
// If the session has expired, it re-authenticates once and retries.
func (c *Client) forward(ctx context.Context, apiURL string) (json.RawMessage, error) {
	data, err := c.doForward(ctx, apiURL)
	if errors.Is(err, ErrSessionExpired) {
		if loginErr := c.Login(ctx); loginErr != nil {
			return nil, fmt.Errorf("fortimgr: re-login after session expired: %w", loginErr)
		}
		return c.doForward(ctx, apiURL)
	}
	return data, err
}

// doForward performs a single FlatUI forward request without retry.
func (c *Client) doForward(ctx context.Context, apiURL string) (json.RawMessage, error) {
	return c.doForwardExtra(ctx, apiURL, nil)
}

// get forwards a request and unmarshals the data payload into []T.
func get[T any](ctx context.Context, c *Client, apiURL string) ([]T, error) {
	data, err := c.forward(ctx, apiURL)
	if err != nil {
		return nil, err
	}
	var items []T
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("fortimgr: unmarshal response: %w", err)
	}
	return items, nil
}

// forwardExtra sends a FlatUI forward request whose params[0] merges extra
// fields (e.g. "filter" or "option") alongside the URL. Used by endpoints
// that require query-time filters.
func (c *Client) forwardExtra(ctx context.Context, apiURL string, extra map[string]any) (json.RawMessage, error) {
	data, err := c.doForwardExtra(ctx, apiURL, extra)
	if errors.Is(err, ErrSessionExpired) {
		if loginErr := c.Login(ctx); loginErr != nil {
			return nil, fmt.Errorf("fortimgr: re-login after session expired: %w", loginErr)
		}
		return c.doForwardExtra(ctx, apiURL, extra)
	}
	return data, err
}

// doForwardExtra performs a single FlatUI forward request without retry,
// merging extra keys into params[0].
func (c *Client) doForwardExtra(ctx context.Context, apiURL string, extra map[string]any) (json.RawMessage, error) {
	param := map[string]any{"url": apiURL}
	for k, v := range extra {
		if k == "url" {
			continue
		}
		param[k] = v
	}
	payload := map[string]any{
		"id":     atomic.AddInt64(&c.requestID, 1),
		"method": "get",
		"params": []map[string]any{param},
	}

	var result flatUIResponse
	if err := c.postModule(ctx, "/cgi-bin/module/forward", "X-CSRFToken", payload, &result); err != nil {
		return nil, err
	}
	return checkResponse(&result)
}

// getExtra forwards a request with extra params and unmarshals the data
// payload into []T. Use this when the endpoint needs a filter, option, or
// other field alongside the URL.
func getExtra[T any](ctx context.Context, c *Client, apiURL string, extra map[string]any) ([]T, error) {
	data, err := c.forwardExtra(ctx, apiURL, extra)
	if err != nil {
		return nil, err
	}
	var items []T
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("fortimgr: unmarshal response: %w", err)
	}
	return items, nil
}

// defaultPageSize is the number of rows requested per forward call
// when a List method is invoked without WithPageSize. Chosen to balance
// round-trip count against per-request memory/bandwidth; FortiManager
// handles 1000-row responses comfortably on every endpoint we've tested.
const defaultPageSize = 1000

// buildListConfig applies functional List options to a fresh listConfig.
// Always returns a config with pageSize in [1, 10000]; values outside that
// range (or unset) fall back to defaultPageSize.
func buildListConfig(opts []ListOption) listConfig {
	cfg := listConfig{pageSize: defaultPageSize}
	for _, o := range opts {
		if o != nil {
			o.applyList(&cfg)
		}
	}
	if cfg.pageSize < 1 || cfg.pageSize > 10000 {
		cfg.pageSize = defaultPageSize
	}
	return cfg
}

// maxPagedIterations is a belt-and-suspenders safety cap to prevent an
// infinite loop in getPaged even if the dedup and count-based termination
// checks both fail. At the default 1000-row pageSize it allows fetching
// 10 million rows per List call, which is orders of magnitude larger than
// any realistic FortiManager deployment.
const maxPagedIterations = 10000

// getPaged fetches every page of a list endpoint sequentially and
// concatenates the results into a single slice of T.
//
// Each page request calls getExtra with extras merged against a per-page
// range parameter: {"range": [offset, pageSize]}. The range key in extras
// is ignored to prevent callers from overriding pagination state. Session
// expiry between pages is transparently handled by forwardExtra's retry
// path (login + one retry per affected page). Context cancellation is
// respected at the boundary between pages.
//
// Termination (in priority order):
//
//  1. A page returns fewer rows than pageSize → break (normal
//     completion, includes the empty 0-row case).
//  2. A page returns MORE rows than pageSize → break. The endpoint
//     ignored our range parameter and returned its full dataset on the
//     first call; continuing would spin forever with the same result.
//  3. A page returns EXACTLY pageSize rows that byte-match page 1 →
//     break. This catches the other "endpoint ignores offset" pattern
//     where the return count accidentally equals pageSize (e.g. a
//     dataset of 2 rows fetched with WithPageSize(2) against an
//     endpoint that ignores the offset field). Without this check we
//     would loop forever appending duplicate rows.
//  4. An absolute maxPagedIterations safety cap — if neither 1, 2, nor
//     3 ever triggers, getPaged aborts and returns an error rather
//     than spinning forever. This should never fire in practice.
//
// On any page error the accumulated result is discarded and (nil, err)
// is returned — callers never receive partial data with a nil error.
//
// Pagination is NOT a point-in-time snapshot of FortiManager state: rows
// added to the underlying list between two page fetches may appear in a
// later page, and rows deleted may be skipped. Callers that need
// consistency must serialize against concurrent writers.
func getPaged[T any](ctx context.Context, c *Client, apiURL string, extras map[string]any, cfg listConfig) ([]T, error) {
	var all []T
	var page1Bytes []byte
	offset := 0

	for page := 1; page <= maxPagedIterations; page++ {
		// Merge caller extras with the per-page range. Caller-provided
		// "range" is deliberately ignored so pagination state can't be
		// clobbered.
		ext := make(map[string]any, len(extras)+1)
		for k, v := range extras {
			if k == "range" {
				continue
			}
			ext[k] = v
		}
		ext["range"] = []int{offset, cfg.pageSize}

		items, err := getExtra[T](ctx, c, apiURL, ext)
		if err != nil {
			return nil, err
		}

		// Termination rule 3 (same-data detection) — compute BEFORE
		// appending so we don't accumulate duplicates. If page 2 (or
		// later) returns a byte-identical page to page 1, the endpoint
		// is ignoring our offset; stop here and return what we already
		// appended on page 1.
		if page > 1 && len(items) == cfg.pageSize && len(page1Bytes) > 0 {
			currBytes, marshalErr := json.Marshal(items)
			if marshalErr == nil && bytes.Equal(currBytes, page1Bytes) {
				return all, nil
			}
		}
		if page == 1 && len(items) == cfg.pageSize {
			// Only snapshot page 1's bytes when the full-page case
			// matters (returned exactly pageSize). Under-full and
			// over-full short-circuit below without needing the snapshot.
			if snap, marshalErr := json.Marshal(items); marshalErr == nil {
				page1Bytes = snap
			}
		}

		all = append(all, items...)

		if cfg.onPage != nil {
			cfg.onPage(len(all), page)
		}

		// Termination rule 1 — under-full page.
		if len(items) < cfg.pageSize {
			return all, nil
		}
		// Termination rule 2 — over-full page (endpoint ignored range).
		if len(items) > cfg.pageSize {
			return all, nil
		}

		offset += cfg.pageSize
	}
	return nil, fmt.Errorf("fortimgr: pagination exceeded safety cap of %d iterations at %s — endpoint may be broken or dataset is impossibly large", maxPagedIterations, apiURL)
}

// proxyRequest is the JSON body sent to /cgi-bin/module/flatui_proxy.
type proxyRequest struct {
	URL    string `json:"url"`
	Method string `json:"method"`
	Params any    `json:"params,omitempty"`
}

// proxyResponse is the JSON envelope from /cgi-bin/module/flatui_proxy.
type proxyResponse struct {
	Result []flatUIResult `json:"result"`
}

// proxy sends a FlatUI proxy request and returns the data payload.
// If the session has expired, it re-authenticates once and retries.
func (c *Client) proxy(ctx context.Context, apiURL, method string) (json.RawMessage, error) {
	return c.proxyParams(ctx, apiURL, method, nil)
}

// proxyParams sends a FlatUI proxy request with an optional params object and
// returns the data payload.
func (c *Client) proxyParams(ctx context.Context, apiURL, method string, params any) (json.RawMessage, error) {
	data, err := c.doProxyParams(ctx, apiURL, method, params)
	if errors.Is(err, ErrSessionExpired) {
		if loginErr := c.Login(ctx); loginErr != nil {
			return nil, fmt.Errorf("fortimgr: re-login after session expired: %w", loginErr)
		}
		return c.doProxyParams(ctx, apiURL, method, params)
	}
	return data, err
}

// doProxyParams performs a single FlatUI proxy request without retry.
func (c *Client) doProxyParams(ctx context.Context, apiURL, method string, params any) (json.RawMessage, error) {
	payload := proxyRequest{
		URL:    apiURL,
		Method: method,
		Params: params,
	}

	var result proxyResponse
	if err := c.postModule(ctx, "/cgi-bin/module/flatui_proxy", "xsrf-token", payload, &result); err != nil {
		return nil, err
	}
	return checkProxyResponse(&result)
}

func (c *Client) postModule(ctx context.Context, modulePath, csrfHeader string, payload, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("fortimgr: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.address+modulePath, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("fortimgr: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(csrfHeader, c.csrfToken)
	if c.config.userAgent != "" {
		req.Header.Set("User-Agent", c.config.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if isCertificateError(err) {
			return fmt.Errorf("%w: %v", ErrCertificate, err)
		}
		return fmt.Errorf("fortimgr: send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("fortimgr: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fortimgr: HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("fortimgr: parse response: %w", err)
	}
	return nil
}

// checkProxyResponse validates the flatui_proxy response envelope.
func checkProxyResponse(resp *proxyResponse) (json.RawMessage, error) {
	if len(resp.Result) == 0 {
		return nil, fmt.Errorf("fortimgr: empty response")
	}

	status := resp.Result[0].Status
	if status.Code == -6 {
		return nil, ErrSessionExpired
	}
	if status.Code == -11 {
		return nil, fmt.Errorf("%w: %s", ErrPermission, status.Message)
	}
	if status.Code != 0 {
		return nil, &APIError{Code: status.Code, Message: status.Message}
	}

	return resp.Result[0].Data, nil
}
