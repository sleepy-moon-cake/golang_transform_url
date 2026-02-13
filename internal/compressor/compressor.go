package compressor

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	http.ResponseWriter
	zw                *gzip.Writer
	compress          bool
	writeHeaderCalled bool
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		ResponseWriter:    w,
		zw:                gzip.NewWriter(w),
		compress:          false,
		writeHeaderCalled: false,
	}
}

func (c *compressWriter) Write(p []byte) (int, error) {
	if !c.writeHeaderCalled {
		c.WriteHeader(http.StatusOK)
	}

	if c.compress {
		return c.zw.Write(p)
	}
	return c.ResponseWriter.Write(p)
}

func (c *compressWriter) Close() error {
	if c.compress {
		return c.zw.Close()
	}
	return nil
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if c.writeHeaderCalled {
		return
	}

	c.writeHeaderCalled = true

	ct := c.Header().Get("Content-Type")

	if statusCode < 300 && (strings.HasPrefix(ct, "application/json") || strings.HasPrefix(ct, "text/html")) {
		c.Header().Set("Content-Encoding", "gzip")
		c.compress = true
	}
	c.ResponseWriter.WriteHeader(statusCode)
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func Compressor(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if shouldDecompressRequest(w, r) {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
		}

		if shouldCompressResponse(r) {
			cw := newCompressWriter(w)
			defer cw.Close()
			w = cw
		}

		h.ServeHTTP(w, r)
	})
}

func shouldCompressResponse(r *http.Request) bool {
	if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		return false
	}
	return true
}

func shouldDecompressRequest(_ http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
		return false
	}

	for _, v := range r.Header.Values("Content-Encoding") {
		if isSupport := strings.Contains(v, "gzip"); isSupport {
			ct := r.Header.Get("Content-Type")
			return strings.HasPrefix(ct, "application/json") || strings.HasPrefix(ct, "text/html")
		}
	}
	return false
}
