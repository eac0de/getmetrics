package compressor

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Реализует http.ResponseWriter, нужен для сжатия и отправки сжатых данных
type compressWriter struct {
	w            http.ResponseWriter
	zw           *gzip.Writer
	contentTypes string
}

func newCompressWriter(w http.ResponseWriter, contentTypes string) *compressWriter {
	return &compressWriter{
		w:            w,
		zw:           gzip.NewWriter(w),
		contentTypes: contentTypes,
	}
}

func (cw *compressWriter) Write(p []byte) (int, error) {
	contentType := cw.w.Header().Get("Content-Type")
	isTypeForCompress := strings.Contains(cw.contentTypes, contentType)
	if isTypeForCompress {
		return cw.zw.Write(p)
	}
	return cw.w.Write(p)
}

func (cw *compressWriter) Header() http.Header {
	return cw.w.Header()
}

func (cw *compressWriter) WriteHeader(statusCode int) {
	contentType := cw.w.Header().Get("Content-Type")
	isTypeForCompress := strings.Contains(cw.contentTypes, contentType)
	if isTypeForCompress {
		cw.w.Header().Set("Content-Encoding", "gzip")
	}
	cw.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// Реализует io.ReadCloser, нужен для чтения сжатых данных
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

func GzipData(p []byte) ([]byte, error) {
	var b bytes.Buffer
	w, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %v", err)
	}
	_, err = w.Write(p)
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %v", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %v", err)
	}
	return b.Bytes(), nil
}
