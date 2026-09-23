package promgrph

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var colorScheme = []string{
	"#4E79A7", "#F28E2B", "#E15759", "#76B7B2", "#59A14F",
	"#EDC948", "#B07AA1", "#FF9DA7", "#9C755F", "#BAB0AC",
}

type GraphOpts struct {
	Title     string
	Stacked   bool
	Legend    string // default if empty, otherwise something fixed
	Fill      int    // 0..100% fill in the area under the graph.
	FixedYMin *int
	FixedYMax *int
}

// Returns the actual graphs.
// Usage:
//
//	m.HandleFunc("GET /cpuload.png", c.MakeSVGHandler("node_cpu_seconds_total[5m]", GraphOpts{Title: "hello"}))
//
// query param options:
//   - width: in pixels
//   - height: in pixels
//   - period: how far back in time. in duration: "24h". Default 10m.
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
		delta, ok := readDuration(w, r, "period", 10*time.Minute)
		if !ok {
			return
		}
		// Calculate step: aim for ~1 data point per pixel of graph width
		// Graph width = width - leftPad(60) - legendGap(10) - legendWidth(119) - rightPad(25)
		graphWidth := width - 60 - 10 - 119 - 25
		if graphWidth <= 0 {
			graphWidth = 400
		}
		step := max(time.Second, delta/time.Duration(graphWidth))
		q := promQuery{
			Expr:  expr,
			Start: now.Add(-delta),
			End:   now,
			Step:  step,
		}
		g, err := makeGraph(ctx, c, q, opts, width, height)
		if err != nil {
			slog.ErrorContext(ctx, "query failed", "error", err)
			w.Header().Set("Content-Type", "image/svg+xml")
			// w.WriteHeader(500)
			w.Write(errorSVG(width, height))
			return
		}

		var buf bytes.Buffer
		renderSVG(&buf, g)
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
		w.Write(buf.Bytes())
	}
}

func errorSVG(width, height int) []byte {
	msg := fmt.Sprintf("<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"%d\" height=\"%d\" viewBox=\"0 0 %d %d\">"+
		"<text x=\"%d\" y=\"%d\" text-anchor=\"middle\" dominant-baseline=\"middle\">Graph unavailable right now</text>"+
		"</svg>", width, height, width, height, width/2, height/2)
	return []byte(msg)
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

func makeGraph(ctx context.Context, c *Client, q promQuery, opts GraphOpts, width, height int) (Graph, error) {
	resp, err := c.runQuery(ctx, q)
	if err != nil {
		return Graph{}, err
	}

	var xTicks []AxisTick
	period, format := nicePeriod(q.End.Sub(q.Start))
	for t := q.Start.Truncate(period); !t.After(q.End); t = t.Add(period) {
		if !t.Before(q.Start) {
			xTicks = append(xTicks, AxisTick{V: float64(t.Unix()), Label: t.Format(format)})
		}
	}

	g := Graph{
		Width:     width,
		Height:    height,
		Title:     opts.Title,
		Stacked:   opts.Stacked,
		Step:      int(q.Step.Seconds()),
		Fill:      opts.Fill,
		XAxis: Axis{
			Start: float64(q.Start.Unix()),
			End:   float64(q.End.Unix()),
			Label: "Time!",
			Ticks: xTicks,
		},
	}
	// Convert FixedYMin and FixedYMax from *int to *float64
	if opts.FixedYMin != nil {
		fymin := float64(*opts.FixedYMin)
		g.FixedYMin = &fymin
	}
	if opts.FixedYMax != nil {
		fymax := float64(*opts.FixedYMax)
		g.FixedYMax = &fymax
	}
	for i, r := range resp {
		l := Line{
			Color: colorScheme[i%len(colorScheme)],
			Label: makeLabel(opts.Legend, r.Metric),
		}
		var s Section
		for _, v := range r.Values {
			val, err := strconv.ParseFloat(v[1].(string), 64)
			if err != nil {
				// Fallback to 0 if parsing fails
				val = 0
			}
			x := interp(
				v[0].(float64),
				float64(q.Start.Unix()),
				float64(q.End.Unix()),
				float64(g.XAxis.Start),
				float64(g.XAxis.End),
			)
			s = append(s, [2]float64{x, val})
		}
		l.Points = s
		g.Lines = append(g.Lines, l)
	}

	// Pre-calculate stacked lines if requested
	if opts.Stacked {
		stackLines(g.Lines)
	}

	yMin, yMax := computeYBounds(g.Lines)

	if g.FixedYMin != nil {
		yMin = *g.FixedYMin
	}
	if g.FixedYMax != nil {
		yMax = *g.FixedYMax
	}

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

	return g, nil
}

func makeLabel(fixed string, m Metric) string {
	if fixed != "" {
		t, err := template.New("label").Parse(fixed)
		if err != nil {
			return "[broken template]"
		}
		buf := &strings.Builder{}
		if err := t.Execute(buf, m); err != nil {
			return fmt.Sprintf("broken template: %s", err)
		}
		return buf.String()
	}
	if name := m["__name__"]; name != "" {
		return name
	}
	return m["job"]
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
