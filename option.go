package fortimgr

import (
	"net/http"
	"time"
)

const (
	defaultTimeout   = 30 * time.Second
	defaultUserAgent = "fortimgr-go/0.1"
)

// ClientOption configures the Client.
type ClientOption interface {
	apply(*clientConfig)
}

// clientConfig holds resolved options.
type clientConfig struct {
	username           string
	password           string
	insecureTLS        bool
	x509NegativeSerial bool
	timeout            time.Duration
	transport          http.RoundTripper
	httpClient         *http.Client
	userAgent          string
}

// WithCredentials sets username and password for FlatUI auth.
type withCredentials struct{ username, password string }

func (o withCredentials) apply(c *clientConfig) { c.username = o.username; c.password = o.password }
func WithCredentials(username, password string) ClientOption {
	return withCredentials{username, password}
}

// WithInsecureTLS disables TLS certificate verification.
type withInsecureTLS struct{}

func (withInsecureTLS) apply(c *clientConfig) { c.insecureTLS = true }
func WithInsecureTLS() ClientOption            { return withInsecureTLS{} }

// WithTimeout sets the HTTP client timeout.
type withTimeout struct{ d time.Duration }

func (o withTimeout) apply(c *clientConfig) { c.timeout = o.d }
func WithTimeout(d time.Duration) ClientOption { return withTimeout{d} }

// WithTransport sets a custom RoundTripper (e.g. for rate limiting).
type withTransport struct{ rt http.RoundTripper }

func (o withTransport) apply(c *clientConfig) { c.transport = o.rt }
func WithTransport(rt http.RoundTripper) ClientOption { return withTransport{rt} }

// WithHTTPClient replaces the entire HTTP client.
type withHTTPClient struct{ c *http.Client }

func (o withHTTPClient) apply(c *clientConfig) { c.httpClient = o.c }
func WithHTTPClient(hc *http.Client) ClientOption { return withHTTPClient{hc} }

// WithUserAgent overrides the User-Agent header.
type withUserAgent struct{ ua string }

func (o withUserAgent) apply(c *clientConfig) { c.userAgent = o.ua }
func WithUserAgent(ua string) ClientOption { return withUserAgent{ua} }

// WithX509NegativeSerial enables Go's x509negativeserial GODEBUG flag.
// Some FortiManager appliances use TLS certificates with negative serial
// numbers (non-RFC 5280). Without this option, Go's TLS client rejects
// these certificates with "x509: negative serial number".
type withX509NegativeSerial struct{}

func (withX509NegativeSerial) apply(c *clientConfig) { c.x509NegativeSerial = true }
func WithX509NegativeSerial() ClientOption            { return withX509NegativeSerial{} }
