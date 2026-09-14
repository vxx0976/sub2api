package handler

import (
	"net/http"
	"strings"

	pkghttputil "github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
	"go.uber.org/zap"
)

// logRequestBodyReadFailure records a bounded, payload-free reason for a body
// read failure. Clients receive the generic message plus a fixed reason from
// pkghttputil.RequestBodyReadFailureMessage; operators get enough information
// to distinguish compression failures from a disconnected/truncated upload
// without logging request content.
func logRequestBodyReadFailure(reqLog *zap.Logger, req *http.Request, err error) {
	if reqLog == nil || err == nil {
		return
	}

	contentLength := int64(-1)
	contentEncoding := "identity"
	if req != nil {
		contentLength = req.ContentLength
		contentEncoding = requestContentEncodingCategory(req.Header.Get("Content-Encoding"))
	}

	reqLog.Warn("read request body failed",
		zap.String("error_kind", requestBodyReadErrorKind(err)),
		zap.String("content_encoding", contentEncoding),
		zap.Int64("content_length", contentLength),
	)
}

func requestContentEncodingCategory(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "identity":
		return "identity"
	case "gzip", "x-gzip":
		return "gzip"
	case "zstd":
		return "zstd"
	case "deflate":
		return "deflate"
	default:
		return "other"
	}
}

func requestBodyReadErrorKind(err error) string {
	return pkghttputil.RequestBodyReadErrorKind(err)
}

func requestBodyReadFailureMessage(err error) string {
	return pkghttputil.RequestBodyReadFailureMessage(err)
}
