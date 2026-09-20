package promgrph

import (
	"fmt"
	"html"
	"io"
)

func renderSVG(w io.Writer, g Graph) {
	const (
		leftPad     = 60
		topPad      = 30
		bottomPad   = 25
		rightPad    = 25
		legendGap   = 10
		legendWidth = 119
		legendPadX  = 10
		legendPadY  = 8
	)
	graphWidth := g.Width - leftPad - legendGap - legendWidth - rightPad
	graphHeight := g.Height - topPad - bottomPad
	graphTop := topPad
	graphBottom := g.Height - bottomPad
	legendX := leftPad + graphWidth + legendGap

	// renderPoint translates data coordinates to pixel coordinates in the graph area
	// Clamps yVal to YAxis range
	renderPoint := func(xVal, yVal int) (int, int) {
		xRange := g.XAxis.End - g.XAxis.Start
		yRange := g.YAxis.End - g.YAxis.Start

		// Clamp yVal to YAxis range
		clampedY := min(max(yVal, g.YAxis.Start), g.YAxis.End)

		x := leftPad + ((xVal - g.XAxis.Start) * graphWidth / xRange)
		y := graphTop + graphHeight - ((clampedY - g.YAxis.Start) * graphHeight / yRange)
		return x, y
	}

	fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">
  <style>
    .background { fill: #ffffff; }
    .title { font-family: sans-serif; font-size: 16px; text-anchor: middle; }
    .axis { stroke: #000000; stroke-width: 1px; }
    .tick { stroke: #cccccc; stroke-width: 1px; }
    .tick text { font-family: sans-serif; font-size: 12px; text-anchor: middle; }
    .legend { font-family: sans-serif; font-size: 12px; }
    .legend-color { width: 12px; height: 12px; }
    .line { stroke-width: 1px; }
  </style>
  <rect class="background" x="0" y="0" width="%d" height="%d"/>
  <text class="title" x="%d" y="20">%s</text>
`,
		g.Width, g.Height, g.Width, g.Height,
		g.Width, g.Height,
		g.Width/2, html.EscapeString(g.Title),
	)

	// Draw Y axis
	fmt.Fprintf(w, `<g class="axis">
    <line x1="%d" y1="%d" x2="%d" y2="%d"/>
  </g>
`, leftPad, graphTop, leftPad, graphBottom)

	// Draw Y axis ticks
	for _, tick := range g.YAxis.Ticks {
		yPos := graphTop + graphHeight - ((tick.V - g.YAxis.Start) * graphHeight / (g.YAxis.End - g.YAxis.Start))
		label := tick.Label
		if label == "" {
			label = fmt.Sprintf("%d", tick.V)
		}
		fmt.Fprintf(w, `
    <g class="tick">
      <line x1="%d" y1="%d" x2="%d" y2="%d"/>
      <text x="%d" y="%d" style="text-anchor: end">%s</text>
    </g>`,
			leftPad-5, yPos, leftPad, yPos,
			leftPad-10, yPos+4, label)
	}

	// Draw X axis
	fmt.Fprintf(w, `<g class="axis">
    <line x1="%d" y1="%d" x2="%d" y2="%d"/>
  </g>
`, leftPad, graphBottom, leftPad+graphWidth, graphBottom)

	// Draw X axis ticks
	for _, tick := range g.XAxis.Ticks {
		xPos := leftPad + ((tick.V - g.XAxis.Start) * graphWidth / (g.XAxis.End - g.XAxis.Start))
		label := tick.Label
		if label == "" {
			label = fmt.Sprintf("%d", tick.V)
		}
		fmt.Fprintf(w, `
    <g class="tick">
      <line x1="%d" y1="%d" x2="%d" y2="%d"/>
      <text x="%d" y="%d">%s</text>
    </g>`,
			xPos, graphBottom, xPos, graphBottom+5,
			xPos, graphBottom+20, label)
	}

	// Draw lines
	// Note: Stacked lines are pre-calculated in graph.go via stackLines()
	// SVG draws lines between measurements, breaking only when measurements are missing
	for _, line := range g.Lines {
		// Skip empty lines
		if len(line.Points) == 0 {
			continue
		}

		// Draw area shading first (bottom layer): 1-pixel-wide column from point down to X axis
		// Only draw when x increases to avoid overlapping measurements
		lastShadedX := -1
		for _, p := range line.Points {
			x, y := renderPoint(p[0], p[1])
			if x > lastShadedX {
				lastShadedX = x
				fmt.Fprintf(w, `  <rect x="%d" y="%d" width="1" height="%d" fill="%s" fill-opacity="0.2"/>
`,
					x, y, graphBottom-y, line.Color)
			}
		}

		// Draw the path on top of shading, connecting measurements unless there's a gap
		fmt.Fprintf(w, "  <path class=\"line\" stroke=\"%s\" stroke-width=\"1\" fill=\"none\" d=\"", line.Color)
		first := true
		for i, p := range line.Points {
			x, y := renderPoint(p[0], p[1])
			if first {
				fmt.Fprintf(w, "M %d %d", x, y)
				first = false
			} else {
				// Break line if X values differ by more than the query step
				if p[0]-line.Points[i-1][0] > max(1, g.Step) {
					fmt.Fprintf(w, " M %d %d", x, y)
				} else {
					fmt.Fprintf(w, " L %d %d", x, y)
				}
			}
		}
		fmt.Fprintf(w, "\"/>\n")

		// Also draw 1x1 pixels at each point
		for _, p := range line.Points {
			x, y := renderPoint(p[0], p[1])
			fmt.Fprintf(w, "  <rect x=\"%d\" y=\"%d\" width=\"1\" height=\"1\" fill=\"%s\"/>\n", x, y, line.Color)
		}
	}

	// Draw legend on the right
	legendBgHeight := g.Height - topPad - bottomPad
	legendY := graphTop + legendPadY
	fmt.Fprintf(w, `
  <defs>
    <clipPath id="legendClip">
      <rect x="%d" y="%d" width="%d" height="%d" rx="4" ry="4"/>
    </clipPath>
  </defs>
  <g class="legend-bg" clip-path="url(#legendClip)">
    <rect x="%d" y="%d" width="%d" height="%d" rx="4" ry="4" fill="#f5f5f5" fill-opacity="0.8"/>
`,
		legendX, graphTop, legendWidth, legendBgHeight,
		legendX, graphTop, legendWidth, legendBgHeight)
	for _, line := range g.Lines {
		if line.Label == "" {
			continue
		}
		fmt.Fprintf(w, `
    <g class="legend">
      <rect class="legend-color" x="%d" y="%d" rx="2" ry="2" fill="%s"/>
      <text x="%d" y="%d" style="text-anchor: start; dominant-baseline: central">%s</text>
    </g>`,
			legendX+10, legendY, line.Color,
			legendX+24, legendY+6,
			html.EscapeString(line.Label),
		)
		legendY += 20
	}
	fmt.Fprintf(w, "\n  </g>")
	fmt.Fprint(w, `</svg>`)
}
