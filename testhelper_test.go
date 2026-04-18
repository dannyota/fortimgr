package fortimgr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// sliceFixtureByRange parses a fixture JSON array and a forward request
// range parameter ([offset, count]), returning the sliced sub-array as a
// JSON string. Out-of-range offsets return an empty array. Non-array
// fixtures or malformed range values return (_, false) so the caller
// falls back to the full fixture.
func sliceFixtureByRange(fixture string, rng any) (string, bool) {
	rangeArr, ok := rng.([]any)
	if !ok || len(rangeArr) != 2 {
		return "", false
	}
	offsetF, ok1 := rangeArr[0].(float64)
	countF, ok2 := rangeArr[1].(float64)
	if !ok1 || !ok2 {
		return "", false
	}
	offset := int(offsetF)
	count := int(countF)
	if offset < 0 || count <= 0 {
		return "", false
	}

	var arr []json.RawMessage
	if err := json.Unmarshal([]byte(fixture), &arr); err != nil {
		return "", false
	}

	if offset >= len(arr) {
		return "[]", true
	}
	end := offset + count
	if end > len(arr) {
		end = len(arr)
	}

	out, err := json.Marshal(arr[offset:end])
	if err != nil {
		return "", false
	}
	return string(out), true
}

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
		_, _ = w.Write([]byte(`{}`))
	})

	mux.HandleFunc("/cgi-bin/module/forward", func(w http.ResponseWriter, r *http.Request) {
		if token := r.Header.Get("X-CSRFToken"); token != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// Parse the request body as a generic map so we can inspect both
		// url and range (and any other forwarded params) without tying the
		// shape to a struct.
		var req struct {
			Params []map[string]any `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Params) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		param := req.Params[0]

		apiURL, _ := param["url"].(string)
		data, ok := forwardFixtures[apiURL]
		if !ok {
			_, _ = fmt.Fprintf(w, `{"code":0,"data":{"result":[{"status":{"code":-2,"message":"unknown url: %s"}}]}}`, apiURL)
			return
		}

		// If the request carries a range parameter, slice the fixture array
		// accordingly. This lets pagination tests exercise the multi-page
		// loop with small fixtures + small WithPageSize values, using the
		// shared test client instead of a custom httptest server.
		//
		// Fixtures without a range request pass through untouched so every
		// existing test keeps working unchanged.
		if rng, has := param["range"]; has {
			if sliced, ok := sliceFixtureByRange(data, rng); ok {
				data = sliced
			}
		}

		_, _ = fmt.Fprintf(w, `{"code":0,"data":{"result":[{"status":{"code":0,"message":"OK"},"data":%s}]}}`, data)
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
			_, _ = fmt.Fprintf(w, `{"result":[{"status":{"code":-2,"message":"unknown url: %s"}}]}`, req.URL)
			return
		}

		_, _ = fmt.Fprintf(w, `{"result":[{"status":{"code":0,"message":"OK"},"data":%s}]}`, data)
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
