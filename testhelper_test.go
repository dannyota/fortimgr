package fortimgr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestClient creates an httptest.Server with fixture routing and returns a
// logged-in Client. Fixtures map API URL → JSON data payload (the array inside
// result[0].data). The server is cleaned up when the test finishes.
func newTestClient(t *testing.T, fixtures map[string]string) *Client {
	t.Helper()
	return newTestClientWithProxy(t, fixtures, nil)
}

// newTestClientWithProxy creates a test client with both forward and proxy fixtures.
// forwardFixtures are served via /cgi-bin/module/forward (result[0].data).
// proxyFixtures are served via /cgi-bin/module/flatui_proxy (result[0].data).
func newTestClientWithProxy(t *testing.T, forwardFixtures, proxyFixtures map[string]string) *Client {
	t.Helper()

	mux := http.NewServeMux()

	mux.HandleFunc("/cgi-bin/module/flatui_auth", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:  "HTTP_CSRF_TOKEN",
			Value: "test-token",
			Path:  "/",
		})
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	mux.HandleFunc("/cgi-bin/module/forward", func(w http.ResponseWriter, r *http.Request) {
		if token := r.Header.Get("X-CSRFToken"); token != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		var req struct {
			Params []struct {
				URL string `json:"url"`
			} `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Params) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		apiURL := req.Params[0].URL
		data, ok := forwardFixtures[apiURL]
		if !ok {
			fmt.Fprintf(w, `{"code":0,"data":{"result":[{"status":{"code":-2,"message":"unknown url: %s"}}]}}`, apiURL)
			return
		}

		fmt.Fprintf(w, `{"code":0,"data":{"result":[{"status":{"code":0,"message":"OK"},"data":%s}]}}`, data)
	})

	mux.HandleFunc("/cgi-bin/module/flatui_proxy", func(w http.ResponseWriter, r *http.Request) {
		if token := r.Header.Get("xsrf-token"); token != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		var req struct {
			URL    string `json:"url"`
			Method string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Use "url:method" as lookup key to support custom methods.
		key := req.URL
		data, ok := proxyFixtures[key]
		if !ok {
			// Try with method suffix.
			data, ok = proxyFixtures[req.URL+":"+req.Method]
		}
		if !ok {
			fmt.Fprintf(w, `{"result":[{"status":{"code":-2,"message":"unknown url: %s"}}]}`, req.URL)
			return
		}

		fmt.Fprintf(w, `{"result":[{"status":{"code":0,"message":"OK"},"data":%s}]}`, data)
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client, err := NewClient(server.URL, WithCredentials("admin", "pass"))
	if err != nil {
		t.Fatal(err)
	}

	if err := client.Login(context.Background()); err != nil {
		t.Fatal(err)
	}

	return client
}
