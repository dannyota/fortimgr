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

// listConfig holds resolved pagination options for a single List call.
type listConfig struct {
	pageSize int
	onPage   func(fetched int, page int)
}

// ListOption configures pagination behavior for SDK List* methods.
//
// By default every List method transparently fetches all pages from the
// FortiManager API (1000 rows per forward request) and returns the
// concatenated result. Callers that want a different page size or a
// progress callback pass ListOptions as variadic trailing arguments:
//
//	addrs, err := client.ListAddresses(ctx, "root",
//	    fortimgr.WithPageSize(500),
//	    fortimgr.WithPageCallback(func(fetched, page int) {
//	        log.Printf("fetched %d addresses so far (page %d)", fetched, page)
//	    }),
//	)
//
// The uses a distinct method name (applyList) from ClientOption.apply so
// the two option types coexist without collision.
type ListOption interface {
	applyList(*listConfig)
}

// listOptFunc adapts a plain func to the ListOption interface.
type listOptFunc func(*listConfig)

func (f listOptFunc) applyList(c *listConfig) { f(c) }

// WithPageSize overrides the default page size (1000 rows per forward
// request) for a single List call. Valid range is 1..10000; values
// outside that range are silently ignored and the default is used.
//
// Smaller page sizes trade latency for lower per-request memory;
// larger page sizes reduce round-trip count. 1000 is a good default
// for most ADOMs.
func WithPageSize(n int) ListOption {
	return listOptFunc(func(c *listConfig) {
		if n >= 1 && n <= 10000 {
			c.pageSize = n
		}
	})
}

// WithPageCallback registers a function invoked after each page has
// been fetched and appended to the result. fetched is the cumulative
// row count; page is the 1-based page number that just completed.
//
// Useful for progress reporting on large lists. The callback runs
// synchronously on the goroutine making the List call, so it should
// not block for long. Returning from the callback does not abort the
// fetch — use context cancellation on the parent ctx for early exit.
func WithPageCallback(fn func(fetched, page int)) ListOption {
	return listOptFunc(func(c *listConfig) {
		c.onPage = fn
	})
}
