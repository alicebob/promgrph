package promgrph

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type FakeProm struct {
	URL string
}

func NewFakeProm(t *testing.T) *FakeProm {
	m := http.NewServeMux()
	m.HandleFunc("GET /stuff", func(w http.ResponseWriter, r *http.Request) {
		// I don't know what this does
	})
	s := httptest.NewServer(m)
	t.Cleanup(func() { s.Close() })

	return &FakeProm{
		URL: s.URL,
	}
}
