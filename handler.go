package promgrph

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

var colorScheme = []string{
	"#4E79A7", "#F28E2B", "#E15759", "#76B7B2", "#59A14F",
	"#EDC948", "#B07AA1", "#FF9DA7", "#9C755F", "#BAB0AC",
}

type GraphOpts struct {
	Title   string
	Stacked bool
	Legend  string // default if empty, otherwise something fixed
}

// Returns the actual graphs.
// Usage:
//
//	m.HandleFunc("GET /cpuload.png", c.MakeSVGHandler("node_cpu_seconds_total[5m]", GraphOpts{Title: "hello"}))
//
// query param options:
//   - width: in pixels
//   - height: in pixels
//   - period: how far back in time. in duration: "24h". Default 1h.
func (c *Client) MakeSVGHandler(expr string, opts GraphOpts) http.HandlerFunc {
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
		delta, ok := readDuration(w, r, "period", time.Hour)
		if !ok {
			return
		}
		q := promQuery{
			Expr:  expr,
			Start: now.Add(-delta),
			End:   now,
		}
		resp, err := c.runQuery(ctx, q)
		if err != nil {
			slog.ErrorContext(ctx, "query failed", "error", err)
			w.WriteHeader(500)
			w.Write([]byte("failed"))
		}

		var xTicks []AxisTick
		period, format := nicePeriod(delta)
		for t := q.Start.Truncate(period); !t.After(q.End); t = t.Add(period) {
			if !t.Before(q.Start) {
				xTicks = append(xTicks, AxisTick{int(t.Unix()), t.Format(format)})
			}
		}

		g := Graph{
			Width:   width,
			Height:  height,
			Title:   opts.Title,
			Stacked: opts.Stacked,
			XAxis: Axis{
				Start: int(q.Start.Unix()),
				End:   int(q.End.Unix()),
				Label: "Time!",
				Ticks: xTicks,
			},
		}
		for i, r := range resp {
			l := Line{
				Color: colorScheme[i%len(colorScheme)],
				Fill:  true,
				Label: makeLabel(opts.Legend, r.Metric),
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

// see readInt()
func readDuration(w http.ResponseWriter, r *http.Request, field string, def time.Duration) (time.Duration, bool) {
	s := r.FormValue(field)
	if s == "" {
		return def, true
	}
	n, err := time.ParseDuration(s)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "invalid value for argument %q", field)
		return 0, false
	}
	return n, true
}

func makeLabel(fixed string, m Metric) string {
	if fixed != "" {
		return fixed
	}
	if name := m.Name; name != "" {
		return name
	}
	return m.Job
}

func nicePeriod(d time.Duration) (time.Duration, string) {
	// FIXME: revisit this, and do something with the width (which we know kinda)
	switch {
	case d <= time.Minute:
		return 10 * time.Second, "15:04:05"
	case d <= 10*time.Minute:
		return time.Minute, "15:04"
	case d <= time.Hour:
		return 10 * time.Minute, "15:04"
	case d <= 24*time.Hour:
		return 3 * time.Hour, "15:04"
	case d <= 7*24*time.Hour:
		return 24 * time.Hour, "01-02T15:04"
	default:
		return 7 * 24 * time.Hour, "01-02T15:04"
	}
}
