package httputil

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequestBodyReadFailureMessage(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{io.ErrUnexpectedEOF, "Failed to read request body: upload ended before the declared body was received"},
		{fmt.Errorf("read: %w", context.Canceled), "Failed to read request body: client connection closed during upload"},
		{errors.New(`decode Content-Encoding "gzip": unexpected EOF`), "Failed to read request body: invalid compressed body for Content-Encoding"},
		{errors.New(`decode Content-Encoding "br": unsupported Content-Encoding`), "Failed to read request body: unsupported Content-Encoding"},
		{errors.New("boom"), "Failed to read request body"},
	}
	for _, tc := range cases {
		require.Equal(t, tc.want, RequestBodyReadFailureMessage(tc.err))
	}
	require.Equal(t, "max_bytes", RequestBodyReadErrorKind(&http.MaxBytesError{Limit: 1}))
}

func TestRequestBodyReadFailureMessageDoesNotEchoPayload(t *testing.T) {
	err := errors.New(`decode Content-Encoding "gzip": invalid header secret-payload-marker`)
	require.False(t, strings.Contains(RequestBodyReadFailureMessage(err), "secret-payload-marker"))
}

func TestRequestBodyReadErrorKindH2CAndTransport(t *testing.T) {
	require.Equal(t, "client_disconnect", RequestBodyReadErrorKind(errors.New("stream error: stream ID 1; CANCEL")))
	require.Equal(t, "client_disconnect", RequestBodyReadErrorKind(errors.New("client disconnected")))
	require.Equal(t, "truncated_body", RequestBodyReadErrorKind(errors.New("request declared a Content-Length of 10 but only wrote 3 bytes")))
	require.Equal(t, "client_disconnect", RequestBodyReadErrorKind(&net.OpError{Op: "read", Err: syscall.ECONNRESET}))
	require.Equal(t, "transport", RequestBodyReadErrorKind(&net.OpError{Op: "read", Err: os.ErrDeadlineExceeded}))
}
