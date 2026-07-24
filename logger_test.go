// Package traefik_middleware_request_logger_test This file contains the tests for the logger middleware.
package traefik_middleware_request_logger_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestRedactBodyFields(t *testing.T) {
	ctx := context.Background()
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"authorizationCode":"Y2FydmFnbzpscWwx","requiresMfa":false}`))
	})

	cfg := &traefikmiddlewarerequestlogger.Config{
		RequestIDHeaderName: "X-Request-ID",
		LogTarget:           "stdout",
		Limits:              traefikmiddlewarerequestlogger.ConfigLimit{MaxBodySize: 204800},
		ContentTypes:        []string{"application/json"},
	}

	handler, err := traefikmiddlewarerequestlogger.New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}

	// same shape as a real OAuth login request body
	body := `{"clientId":"dms","email":"user@example.com","password":"Kristian1234","responseType":"code"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/oauth/authorize", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	stdout := captureStdout(t, func() {
		handler.ServeHTTP(httptest.NewRecorder(), req)
	})

	for _, leak := range []string{"Kristian1234", "Y2FydmFnbzpscWwx"} {
		if strings.Contains(stdout, leak) {
			t.Errorf("secret %q leaked into log: %s", leak, stdout)
		}
	}
	for _, want := range []string{`password`, `authorizationCode`, "user@example.com"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("expected %q to stay in log: %s", want, stdout)
		}
	}
	if !strings.Contains(stdout, "[REDACTED]") {
		t.Errorf("expected [REDACTED] marker in log: %s", stdout)
	}
}

func TestRedactFormBody(t *testing.T) {
	ctx := context.Background()
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := &traefikmiddlewarerequestlogger.Config{
		RequestIDHeaderName: "X-Request-ID",
		LogTarget:           "stdout",
		Limits:              traefikmiddlewarerequestlogger.ConfigLimit{MaxBodySize: 204800},
	}

	handler, err := traefikmiddlewarerequestlogger.New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("email=a@b.cz&password=Hunter2!&x=1"))
	req.Header.Set("Content-Type", "text/plain")

	stdout := captureStdout(t, func() {
		handler.ServeHTTP(httptest.NewRecorder(), req)
	})

	if strings.Contains(stdout, "Hunter2") {
		t.Errorf("form password leaked: %s", stdout)
	}
	if !strings.Contains(stdout, "password=[REDACTED]") {
		t.Errorf("expected redacted form field: %s", stdout)
	}
}

func TestRedactCustomFields(t *testing.T) {
	ctx := context.Background()
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := &traefikmiddlewarerequestlogger.Config{
		RequestIDHeaderName: "X-Request-ID",
		LogTarget:           "stdout",
		Limits:              traefikmiddlewarerequestlogger.ConfigLimit{MaxBodySize: 204800},
		RedactBodyFields:    []string{"pin"},
	}

	handler, err := traefikmiddlewarerequestlogger.New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/verify", strings.NewReader(`{"pin":"1234","user":"a"}`))
	req.Header.Set("Content-Type", "application/json")

	stdout := captureStdout(t, func() {
		handler.ServeHTTP(httptest.NewRecorder(), req)
	})

	if strings.Contains(stdout, "1234") {
		t.Errorf("custom field leaked: %s", stdout)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	return string(out)
}
