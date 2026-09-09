package promgrph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
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
		resp, err := c.runQuery(ctx, expr)
		if err != nil {
			slog.ErrorContext(ctx, "query failed", "error", err)
			w.WriteHeader(500)
			w.Write([]byte("failed"))
		}

		w.Header().Set("Content-type", "image/svg")
		w.Write([]byte(fmt.Sprintf("very much todo: %#v", resp)))
	}
}

type (
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

func (c *Client) runQuery(ctx context.Context, expr string) ([]QueryResult, error) {
	now := time.Now().UTC()
	args := &url.Values{}
	args.Set("query", expr)
	args.Set("start", fmt.Sprintf("%d", now.Add(-time.Hour).Unix()))
	args.Set("end", fmt.Sprintf("%d", now.Unix()))
	args.Set("step", "14") // random
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
