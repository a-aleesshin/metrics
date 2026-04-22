package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

const (
	encGzip = "gzip"
	ctJSON  = "application/json"
	ctHTML  = "text/html"
)

func DecompressRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Encoding"), encGzip) {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewReader(r.Body)

		if err != nil {
			http.Error(w, "invalid gzip body", http.StatusBadRequest)
			return
		}

		defer gz.Close()
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
	if cw.zr != nil {
		return cw.zr.Close()
	}

	return nil
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
		cw.zr = gzip.NewWriter(cw.ResponseWriter)
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
