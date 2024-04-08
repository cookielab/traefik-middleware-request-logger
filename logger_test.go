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
