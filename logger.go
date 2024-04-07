// Package traefik_middleware_request_logger Traefik middleware to log incoming requests and outgoing responses.
package traefik_middleware_request_logger //nolint:revive,stylecheck

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

// Config holds the plugin configuration.
type Config struct {
	RequestIDHeaderName string      `json:"RequestIDHeaderName,omitempty"` //nolint:tagliatelle // This is a configuration option
	StatusCodes         []int       `json:"StatusCodes,omitempty"`         //nolint:tagliatelle // This is a configuration option
	ContentTypes        []string    `json:"ContentTypes,omitempty"`        //nolint:tagliatelle // This is a configuration option
	Limits              ConfigLimit `json:"Limits,omitempty"`              //nolint:tagliatelle // This is a configuration option
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
	Body             string            `json:"body"`              //nolint:tagliatelle
	Verb             string            `json:"verb"`              //nolint:tagliatelle
	IPAddress        string            `json:"ip_address"`        //nolint:tagliatelle
	Time             string            `json:"time"`              //nolint:tagliatelle
	TransferEncoding string            `json:"transfer_encoding"` //nolint:tagliatelle
}

type responseData struct {
	Time             string            `json:"time"`              //nolint:tagliatelle
	Status           int               `json:"status"`            //nolint:tagliatelle
	Headers          map[string]string `json:"headers"`           //nolint:tagliatelle
	Body             string            `json:"body"`              //nolint:tagliatelle
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
	}, nil
}

func (p *logRequest) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID, _ := generateRandomID(16)

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
		w:    w,
		buf:  &bytes.Buffer{},
		code: 200,
	}

	p.next.ServeHTTP(resp, r)
	defer resp.Flush()

	respBodyBytes := resp.buf.Bytes()
	_, _ = w.Write(respBodyBytes)

	headers := make(map[string]string)
	for name, values := range r.Header {
		headers[name] = values[0] // Take the first value of the header
	}

	reqData := requestData{
		URI:     r.URL.String(),
		Host:    r.Host,
		Headers: headers,
		Time:    time.Now().Format(time.RFC3339),
		Verb:    r.Method,
	}

	if allowBodySize(len(requestBody), p.maxBodySize) && allowContentType(r.Header.Get("Content-Type"), p.contentTypes) {
		reqData.Body = string(requestBody)
	} else {
		reqData.Body = fmt.Sprintf("Request body too large to log or wrong content type. Size: %d bytes, Content-type: %s", len(requestBody), r.Header.Get("Content-Type"))
	}

	responseBody := io.NopCloser(bytes.NewBuffer(respBodyBytes))
	responseBodyBytes, _ := io.ReadAll(responseBody)

	respHeaders := make(map[string]string)
	for name, values := range resp.Header() {
		respHeaders[name] = values[0] // Take the first value of the header
	}

	resData := responseData{
		Headers: respHeaders,
		Status:  resp.code,
		Time:    time.Now().Format(time.RFC3339),
	}

	if allowBodySize(len(responseBodyBytes), p.maxBodySize) && allowContentType(resp.Header().Get("Content-Type"), p.contentTypes) {
		resData.Body = string(responseBodyBytes)
	} else {
		resData.Body = fmt.Sprintf("Response body too large to log or wrong content type. Size: %d bytes, Content-type: %s", len(responseBodyBytes), resp.Header().Get("Content-Type"))
	}

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

	if allowStatusCode(resp.code, p.statusCodes) {
		_, err = os.Stdout.WriteString(string(jsonData) + "\n")
		if err != nil {
			fmt.Println(err)
		}
	}
}

type wrappedResponseWriter struct {
	w    http.ResponseWriter
	buf  *bytes.Buffer
	code int
}

func (w *wrappedResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *wrappedResponseWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

func (w *wrappedResponseWriter) WriteHeader(code int) {
	w.code = code
}

func (w *wrappedResponseWriter) Flush() {
	w.w.WriteHeader(w.code)
	_, err := io.Copy(w.w, w.buf)
	if err != nil {
		fmt.Println(err)
	}
}

func (w *wrappedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.w.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("%T is not an http.Hijacker", w.w)
	}

	return hijacker.Hijack()
}

func generateRandomID(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
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
