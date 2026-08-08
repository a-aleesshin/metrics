package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

const (
	encGzip = "gzip"
	ctJSON  = "application/json"
	ctHTML  = "text/html"
)

var gzipWriterPool = sync.Pool{
	New: func() any { return gzip.NewWriter(io.Discard) },
}

var gzipReaderPool = sync.Pool{
	New: func() any { return new(gzip.Reader) },
}

func DecompressRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Encoding"), encGzip) {
			next.ServeHTTP(w, r)
			return
		}

		gz := gzipReaderPool.Get().(*gzip.Reader)

		if err := gz.Reset(r.Body); err != nil {
			gzipReaderPool.Put(gz)
			http.Error(w, "invalid gzip body", http.StatusBadRequest)
			return
		}

		defer func() {
			_ = gz.Close()
			gzipReaderPool.Put(gz)
		}()
		defer r.Body.Close()

		r.Body = io.NopCloser(gz) // насколько понял стоит обярнуть, чтобы точно знать что мы записываем ReadCloser
		r.Header.Del("Content-Encoding")
		r.Header.Del("Content-Length")

		next.ServeHTTP(w, r)
		return
	})
}

type compressWriter struct {
	http.ResponseWriter
	reqAcceptsGzip bool

	status      int
	wroteHeader bool

	decided bool
	useGzip bool
	zr      *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter, reqAcceptsGzip bool) *compressWriter {
	return &compressWriter{
		ResponseWriter: w,
		reqAcceptsGzip: reqAcceptsGzip,
	}
}

func (cw *compressWriter) WriteHeader(code int) {
	if cw.wroteHeader {
		return
	}

	cw.status = code
	cw.wroteHeader = true
}

func (cw *compressWriter) Write(b []byte) (int, error) {
	if !cw.decided {
		cw.decide()
	}

	if cw.useGzip {
		return cw.zr.Write(b)
	}

	return cw.ResponseWriter.Write(b)
}

func (cw *compressWriter) Close() error {
	if cw.zr == nil {
		return nil
	}

	err := cw.zr.Close()
	gzipWriterPool.Put(cw.zr)
	cw.zr = nil

	return err
}

func (cw *compressWriter) decide() {
	if cw.decided {
		return
	}

	cw.decided = true

	if !cw.wroteHeader {
		cw.status = http.StatusOK
		cw.wroteHeader = true
	}

	ct := strings.ToLower(cw.Header().Get("Content-Type"))
	eligibleType := strings.Contains(ct, ctJSON) || strings.Contains(ct, ctHTML)

	alreadyEncoded := cw.Header().Get("Content-Encoding") != ""
	shouldGzip := cw.reqAcceptsGzip && eligibleType && !alreadyEncoded

	cw.Header().Add("Vary", "Accept-Encoding")

	if shouldGzip {
		cw.useGzip = true
		cw.Header().Set("Content-Encoding", encGzip)
		cw.Header().Del("Content-Length")

		zw := gzipWriterPool.Get().(*gzip.Writer)
		zw.Reset(cw.ResponseWriter)
		cw.zr = zw
	}

	cw.ResponseWriter.WriteHeader(cw.status)
}

func CompressResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok := strings.Contains(r.Header.Get("Accept-Encoding"), encGzip)
		cw := newCompressWriter(w, ok)

		next.ServeHTTP(cw, r)

		if !cw.decided {
			cw.decide()
		}

		_ = cw.Close()
	})
}
