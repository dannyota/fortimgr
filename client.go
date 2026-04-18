package fortimgr

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"sync"
)

// Client communicates with FortiManager via the FlatUI Web UI API.
type Client struct {
	address    string
	config     clientConfig
	httpClient *http.Client
	cookieJar  *cookiejar.Jar
	csrfToken  string
	requestID  int64
}

// NewClient creates a new FortiManager client.
// Address is the base URL (e.g. "https://fm.example.com").
// At minimum, WithCredentials must be provided.
//
// HTTP client precedence: WithHTTPClient > WithTransport > default.
// When WithHTTPClient is used, WithTransport, WithTimeout, and WithInsecureTLS
// are ignored — the caller controls the full HTTP stack.
func NewClient(address string, opts ...ClientOption) (*Client, error) {
	if address == "" {
		return nil, fmt.Errorf("fortimgr: address is required")
	}

	cfg := clientConfig{
		timeout:   defaultTimeout,
		userAgent: defaultUserAgent,
	}
	for _, o := range opts {
		o.apply(&cfg)
	}

	if cfg.username == "" || cfg.password == "" {
		return nil, fmt.Errorf("fortimgr: credentials are required (use WithCredentials)")
	}

	if cfg.x509NegativeSerial {
		setX509NegativeSerial()
	}

	address = strings.TrimRight(address, "/")

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("fortimgr: create cookie jar: %w", err)
	}

	var httpClient *http.Client
	switch {
	case cfg.httpClient != nil:
		httpClient = cfg.httpClient
		httpClient.Jar = jar
	case cfg.transport != nil:
		httpClient = &http.Client{
			Transport: cfg.transport,
			Timeout:   cfg.timeout,
			Jar:       jar,
		}
	default:
		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: cfg.insecureTLS,
				},
			},
			Timeout: cfg.timeout,
			Jar:     jar,
		}
	}

	return &Client{
		address:    address,
		config:     cfg,
		httpClient: httpClient,
		cookieJar:  jar,
	}, nil
}

// Login authenticates with FortiManager and obtains a CSRF token.
func (c *Client) Login(ctx context.Context) error {
	payload := map[string]any{
		"url":    "/gui/userauth",
		"method": "login",
		"params": map[string]any{
			"username":  c.config.username,
			"secretkey": c.config.password,
			"logintype": 0,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("fortimgr: marshal login request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.address+"/cgi-bin/module/flatui_auth", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("fortimgr: create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.config.userAgent != "" {
		req.Header.Set("User-Agent", c.config.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if isCertificateError(err) {
			return fmt.Errorf("%w: %v", ErrCertificate, err)
		}
		return fmt.Errorf("fortimgr: login request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check the actual response headers for the CSRF token, not the cookie jar.
	// Using the jar would incorrectly find stale tokens from previous sessions.
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "HTTP_CSRF_TOKEN" {
			c.csrfToken = cookie.Value
			return nil
		}
	}

	return ErrAuth
}

// Logout terminates the FortiManager session.
// The CSRF token is always cleared, even if the request fails.
func (c *Client) Logout(ctx context.Context) error {
	if c.csrfToken == "" {
		return nil
	}
	defer func() { c.csrfToken = "" }()

	payload := map[string]any{
		"url":    "/gui/userauth",
		"method": "logout",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("fortimgr: marshal logout request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.address+"/cgi-bin/module/flatui_auth", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("fortimgr: create logout request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRFToken", c.csrfToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fortimgr: logout request: %w", err)
	}
	_ = resp.Body.Close()

	return nil
}

// Close logs out and releases resources.
func (c *Client) Close() error {
	return c.Logout(context.Background())
}

// LoggedIn returns true if the client has an active session.
func (c *Client) LoggedIn() bool {
	return c.csrfToken != ""
}

var x509NegativeSerialOnce sync.Once

// setX509NegativeSerial enables Go's x509negativeserial GODEBUG flag.
// Uses sync.Once to safely handle concurrent or repeated calls.
func setX509NegativeSerial() {
	x509NegativeSerialOnce.Do(func() {
		current := os.Getenv("GODEBUG")
		if strings.Contains(current, "x509negativeserial=1") {
			return
		}
		if current == "" {
			_ = os.Setenv("GODEBUG", "x509negativeserial=1")
		} else {
			_ = os.Setenv("GODEBUG", current+",x509negativeserial=1")
		}
	})
}

// validName checks that an ADOM or package name contains only safe characters.
func validName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_' || r == '.':
		default:
			return false
		}
	}
	return true
}
