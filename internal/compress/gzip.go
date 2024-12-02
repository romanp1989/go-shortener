package compress

import (
	"compress/gzip"
	"io"
	"net/http"
)

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// NewCompressWriter Factory for create new gzip writer
func NewCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header return header
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write add data to gzip
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader add header data to gzip
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close Function for stop gzip writer
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// NewCompressReader Factory for create gzip reader
func NewCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read Get data from gzip
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close Stop gzip reader
func (c compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
