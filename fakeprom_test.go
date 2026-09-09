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
	m.HandleFunc("GET /api/v1/query_range", func(w http.ResponseWriter, r *http.Request) {
		// https://prometheus.io/docs/prometheus/latest/querying/api/
		w.Write([]byte(`{
   "status" : "success",
   "data" : {
      "resultType" : "matrix",
      "result" : [
         {
            "metric" : {
               "__name__" : "up",
               "job" : "prometheus",
               "instance" : "localhost:9090"
            },
            "values" : [
               [ 1435781430.781, "1" ],
               [ 1435781445.781, "1" ],
               [ 1435781460.781, "1" ]
            ]
         },
         {
            "metric" : {
               "__name__" : "up",
               "job" : "node",
               "instance" : "localhost:9091"
            },
            "values" : [
               [ 1435781430.781, "0" ],
               [ 1435781445.781, "0" ],
               [ 1435781460.781, "1" ]
            ]
         }
      ]
   }
}`))
	})
	s := httptest.NewServer(m)
	t.Cleanup(func() { s.Close() })

	return &FakeProm{
		URL: s.URL,
	}
}
