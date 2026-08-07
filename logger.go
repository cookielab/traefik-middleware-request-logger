// Package traefik_middleware_request_logger Traefik middleware to log incoming requests and outgoing responses.
package traefik_middleware_request_logger //nolint:revive,stylecheck

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid" //nolint:depguard
)

// Config holds the plugin configuration.
type Config struct {
	RequestIDHeaderName string      `json:"RequestIDHeaderName,omitempty"` //nolint:tagliatelle // This is a configuration option
	StatusCodes         []int       `json:"StatusCodes,omitempty"`         //nolint:tagliatelle // This is a configuration option
	ContentTypes        []string    `json:"ContentTypes,omitempty"`        //nolint:tagliatelle // This is a configuration option
	LogTarget           string      `json:"LogTarget,omitempty"`           //nolint:tagliatelle // This is a configuration option
	LogTargetURL        string      `json:"LogTargetUrl,omitempty"`        //nolint:tagliatelle // This is a configuration option
	SkipHeaders         []string    `json:"SkipHeaders,omitempty"`         //nolint:tagliatelle // This is a configuration option
	Limits              ConfigLimit `json:"Limits,omitempty"`              //nolint:tagliatelle // This is a configuration option
	SkipBodyParse       bool        `json:"SkipBodyParse,omitempty"`       //nolint:tagliatelle // This is a configuration option
	RedactBodyFields    []string    `json:"RedactBodyFields,omitempty"`    //nolint:tagliatelle // This is a configuration option
}

// defaultRedactBodyFields is used when RedactBodyFields is not configured.
// Redaction is on by default so credentials never leak into logs unnoticed.
func defaultRedactBodyFields() []string {
	return []string{
		"password", "passwd", "pwd", "secret", "client_secret", "clientSecret",
		"token", "access_token", "accessToken", "refresh_token", "refreshToken",
		"id_token", "idToken", "api_key", "apiKey", "apikey",
		"authorization", "authorizationCode", "authorization_code",
	}
}

// ConfigLimit holds the plugin configuration.
type ConfigLimit struct {
	MaxBodySize int `json:"MaxBodySize,omitempty"` //nolint:tagliatelle // This is a configuration option
}

// CreateConfig creates and initializes the plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// logRequest holds the plugin configuration.
type logRequest struct {
	name                string
	next                http.Handler
	contentTypes        []string
	statusCodes         []int
	maxBodySize         int
	requestIDHeaderName string
	logTarget           string
	logTargetURL        string
	skipHeaders         []string
	skipBodyParse       bool
	redactFields        map[string]bool
	redactJSONRe        *regexp.Regexp
	redactKVRe          *regexp.Regexp
}

// RequestLog holds the plugin configuration.
type requestLog struct {
	RequestID string       `json:"request_id"` //nolint:tagliatelle
	Request   requestData  `json:"request"`    //nolint:tagliatelle
	Response  responseData `json:"response"`   //nolint:tagliatelle
	Direction string       `json:"direction"`  //nolint:tagliatelle
	Metadata  string       `json:"metadata"`   //nolint:tagliatelle
}

type requestData struct {
	URI              string            `json:"uri"`               //nolint:tagliatelle
	Host             string            `json:"host"`              //nolint:tagliatelle
	Headers          map[string]string `json:"headers"`           //nolint:tagliatelle
	Body             interface{}       `json:"body"`              //nolint:tagliatelle
	Verb             string            `json:"verb"`              //nolint:tagliatelle
	IPAddress        string            `json:"ip_address"`        //nolint:tagliatelle
	Time             string            `json:"time"`              //nolint:tagliatelle
	TransferEncoding string            `json:"transfer_encoding"` //nolint:tagliatelle
}

type responseData struct {
	Time             string            `json:"time"`              //nolint:tagliatelle
	Status           int               `json:"status"`            //nolint:tagliatelle
	Headers          map[string]string `json:"headers"`           //nolint:tagliatelle
	Body             interface{}       `json:"body"`              //nolint:tagliatelle
	TransferEncoding string            `json:"transfer_encoding"` //nolint:tagliatelle
}

// New creates and returns a new plugin instance.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	fields := config.RedactBodyFields
	if len(fields) == 0 {
		fields = defaultRedactBodyFields()
	}
	redactFields := make(map[string]bool, len(fields))
	quoted := make([]string, 0, len(fields))
	for _, f := range fields {
		redactFields[strings.ToLower(f)] = true
		quoted = append(quoted, regexp.QuoteMeta(f))
	}
	names := strings.Join(quoted, "|")
	// `"password" : "..."` inside bodies logged as plain strings (incl. escaped values)
	redactJSONRe := regexp.MustCompile(`(?i)"(` + names + `)"\s*:\s*"(?:[^"\\]|\\.)*"`)
	// `password=...` in form/query-style bodies
	redactKVRe := regexp.MustCompile(`(?i)\b(` + names + `)=[^&\s"']+`)

	return &logRequest{
		name:                name,
		next:                next,
		requestIDHeaderName: config.RequestIDHeaderName,
		contentTypes:        config.ContentTypes,
		statusCodes:         config.StatusCodes,
		maxBodySize:         config.Limits.MaxBodySize,
		logTarget:           config.LogTarget,
		logTargetURL:        config.LogTargetURL,
		skipHeaders:         config.SkipHeaders,
		skipBodyParse:       config.SkipBodyParse,
		redactFields:        redactFields,
		redactJSONRe:        redactJSONRe,
		redactKVRe:          redactKVRe,
	}, nil
}

func (p *logRequest) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.NewString()

	if r.Header.Get(p.requestIDHeaderName) != "" {
		requestID = r.Header.Get(p.requestIDHeaderName)
	}
	r.Header.Set(p.requestIDHeaderName, requestID)

	requestBody := &bodyCapture{
		maxBodySize: int64(p.maxBodySize),
		discard:     r.ContentLength > int64(p.maxBodySize),
	}
	if r.Body != nil && !requestBody.discard {
		r.Body = &readCloser{
			Reader: io.TeeReader(r.Body, requestBody),
			Closer: r.Body,
		}
	}

	resp := &wrappedResponseWriter{
		w:    w,
		body: &bodyCapture{maxBodySize: int64(p.maxBodySize)},
		code: 200,
	}

	p.next.ServeHTTP(resp, r)
	if r.Body != nil {
		_, _ = io.Copy(io.Discard, r.Body)
	}
	resp.WriteHeader(resp.code) // ensure the status line is sent even for empty responses

	if !allowStatusCode(resp.code, p.statusCodes) {
		return
	}

	headers := make(map[string]string)
	for name, values := range r.Header {
		if len(values) > 0 && allowedHeader(name, p.skipHeaders) {
			headers[name] = values[0] // Take the first value of the header
		}
	}

	reqData := requestData{
		URI:     r.URL.String(),
		Host:    r.Host,
		Headers: headers,
		Time:    time.Now().Format(time.RFC3339),
		Verb:    r.Method,
	}

	requestBodySize := requestBody.size
	if r.ContentLength > requestBodySize {
		requestBodySize = r.ContentLength
	}
	reqData.Body = p.allowedBody(requestBody.buf.Bytes(), requestBodySize, r.Header.Get("Content-Type"))

	respHeaders := make(map[string]string)
	for name, values := range resp.Header() {
		if len(values) > 0 && allowedHeader(name, p.skipHeaders) {
			respHeaders[name] = values[0] // Take the first value of the header
		}
	}

	resData := responseData{
		Headers: respHeaders,
		Status:  resp.code,
		Time:    time.Now().Format(time.RFC3339),
	}

	resData.Body = p.allowedBody(resp.body.buf.Bytes(), resp.body.size, resp.Header().Get("Content-Type"))

	log := requestLog{
		RequestID: requestID,
		Request:   reqData,
		Response:  resData,
		Direction: "Incomming",
		Metadata:  "",
	}

	jsonData, err := json.Marshal(log)
	if err != nil {
		fmt.Println(err)
	}

	if p.logTarget == "stdout" {
		_, err = os.Stdout.WriteString(string(jsonData) + "\n")
		if err != nil {
			fmt.Println(err)
		}
	}

	if p.logTarget == "stderr" {
		_, err = os.Stderr.WriteString(string(jsonData) + "\n")
		if err != nil {
			fmt.Println(err)
		}
	}

	if p.logTarget == "url" && p.logTargetURL != "" {
		go http.Post(p.logTargetURL, "application/json", bytes.NewBuffer(jsonData)) //nolint:errcheck
	}
}

type readCloser struct {
	io.Reader
	io.Closer
}

// bodyCapture retains the body only while it fits the logging limit. Once the
// limit is crossed, the retained bytes are released and subsequent writes are
// counted without being stored.
type bodyCapture struct {
	buf         bytes.Buffer
	maxBodySize int64
	size        int64
	discard     bool
}

func (b *bodyCapture) Write(p []byte) (int, error) {
	b.size += int64(len(p))
	if b.discard || b.size > b.maxBodySize {
		b.discard = true
		b.buf = bytes.Buffer{}

		return len(p), nil
	}

	return b.buf.Write(p)
}

// wrappedResponseWriter passes the response through to the client as it is written
// and keeps a copy of the body (capped at maxBodySize) for logging. It must never
// hold the body back, otherwise streaming responses (SSE, chunked) stall until the
// upstream handler returns.
type wrappedResponseWriter struct {
	w           http.ResponseWriter
	body        *bodyCapture
	code        int
	wroteHeader bool
}

func (w *wrappedResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *wrappedResponseWriter) Write(b []byte) (int, error) {
	w.WriteHeader(w.code)

	w.body.Write(b) //nolint:errcheck // bytes.Buffer.Write never returns an error

	return w.w.Write(b)
}

func (w *wrappedResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.code = code
	w.w.WriteHeader(code)
}

func (w *wrappedResponseWriter) Flush() {
	w.WriteHeader(w.code)

	if flusher, ok := w.w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *wrappedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.w.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("%T is not an http.Hijacker", w.w)
	}

	return hijacker.Hijack()
}

func allowContentType(contentType string, contentTypes []string) bool {
	if len(contentTypes) == 0 {
		return true
	}
	if contentType == "" {
		return false
	}
	for _, ct := range contentTypes {
		if ct == contentType {
			return true
		}
	}
	return false
}

func allowStatusCode(statusCode int, statusCodes []int) bool {
	if len(statusCodes) == 0 {
		return true
	}
	for _, sc := range statusCodes {
		if sc == statusCode {
			return true
		}
	}
	return false
}

// allowedBody renders the body for the log. bodySize is the full size of the original
// body, which may be larger than body itself when the capture was truncated.
func (p *logRequest) allowedBody(body []byte, bodySize int64, contentType string) interface{} {
	if bodySize == 0 {
		return nil
	}
	if bodySize > int64(p.maxBodySize) || !allowContentType(contentType, p.contentTypes) {
		return fmt.Sprintf("Request body too large to log or wrong content type. Size: %d bytes, Content-type: %s", bodySize, contentType)
	}

	if p.skipBodyParse {
		// Try to parse the body as JSON
		var parsedBody interface{}
		if contentType == "application/json" {
			err := json.Unmarshal(body, &parsedBody)
			if err == nil {
				return p.redact(parsedBody)
			}
		}
	}
	// If not JSON, return as string
	return p.redact(string(body))
}

// redact masks values of sensitive fields in a logged body, whether it was
// parsed into JSON (maps/slices) or kept as a raw string.
func (p *logRequest) redact(body interface{}) interface{} {
	switch t := body.(type) {
	case map[string]interface{}:
		for k, v := range t {
			if p.redactFields[strings.ToLower(k)] {
				t[k] = "[REDACTED]"
			} else {
				t[k] = p.redact(v)
			}
		}
		return t
	case []interface{}:
		for i, v := range t {
			t[i] = p.redact(v)
		}
		return t
	case string:
		s := p.redactJSONRe.ReplaceAllString(t, `"$1":"[REDACTED]"`)
		return p.redactKVRe.ReplaceAllString(s, "$1=[REDACTED]")
	default:
		return body
	}
}

func allowedHeader(headerName string, skipHeaders []string) bool {
	if len(skipHeaders) == 0 {
		return true
	}
	for _, sh := range skipHeaders {
		if sh == headerName {
			return false
		}
	}
	return true
}
