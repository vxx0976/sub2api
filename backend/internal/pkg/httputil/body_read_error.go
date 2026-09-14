package httputil

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"syscall"
)

// RequestBodyReadErrorKind classifies a request body read failure into a
// bounded, payload-free reason. The raw error text is never returned because
// decoder errors may echo request bytes.
func RequestBodyReadErrorKind(err error) string {
	if err == nil {
		return "none"
	}
	var maxErr *http.MaxBytesError
	if errors.As(err, &maxErr) {
		return "max_bytes"
	}
	lower := strings.ToLower(err.Error())
	if strings.Contains(lower, "decode content-encoding") {
		if strings.Contains(lower, "unsupported content-encoding") {
			return "unsupported_content_encoding"
		}
		return "decode_content_encoding"
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.EPIPE) {
		return "client_disconnect"
	}
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return "truncated_body"
	}
	// h2c 下 x/net/http2 用自有错误文案表达同样的情况，不 wrap 标准哨兵错误。
	if strings.Contains(lower, "client disconnected") || strings.Contains(lower, "stream error") {
		return "client_disconnect"
	}
	if strings.Contains(lower, "request declared a content-length") {
		return "truncated_body"
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return "transport"
	}
	return "io_read"
}

// RequestBodyReadFailureMessage returns the client-facing message for a body
// read failure. It keeps the historical "Failed to read request body" prefix
// and appends a fixed reason, so clients and ops_error_logs can tell a
// truncated upload from a bad Content-Encoding without exposing payload bytes.
func RequestBodyReadFailureMessage(err error) string {
	const base = "Failed to read request body"
	switch RequestBodyReadErrorKind(err) {
	case "truncated_body":
		return base + ": upload ended before the declared body was received"
	case "client_disconnect":
		return base + ": client connection closed during upload"
	case "decode_content_encoding":
		return base + ": invalid compressed body for Content-Encoding"
	case "unsupported_content_encoding":
		return base + ": unsupported Content-Encoding"
	case "transport":
		return base + ": network read error"
	default:
		return base
	}
}
