package promgrph

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shoenig/test/must"
)

func TestMakeSVGHandler(t *testing.T) {
	prom := NewFakeProm(t)
	c := NewClient(prom.URL)

	m := http.NewServeMux()
	m.Handle("GET /graph.svg", c.MakeSVGHandler("foobar"))
	s := httptest.NewTestServer(t, m)
	ch := s.Client()

	t.Run("basic graph", func(t *testing.T) {
		resp, err := ch.Get("http://localhost/graph.svg")
		must.NoError(t, err)
		body, _ := io.ReadAll(resp.Body)
		t.Logf("body: %s", body)
		must.Eq(t, "image/svg+xml", resp.Header.Get("content-type"))
	})
}
