// Package traefik_middleware_request_logger_test This file contains the tests for the logger middleware.
package traefik_middleware_request_logger_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	traefikmiddlewarerequestlogger "github.com/cookielab/traefik-middleware-request-logger" //nolint:depguard
)

func TestGetPlaintext(t *testing.T) {
	cfg := prepareConfig([]string{"text/plain"}, 5, []int{})

	ctx := context.Background()
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("helloWorld"))
		if err != nil {
			t.Fatal(err)
		}
	})

	handler, err := traefikmiddlewarerequestlogger.New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	assertHeaderExists(t, req)
	assertBody(t, recorder, "helloWorld")
}

func TestGetJson(t *testing.T) {
	cfg := prepareConfig([]string{"application/json"}, 1024, []int{})

	ctx := context.Background()
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"message": "hello"}`))
		if err != nil {
			t.Fatal(err)
		}
	})

	handler, err := traefikmiddlewarerequestlogger.New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	assertHeaderExists(t, req)
}

func TestPostJson(t *testing.T) {
	cfg := prepareConfig([]string{"application/json"}, 1024, []int{})

	ctx := context.Background()
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"message": "hello"}`))
		if err != nil {
			t.Fatal(err)
		}
	})

	handler, err := traefikmiddlewarerequestlogger.New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost", strings.NewReader(`{"message": "hello"}`))
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	assertHeaderExists(t, req)
}

func TestPostJsonWithBodyLimit(t *testing.T) {
	cfg := prepareConfig([]string{"application/json"}, 5, []int{})

	ctx := context.Background()
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"message": "hello"}`))
		if err != nil {
			t.Fatal(err)
		}
	})

	handler, err := traefikmiddlewarerequestlogger.New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost", strings.NewReader(`{"message": "hello"}`))
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	assertHeaderExists(t, req)
}

func TestPostJsonWithContentType(t *testing.T) {
	cfg := prepareConfig([]string{"application/json"}, 1024, []int{})

	ctx := context.Background()
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"message": "hello"}`))
		if err != nil {
			t.Fatal(err)
		}
	})

	handler, err := traefikmiddlewarerequestlogger.New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost", strings.NewReader(`{"message": "hello"}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/json")

	handler.ServeHTTP(recorder, req)

	assertHeaderExists(t, req)
}

func TestPostJsonWithStatusCode(t *testing.T) {
	cfg := prepareConfig([]string{"application/json"}, 1024, []int{http.StatusInternalServerError})

	ctx := context.Background()
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(`{"message": "Internal Server Error"}`))
		if err != nil {
			t.Fatal(err)
		}
	})

	handler, err := traefikmiddlewarerequestlogger.New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost", strings.NewReader(`{"message": "hello"}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/json")

	handler.ServeHTTP(recorder, req)

	assertHeaderExists(t, req)
}

func TestPostJsonWithStatusCodeFalse(t *testing.T) {
	cfg := prepareConfig([]string{"application/json"}, 1024, []int{http.StatusInternalServerError})

	ctx := context.Background()
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"message": "OK"}`))
		if err != nil {
			t.Fatal(err)
		}
	})

	handler, err := traefikmiddlewarerequestlogger.New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost", strings.NewReader(`{"message": "hello"}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/json")

	handler.ServeHTTP(recorder, req)

	assertHeaderExists(t, req)
}

// TestServerSentEventsAreNotBuffered makes sure the middleware forwards every chunk
// to the client as it is written instead of holding it until the handler returns.
func TestServerSentEventsAreNotBuffered(t *testing.T) {
	cfg := prepareConfig([]string{"application/json"}, 1024, []int{})

	ctx := context.Background()
	rec := httptest.NewRecorder()
	seen := make(chan string, 2)
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		// The interface assertion only resolves under the compiled `go test`; the
		// yaegi interpreter used in CI cannot assert an interpreted type to an
		// interface, so it is treated as optional. The pass-through check below is
		// what actually guards against SSE buffering and runs under both.
		flusher, canFlush := w.(http.Flusher)

		for _, event := range []string{"data: one\n\n", "data: two\n\n"} {
			if _, err := w.Write([]byte(event)); err != nil {
				t.Error(err)

				return
			}

			if !strings.HasSuffix(rec.Body.String(), event) {
				t.Errorf("event %q was buffered instead of written through, client has %q", event, rec.Body.String())
			}

			if canFlush {
				flusher.Flush()

				if !rec.Flushed {
					t.Error("Flush() was not propagated to the underlying ResponseWriter")
				}
			}
			seen <- rec.Body.String()
		}
	})

	handler, err := traefikmiddlewarerequestlogger.New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(rec, req)

	close(seen)
	got := []string{}
	for body := range seen {
		got = append(got, body)
	}

	if len(got) != 2 || got[0] != "data: one\n\n" {
		t.Errorf("first event was not written through before the handler returned: %q", got)
	}
	assertBody(t, rec, "data: one\n\ndata: two\n\n")
}

//nolint:unused
func assertHeader(t *testing.T, req *http.Request, key, expected string) {
	t.Helper()

	if req.Header.Get(key) != expected {
		t.Errorf("invalid header value: %s", req.Header.Get(key))
	}
}

func assertHeaderExists(t *testing.T, req *http.Request) {
	t.Helper()

	if req.Header.Get("X-Request-ID") == "" {
		t.Errorf("missing expected header: %s", "X-Request-ID")
	}
}

func assertBody(t *testing.T, recorder *httptest.ResponseRecorder, expected string) {
	t.Helper()

	if recorder.Body.String() != expected {
		t.Errorf("invalid body value: %s, expected %s", recorder.Body.String(), expected)
	}
}

func prepareConfig(contentTypes []string, maxBodySize int, statusCodes []int) *traefikmiddlewarerequestlogger.Config {
	cfg := traefikmiddlewarerequestlogger.CreateConfig()
	cfg.ContentTypes = contentTypes
	cfg.LogTarget = "stdout"
	cfg.LogTargetURL = ""
	cfg.Limits = traefikmiddlewarerequestlogger.ConfigLimit{MaxBodySize: maxBodySize}
	cfg.RequestIDHeaderName = "X-Request-ID"
	cfg.StatusCodes = statusCodes

	return cfg
}
