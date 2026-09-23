package promgrph

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
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
		// First series: 3 points at 2, 10, 20, then gap, then 30 points in 3 segments
		// Segment 1: 302-311 with value 4
		// Segment 2: 312-321 with value 7
		// Segment 3: 322-331 with value 6
		vals := []string{}
		for i := 302; i <= 311; i++ {
			vals = append(vals, fmt.Sprintf("[%.0f,\"4\"]", toFloat(start, i)))
		}
		for i := 312; i <= 321; i++ {
			vals = append(vals, fmt.Sprintf("[%.0f,\"7\"]", toFloat(start, i)))
		}
		for i := 322; i <= 331; i++ {
			vals = append(vals, fmt.Sprintf("[%.0f,\"6\"]", toFloat(start, i)))
		}
		valStr := strings.Join(vals, ",")

		ts0 := toFloat(start, 2)
		ts1 := toFloat(start, 10)
		ts2 := toFloat(start, 20)
		w.Write([]byte(fmt.Sprintf(`{"status":"success","data":{"resultType":"matrix","result":[
{"metric":{"__name__":"up","job":"test","instance":"localhost:9090"},"values":[[%.0f,"1"],[%.0f,"2"],[%.0f,"3"],%s]},
{"metric":{"__name__":"up","job":"test","instance":"localhost:9091"},"values":[[%.0f,"1"],[%.0f,"2"]]}
]}}`,
			ts0, ts1, ts2, valStr,
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

func NewFakePromWithFloats(t *testing.T) *FakeProm {
	m := http.NewServeMux()
	m.HandleFunc("GET /api/v1/query_range", func(w http.ResponseWriter, r *http.Request) {
		start := r.FormValue("start")
		_ = r.FormValue("end")
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
               "__name__" : "cpu_usage",
               "job" : "prometheus"
            },
            "values" : [
               [ %.0f, "0.75" ],
               [ %.0f, "0.80" ],
               [ %.0f, "0.65" ]
            ]
         }
      ]
   }
}`,
		ts0, ts1, ts2)))
	})
	s := httptest.NewServer(m)
	t.Cleanup(func() { s.Close() })

	return &FakeProm{
		URL: s.URL,
	}
}
