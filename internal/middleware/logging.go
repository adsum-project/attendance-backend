package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/adsum-project/attendance-backend/pkg/router"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	status int
	bytes  int
}

func NewRequestLogger() router.Middleware {
	return func(handler router.Handler) router.Handler {
		return func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()
			wrappedWriter := &responseWriterWrapper{ResponseWriter: w, status: 0, bytes: 0}

			handler(wrappedWriter, req)

			path := req.URL.Path
			if req.URL.RawQuery != "" {
				path += "?" + req.URL.RawQuery
			}

			duration := time.Since(start)
			status := wrappedWriter.status
			if status == 0 {
				status = http.StatusOK
			}

			log.Printf(
				"%s %s %d %dB %s",
				req.Method,
				path,
				status,
				wrappedWriter.bytes,
				duration.Truncate(time.Millisecond),
			)
		}
	}
}

func (r *responseWriterWrapper) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseWriterWrapper) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

func (r *responseWriterWrapper) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
