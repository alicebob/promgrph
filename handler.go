package promgrph

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

// Returns the actual graphs.
// Usage:
//
//	m.HandleFunc("GET /cpuload.png", c.MakeSVGHandler("node_cpu_seconds_total[5m]"))
//
// options:
//   - width: in pixels
//   - height: in pixels
func (c *Client) MakeSVGHandler(expr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		now := time.Now().UTC()
		width, ok := readInt(w, r, "width", 400)
		if !ok {
			return
		}
		height, ok := readInt(w, r, "height", 200)
		if !ok {
			return
		}

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

		var xTicks []AxisTick
		period := 10 * time.Minute
		for t := q.Start.Truncate(period); !t.After(q.End); t = t.Add(period) {
			if !t.Before(q.Start) {
				xTicks = append(xTicks, AxisTick{int(t.Unix()), t.Format("15:04")})
			}
		}

		g := Graph{
			Width:  width,
			Height: height,
			Title:  "My first query!",
			XAxis: Axis{
				Start: int(q.Start.Unix()),
				End:   int(q.End.Unix()),
				Label: "Time!",
				Ticks: xTicks,
			},
		}
		for _, r := range resp {
			l := Line{
				Color: "green", // FIXME
				Fill:  true,
				Label: r.Metric.Name,
			}
			for _, v := range r.Values {
				val, _ := strconv.Atoi(v[1].(string))
				x := interp(
					v[0].(float64),
					float64(q.Start.Unix()),
					float64(q.End.Unix()),
					g.XAxis.Start,
					g.XAxis.End,
				)
				y := val
				l.Points = append(l.Points, [2]int{x, y})
			}
			g.Lines = append(g.Lines, l)
		}

		yMin, yMax := computeYBounds(g.Lines)
		ticks := niceTicks(yMin, yMax, 5)

		// Adjust YAxis range to cover all ticks (niceTicks may extend beyond data)
		if len(ticks) > 0 {
			if ticks[0].V < yMin {
				yMin = ticks[0].V
			}
			if ticks[len(ticks)-1].V > yMax {
				yMax = ticks[len(ticks)-1].V
			}
		}

		g.YAxis = Axis{
			Start: yMin,
			End:   yMax,
			Label: "Numbers!",
			Ticks: ticks,
		}

		w.Header().Set("Content-Type", "image/svg+xml")
		renderSVG(w, g)
		// w.Write([]byte(fmt.Sprintf("very much todo: %#v", resp)))
	}
}

// returns the value and ok.
func readInt(w http.ResponseWriter, r *http.Request, field string, def int) (int, bool) {
	s := r.FormValue(field)
	if s == "" {
		return def, true
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		w.WriteHeader(400)
		fmt.Fprintf(w, "invalid value for argument %q", field)
		return 0, false
	}
	return n, true
}
