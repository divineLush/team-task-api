package middleware

import (
	"encoding/json"
	"net/http"
)

const MaxBodySize = 1 << 20 // 1MB

func BodyLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapper := &bodyLimitWriter{ResponseWriter: w, status: http.StatusOK}
		r.Body = http.MaxBytesReader(wrapper, r.Body, MaxBodySize)
		next.ServeHTTP(wrapper, r)
	})
}

type bodyLimitWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *bodyLimitWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.status = status
	if status == http.StatusRequestEntityTooLarge {
		w.ResponseWriter.Header().Set("Content-Type", "application/json")
		w.ResponseWriter.WriteHeader(status)
		json.NewEncoder(w.ResponseWriter).Encode(map[string]string{"error": "request body too large"})
		return
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *bodyLimitWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(w.status)
	}
	return w.ResponseWriter.Write(b)
}
