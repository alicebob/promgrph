package promgrph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"
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
		ctx := r.Context()
		now := time.Now().UTC()
		q := promQuery{
			Expr:  expr,
			Start: now.Add(-time.Hour),
			End:   now,
		}
		resp, err := c.runQuery(ctx, q)
		if err != nil {
			slog.ErrorContext(ctx, "query failed", "error", err)
			w.WriteHeader(500)
			w.Write([]byte("failed"))
		}

		g := Graph{
			Width:  400,
			Height: 200,
			Title:  "My first query!",
			YAxis:  []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			XAxis:  []int{0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24},
		}
		for _, r := range resp {
			l := Line{
				Color: "green", // FIXME
				Label: r.Metric.Name,
			}
			for _, v := range r.Values {
				val, _ := strconv.Atoi(v[1].(string)) // FIXME
				x := interp(
					v[0].(float64),
					float64(q.Start.Unix()),
					float64(q.End.Unix()),
					g.XAxis[0],
					g.XAxis[len(g.XAxis)-1],
				)
				y := val // FIXME
				l.Points = append(l.Points, [2]int{x, y})
			}
			g.Lines = append(g.Lines, l)
		}

		w.Header().Set("Content-Type", "image/svg+xml")
		renderSVG(w, g)
		// w.Write([]byte(fmt.Sprintf("very much todo: %#v", resp)))
	}
}

type (
	promQuery struct {
		Expr  string
		Start time.Time
		End   time.Time
	}

	MeasurePoint [2]any // is: '[ 1435781430.781, "1" ]'
	QueryResult  struct {
		Metric struct {
			Name     string `json:"__name__"`
			Job      string `json:"job"`
			Instance string `json:"instance"`
		} `json:"metric"`
		Values []MeasurePoint `json:"values"`
	}
	QueryRangeResult struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string        `json:"resultType"`
			Result     []QueryResult `json:"result"`
		}
	}
	ErrorResult struct {
		Status    string `json:"status"`
		ErrorType string `json:"errorType"`
		Error     string `json:"error"`
	}
)

func (c *Client) runQuery(ctx context.Context, q promQuery) ([]QueryResult, error) {
	args := &url.Values{}
	args.Set("query", q.Expr)
	args.Set("start", fmt.Sprintf("%d", q.Start.Unix()))
	args.Set("end", fmt.Sprintf("%d", q.End.Unix()))
	args.Set("step", "14") // random for now
	slog.InfoContext(ctx, "prom request", "path", "GET "+c.Server+"/api/v1/query_range?"+args.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", c.Server+"/api/v1/query_range?"+args.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case 200:
		// happy case
	case 400:
		var payload ErrorResult
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("query error: %s", payload.Error)
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var payload QueryRangeResult
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	if payload.Status != "success" {
		return nil, fmt.Errorf("query error: %s", payload.Status)
	}
	if payload.Data.ResultType != "matrix" {
		return nil, fmt.Errorf("unexpected result type: %s", payload.Data.ResultType)
	}
	return payload.Data.Result, nil
}

func interp(v, inMin, inMax float64, outMin, outMax int) int {
	return outMin + int((v-inMin)*float64(outMax-outMin)/(inMax-inMin))
}
