package promgrph

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

type FakeProm struct {
	URL string
}

func NewFakeProm(t *testing.T) *FakeProm {
	m := http.NewServeMux()
	m.HandleFunc("GET /api/v1/query_range", func(w http.ResponseWriter, r *http.Request) {
		// https://prometheus.io/docs/prometheus/latest/querying/api/
		start := r.FormValue("start")
		_ = r.FormValue("end")
		// Generate timestamps relative to start time
		ts0 := toFloat(start, 0)
		ts1 := toFloat(start, 15)
		ts2 := toFloat(start, 30)
		w.Write([]byte(fmt.Sprintf(`{
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
               [ %.0f, "1" ],
               [ %.0f, "1" ],
               [ %.0f, "1" ]
            ]
         },
         {
            "metric" : {
               "__name__" : "up",
               "job" : "node",
               "instance" : "localhost:9091"
            },
            "values" : [
               [ %.0f, "0" ],
               [ %.0f, "0" ],
               [ %.0f, "1" ]
            ]
         }
      ]
   }
}`, ts0, ts1, ts2, ts0, ts1, ts2)))
	})
	s := httptest.NewServer(m)
	t.Cleanup(func() { s.Close() })

	return &FakeProm{
		URL: s.URL,
	}
}

func NewFakePromWithGaps(t *testing.T) *FakeProm {
	m := http.NewServeMux()
	m.HandleFunc("GET /api/v1/query_range", func(w http.ResponseWriter, r *http.Request) {
		// Return data with regular points, then a gap, then more points
		// Timestamps are relative to the query start time
		start := r.FormValue("start")
		_ = r.FormValue("end")
		ts0 := toFloat(start, 2)
		ts1 := toFloat(start, 10)
		ts2 := toFloat(start, 20)
		ts3 := toFloat(start, 302)
		ts4 := toFloat(start, 313)
		ts5 := toFloat(start, 334)
		w.Write([]byte(fmt.Sprintf(`{"status":"success","data":{"resultType":"matrix","result":[
{"metric":{"__name__":"up","job":"test","instance":"localhost:9090"},"values":[[%.0f,"1"],[%.0f,"2"],[%.0f,"3"],[%.0f,"4"],[%.0f,"5"],[%.0f,"6"]]},
{"metric":{"__name__":"up","job":"test","instance":"localhost:9091"},"values":[[%.0f,"1"],[%.0f,"2"]]}
]}}`,
			ts0, ts1, ts2, ts3, ts4, ts5,
			ts0, ts1,
		)))
	})
	s := httptest.NewServer(m)
	t.Cleanup(func() { s.Close() })

	return &FakeProm{
		URL: s.URL,
	}
}

func toFloat(s string, offset int) float64 {
	v, _ := strconv.Atoi(s)
	return float64(v + offset)
}
