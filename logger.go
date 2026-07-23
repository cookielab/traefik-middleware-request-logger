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
	}, nil
}

func (p *logRequest) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.NewString()

	if r.Header.Get(p.requestIDHeaderName) != "" {
		requestID = r.Header.Get(p.requestIDHeaderName)
	}
	r.Header.Set(p.requestIDHeaderName, requestID)

	requestBody := []byte{}
	if r.Body != nil {
		requestBody, _ = io.ReadAll(r.Body)
	}

	r.Body = io.NopCloser(bytes.NewBuffer(requestBody))

	resp := &wrappedResponseWriter{
		w:           w,
		buf:         &bytes.Buffer{},
		maxBodySize: p.maxBodySize,
		code:        200,
	}

	p.next.ServeHTTP(resp, r)
	resp.WriteHeader(resp.code) // ensure the status line is sent even for empty responses

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

	reqData.Body = allowedBody(requestBody, len(requestBody), r.Header.Get("Content-Type"), p.maxBodySize, p.contentTypes, p.skipBodyParse)

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

	resData.Body = allowedBody(resp.buf.Bytes(), resp.size, resp.Header().Get("Content-Type"), p.maxBodySize, p.contentTypes, p.skipBodyParse)

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

	if allowStatusCode(resp.code, p.statusCodes) && p.logTarget == "stdout" {
		_, err = os.Stdout.WriteString(string(jsonData) + "\n")
		if err != nil {
			fmt.Println(err)
		}
	}

	if allowStatusCode(resp.code, p.statusCodes) && p.logTarget == "stderr" {
		_, err = os.Stderr.WriteString(string(jsonData) + "\n")
		if err != nil {
			fmt.Println(err)
		}
	}

	if allowStatusCode(resp.code, p.statusCodes) && p.logTarget == "url" && p.logTargetURL != "" {
		go http.Post(p.logTargetURL, "application/json", bytes.NewBuffer(jsonData)) //nolint:errcheck
	}
}

// wrappedResponseWriter passes the response through to the client as it is written
// and keeps a copy of the body (capped at maxBodySize) for logging. It must never
// hold the body back, otherwise streaming responses (SSE, chunked) stall until the
// upstream handler returns.
type wrappedResponseWriter struct {
	w           http.ResponseWriter
	buf         *bytes.Buffer
	maxBodySize int
	size        int
	code        int
	wroteHeader bool
}

func (w *wrappedResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *wrappedResponseWriter) Write(b []byte) (int, error) {
	w.WriteHeader(w.code)

	if remaining := w.maxBodySize - w.size; remaining > 0 {
		captured := b
		if len(captured) > remaining {
			captured = captured[:remaining]
		}
		w.buf.Write(captured) //nolint:errcheck // bytes.Buffer.Write never returns an error
	}
	w.size += len(b)

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

func allowBodySize(bodySize, maxBodySize int) bool {
	return bodySize <= maxBodySize
}

// allowedBody renders the body for the log. bodySize is the full size of the original
// body, which may be larger than body itself when the capture was truncated.
func allowedBody(body []byte, bodySize int, contentType string, maxBodySize int, contentTypes []string, skipBodyParse bool) interface{} {
	if bodySize == 0 {
		return nil
	}
	if !allowBodySize(bodySize, maxBodySize) || !allowContentType(contentType, contentTypes) {
		return fmt.Sprintf("Request body too large to log or wrong content type. Size: %d bytes, Content-type: %s", bodySize, contentType)
	}

	if skipBodyParse {
		// Try to parse the body as JSON
		var parsedBody interface{}
		if contentType == "application/json" {
			err := json.Unmarshal(body, &parsedBody)
			if err == nil {
				return parsedBody
			}
		}
	}
	// If not JSON, return as string
	return string(body)
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
