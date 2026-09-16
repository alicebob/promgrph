package promgrph

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/shoenig/test/must"
)

func TestMakeSVGHandler(t *testing.T) {
	prom := NewFakeProm(t)
	c := NewClient(prom.URL)

	m := http.NewServeMux()
	m.Handle("GET /graph.svg", c.MakeSVGHandler("foobar", GraphOpts{
		Title: "<b>hello",
	}))
	s := httptest.NewTestServer(t, m)
	ch := s.Client()

	t.Run("basic graph", func(t *testing.T) {
		resp, err := ch.Get("http://localhost/graph.svg")
		must.NoError(t, err)
		body, _ := io.ReadAll(resp.Body)
		t.Logf("body: %s", body)
		must.Eq(t, "image/svg+xml", resp.Header.Get("content-type"))
	})

	t.Run("graph with gaps", func(t *testing.T) {
		prom := NewFakePromWithGaps(t)
		c := NewClient(prom.URL)
		m := http.NewServeMux()
		m.Handle("GET /gap.svg", c.MakeSVGHandler("foobar", GraphOpts{
			Title:   "gaps",
			Legend:  "test",
			Stacked: false,
		}))
		s := httptest.NewTestServer(t, m)
		defer s.Close()
		ch := s.Client()

		resp, err := ch.Get("http://localhost/gap.svg?width=600")
		must.NoError(t, err)
		body, _ := io.ReadAll(resp.Body)
		svg := string(body)
		// With a gap in the data, we should see multiple "M" commands (at start and at gap)
		// and also "L" commands (connecting points that are close together)
		moveCount := strings.Count(svg, "M ")
		lineCount := strings.Count(svg, " L ")
		must.GreaterEq(t, 2, moveCount)
		must.GreaterEq(t, 2, lineCount)

		must.NoError(t, os.WriteFile("/tmp/graph.svg", []byte(svg), 0600))
	})
}
