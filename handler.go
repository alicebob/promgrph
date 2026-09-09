package promgrph

import (
	"net/http"
)

type Client struct {
	Server string
}

func NewClient(s string) *Client {
	return &Client{
		Server: s,
	}
}

// Returns the actual graphs.
// Usage:
//
//	m.HandleFunc("GET /cpuload.png", c.MakeSVGHandler("node_cpu_seconds_total[5m]"))
func (c *Client) MakeSVGHandler(expr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "image/svg")
		w.Write([]byte("very much todo: " + expr))
	}
}
