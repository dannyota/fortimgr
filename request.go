package fortimgr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
)

// forwardRequest is the JSON body sent to /cgi-bin/module/forward.
type forwardRequest struct {
	ID     int64          `json:"id"`
	Method string         `json:"method"`
	Params []forwardParam `json:"params"`
}

type forwardParam struct {
	URL string `json:"url"`
}

// forward sends a FlatUI forward request and returns the data payload.
// If the session has expired, it re-authenticates once and retries.
func (c *Client) forward(ctx context.Context, apiURL string) (json.RawMessage, error) {
	data, err := c.doForward(ctx, apiURL)
	if err == ErrSessionExpired {
		if loginErr := c.Login(ctx); loginErr != nil {
			return nil, fmt.Errorf("fortimgr: re-login after session expired: %w", loginErr)
		}
		return c.doForward(ctx, apiURL)
	}
	return data, err
}

// doForward performs a single FlatUI forward request without retry.
func (c *Client) doForward(ctx context.Context, apiURL string) (json.RawMessage, error) {
	payload := forwardRequest{
		ID:     atomic.AddInt64(&c.requestID, 1),
		Method: "get",
		Params: []forwardParam{{URL: apiURL}},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("fortimgr: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.address+"/cgi-bin/module/forward", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("fortimgr: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRFToken", c.csrfToken)
	if c.config.userAgent != "" {
		req.Header.Set("User-Agent", c.config.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if isCertificateError(err) {
			return nil, fmt.Errorf("%w: %v", ErrCertificate, err)
		}
		return nil, fmt.Errorf("fortimgr: send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("fortimgr: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fortimgr: HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result flatUIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("fortimgr: parse response: %w", err)
	}

	return checkResponse(&result)
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
	data, err := c.doProxy(ctx, apiURL, method)
	if err == ErrSessionExpired {
		if loginErr := c.Login(ctx); loginErr != nil {
			return nil, fmt.Errorf("fortimgr: re-login after session expired: %w", loginErr)
		}
		return c.doProxy(ctx, apiURL, method)
	}
	return data, err
}

// doProxy performs a single FlatUI proxy request without retry.
func (c *Client) doProxy(ctx context.Context, apiURL, method string) (json.RawMessage, error) {
	payload := proxyRequest{
		URL:    apiURL,
		Method: method,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("fortimgr: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.address+"/cgi-bin/module/flatui_proxy", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("fortimgr: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xsrf-token", c.csrfToken)
	if c.config.userAgent != "" {
		req.Header.Set("User-Agent", c.config.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if isCertificateError(err) {
			return nil, fmt.Errorf("%w: %v", ErrCertificate, err)
		}
		return nil, fmt.Errorf("fortimgr: send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("fortimgr: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fortimgr: HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result proxyResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("fortimgr: parse response: %w", err)
	}

	return checkProxyResponse(&result)
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
