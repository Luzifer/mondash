package main

import "net/http"

// LogResponseWriter wraps the standard http.ResponseWritter allowing for more
// verbose logging
type LogResponseWriter struct {
	Status int
	Size   int
	http.ResponseWriter
}

// NewLogResponseWriter instanciates a new LogResponseWriter
func NewLogResponseWriter(res http.ResponseWriter) *LogResponseWriter {
	// Default the status code to 200
	return &LogResponseWriter{200, 0, res}
}

// Header returns & satisfies the http.ResponseWriter interface
func (w *LogResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// Write satisfies the http.ResponseWriter interface and
// captures data written, in bytes
func (w *LogResponseWriter) Write(data []byte) (int, error) {

	written, err := w.ResponseWriter.Write(data)
	w.Size += written

	return written, err
}

// WriteHeader satisfies the http.ResponseWriter interface and
// allows us to cach the status code
func (w *LogResponseWriter) WriteHeader(statusCode int) {

	w.Status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
