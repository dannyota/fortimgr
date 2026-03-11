package fortimgr

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	t.Run("empty address", func(t *testing.T) {
		_, err := NewClient("", WithCredentials("u", "p"))
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("no credentials", func(t *testing.T) {
		_, err := NewClient("https://example.com")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("defaults", func(t *testing.T) {
		c, err := NewClient("https://example.com", WithCredentials("u", "p"))
		if err != nil {
			t.Fatal(err)
		}
		if c.config.timeout != defaultTimeout {
			t.Errorf("timeout = %v, want %v", c.config.timeout, defaultTimeout)
		}
		if c.config.userAgent != defaultUserAgent {
			t.Errorf("userAgent = %q, want %q", c.config.userAgent, defaultUserAgent)
		}
	})

	t.Run("trailing slash stripped", func(t *testing.T) {
		c, err := NewClient("https://example.com/", WithCredentials("u", "p"))
		if err != nil {
			t.Fatal(err)
		}
		if c.address != "https://example.com" {
			t.Errorf("address = %q, want no trailing slash", c.address)
		}
	})

	t.Run("all options", func(t *testing.T) {
		transport := &http.Transport{}
		c, err := NewClient("https://example.com",
			WithCredentials("admin", "secret"),
			WithInsecureTLS(),
			WithTimeout(5*time.Second),
			WithTransport(transport),
			WithUserAgent("test/1.0"),
		)
		if err != nil {
			t.Fatal(err)
		}
		if c.config.username != "admin" {
			t.Errorf("username = %q", c.config.username)
		}
		if c.config.timeout != 5*time.Second {
			t.Errorf("timeout = %v", c.config.timeout)
		}
		if c.config.userAgent != "test/1.0" {
			t.Errorf("userAgent = %q", c.config.userAgent)
		}
	})

	t.Run("with http client", func(t *testing.T) {
		hc := &http.Client{Timeout: 10 * time.Second}
		c, err := NewClient("https://example.com",
			WithCredentials("u", "p"),
			WithHTTPClient(hc),
		)
		if err != nil {
			t.Fatal(err)
		}
		if c.httpClient != hc {
			t.Error("expected custom http client")
		}
		if c.httpClient.Jar != c.cookieJar {
			t.Error("expected cookie jar set on custom client")
		}
	})

	t.Run("insecure tls", func(t *testing.T) {
		c, err := NewClient("https://example.com",
			WithCredentials("u", "p"),
			WithInsecureTLS(),
		)
		if err != nil {
			t.Fatal(err)
		}
		tr, ok := c.httpClient.Transport.(*http.Transport)
		if !ok {
			t.Fatal("transport is not *http.Transport")
		}
		if !tr.TLSClientConfig.InsecureSkipVerify {
			t.Error("expected InsecureSkipVerify = true")
		}
	})
}

func TestLogin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.SetCookie(w, &http.Cookie{
				Name:  "HTTP_CSRF_TOKEN",
				Value: "abc123",
				Path:  "/",
			})
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c, err := NewClient(server.URL, WithCredentials("admin", "pass"))
		if err != nil {
			t.Fatal(err)
		}

		if err := c.Login(context.Background()); err != nil {
			t.Fatal(err)
		}

		if !c.LoggedIn() {
			t.Error("expected LoggedIn() = true")
		}
		if c.csrfToken != "abc123" {
			t.Errorf("csrfToken = %q, want %q", c.csrfToken, "abc123")
		}
	})

	t.Run("auth failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// No CSRF cookie set — simulates bad credentials.
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c, err := NewClient(server.URL, WithCredentials("bad", "creds"))
		if err != nil {
			t.Fatal(err)
		}

		err = c.Login(context.Background())
		if err != ErrAuth {
			t.Errorf("err = %v, want ErrAuth", err)
		}
		if c.LoggedIn() {
			t.Error("expected LoggedIn() = false")
		}
	})

	t.Run("certificate error", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		defer server.Close()

		// Client without InsecureTLS will reject the self-signed cert.
		c, err := NewClient(server.URL, WithCredentials("u", "p"))
		if err != nil {
			t.Fatal(err)
		}
		// Override transport to NOT trust the test server's cert.
		c.httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		}

		err = c.Login(context.Background())
		if err == nil {
			t.Fatal("expected certificate error")
		}
		if !isCertificateError(err) {
			t.Errorf("expected certificate error, got: %v", err)
		}
	})
}

func TestLogout(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, err := NewClient("https://example.com", WithCredentials("u", "p"))
		if err != nil {
			t.Fatal(err)
		}
		// Logout when not logged in should be a no-op.
		if err := c.Logout(context.Background()); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("success", func(t *testing.T) {
		var gotToken string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotToken = r.Header.Get("X-CSRFToken")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c, err := NewClient(server.URL, WithCredentials("u", "p"))
		if err != nil {
			t.Fatal(err)
		}
		c.csrfToken = "tok123"

		if err := c.Logout(context.Background()); err != nil {
			t.Fatal(err)
		}
		if c.LoggedIn() {
			t.Error("expected LoggedIn() = false after logout")
		}
		if gotToken != "tok123" {
			t.Errorf("CSRF token sent = %q, want %q", gotToken, "tok123")
		}
	})
}

func TestValidName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"simple", "root", true},
		{"with dash", "customer-a", true},
		{"with underscore", "my_adom", true},
		{"with dot", "fw.01", true},
		{"mixed", "FortiGate-VM64_AWS.1", true},
		{"empty", "", false},
		{"slash", "root/../../etc", false},
		{"space", "my adom", false},
		{"special", "adom;drop", false},
		{"percent", "adom%20", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validName(tt.in)
			if got != tt.want {
				t.Errorf("validName(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestInvalidADOM(t *testing.T) {
	client := newTestClient(t, nil)
	_, err := client.ListDevices(context.Background(), "../etc")
	if err == nil {
		t.Fatal("expected error for invalid ADOM name")
	}
	if !errors.Is(err, ErrInvalidName) {
		t.Errorf("err = %v, want ErrInvalidName", err)
	}
}

func TestAutoRelogin(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/module/flatui_auth" {
			http.SetCookie(w, &http.Cookie{
				Name:  "HTTP_CSRF_TOKEN",
				Value: "new-token",
				Path:  "/",
			})
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
			return
		}

		callCount++
		if callCount == 1 {
			// First call: simulate session expired.
			w.Write([]byte(`{"code":-6,"data":{}}`))
			return
		}
		// Second call (after re-login): success.
		w.Write([]byte(`{"code":0,"data":{"result":[{"status":{"code":0,"message":"OK"},"data":[]}]}}`))
	}))
	defer server.Close()

	c, err := NewClient(server.URL, WithCredentials("admin", "pass"))
	if err != nil {
		t.Fatal(err)
	}
	c.csrfToken = "old-token"

	devices, err := c.ListDevices(context.Background(), "root")
	if err != nil {
		t.Fatalf("expected auto-relogin to succeed, got: %v", err)
	}
	if len(devices) != 0 {
		t.Errorf("len = %d, want 0", len(devices))
	}
	if callCount != 2 {
		t.Errorf("forward called %d times, want 2 (expired + retry)", callCount)
	}
	if c.csrfToken != "new-token" {
		t.Errorf("csrfToken = %q, want \"new-token\"", c.csrfToken)
	}
}

func TestClose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c, err := NewClient(server.URL, WithCredentials("u", "p"))
	if err != nil {
		t.Fatal(err)
	}
	c.csrfToken = "tok"

	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if c.LoggedIn() {
		t.Error("expected LoggedIn() = false after Close")
	}
}
