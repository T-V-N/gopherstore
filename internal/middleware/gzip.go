package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type Gzip struct {
	GzipReader gzip.Reader
	Body       io.ReadCloser
}

func (g *Gzip) Close() error {
	err := g.GzipReader.Close()

	if err != nil {
		return err
	}

	err = g.Body.Close()

	return err
}

func (g *Gzip) Read(p []byte) (int, error) {
	return g.GzipReader.Read(p)
}

func GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzipReader, err := gzip.NewReader(r.Body)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			gz := Gzip{*gzipReader, r.Body}
			defer gz.Close()

			r.Body = &gz
		}

		next.ServeHTTP(w, r)
	})
}
