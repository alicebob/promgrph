package promgrph

import (
	"fmt"
	"html"
	"io"
)

func renderSVG(w io.Writer, g Graph) error {
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

	_, err := fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">
  <style>
    .background { fill: #ffffff; }
    .title { font-family: sans-serif; font-size: 16px; text-anchor: middle; }
    .axis { stroke: #000000; stroke-width: 1px; }
    .tick { stroke: #cccccc; stroke-width: 1px; }
    .tick text { font-family: sans-serif; font-size: 12px; text-anchor: middle; }
    .legend { font-family: sans-serif; font-size: 12px; }
    .legend-color { width: 12px; height: 12px; }
  </style>
  <rect class="background" x="0" y="0" width="%d" height="%d"/>
  <text class="title" x="%d" y="20">%s</text>
`,
		g.Width, g.Height, g.Width, g.Height,
		g.Width, g.Height,
		g.Width/2, html.EscapeString(g.Title),
	)
	if err != nil {
		return err
	}

	// Draw Y axis
	_, err = fmt.Fprintf(w, `<g class="axis">
    <line x1="%d" y1="%d" x2="%d" y2="%d"/>
  </g>
`, leftPad, graphTop, leftPad, graphBottom)
	if err != nil {
		return err
	}

	// Draw Y axis ticks
	for _, tick := range g.YAxis.Ticks {
		yPos := graphTop + graphHeight - ((tick.V - g.YAxis.Start) * graphHeight / (g.YAxis.End - g.YAxis.Start))
		label := tick.Label
		if label == "" {
			label = fmt.Sprintf("%d", tick.V)
		}
		_, err = fmt.Fprintf(w, `
    <g class="tick">
      <line x1="%d" y1="%d" x2="%d" y2="%d"/>
      <text x="%d" y="%d" style="text-anchor: end">%s</text>
    </g>`,
			leftPad-5, yPos, leftPad, yPos,
			leftPad-10, yPos+4, label)
		if err != nil {
			return err
		}
	}

	// Draw X axis
	_, err = fmt.Fprintf(w, `<g class="axis">
    <line x1="%d" y1="%d" x2="%d" y2="%d"/>
  </g>
`, leftPad, graphBottom, leftPad+graphWidth, graphBottom)
	if err != nil {
		return err
	}

	// Draw X axis ticks
	for _, tick := range g.XAxis.Ticks {
		xPos := leftPad + ((tick.V - g.XAxis.Start) * graphWidth / (g.XAxis.End - g.XAxis.Start))
		label := tick.Label
		if label == "" {
			label = fmt.Sprintf("%d", tick.V)
		}
		_, err = fmt.Fprintf(w, `
    <g class="tick">
      <line x1="%d" y1="%d" x2="%d" y2="%d"/>
      <text x="%d" y="%d">%s</text>
    </g>`,
			xPos, graphBottom, xPos, graphBottom+5,
			xPos, graphBottom+20, label)
		if err != nil {
			return err
		}
	}

	// Draw lines
	xRange := g.XAxis.End - g.XAxis.Start
	yRange := g.YAxis.End - g.YAxis.Start

	// Stack line points if requested (cumulative Y values)
	if g.Stacked {
		for i := 1; i < len(g.Lines); i++ {
			prev := &g.Lines[i-1]
			curr := &g.Lines[i]
			if len(prev.Points) == len(curr.Points) {
				for j := range curr.Points {
					curr.Points[j][1] += prev.Points[j][1]
				}
			}
		}
	}

	for i, line := range g.Lines {
		// Draw fill area if enabled
		if line.Fill && len(line.Points) > 0 {
			firstX := leftPad + ((line.Points[0][0] - g.XAxis.Start) * graphWidth / xRange)
			firstY := graphTop + graphHeight - ((line.Points[0][1] - g.YAxis.Start) * graphHeight / yRange)
			_, err = fmt.Fprintf(w, "\n  <path d=\"M %d %d", firstX, firstY)
			if err != nil {
				return err
			}
			for _, p := range line.Points[1:] {
				x := leftPad + ((p[0] - g.XAxis.Start) * graphWidth / xRange)
				y := graphTop + graphHeight - ((p[1] - g.YAxis.Start) * graphHeight / yRange)
				_, err = fmt.Fprintf(w, " L %d %d", x, y)
				if err != nil {
					return err
				}
			}
			lastX := leftPad + ((line.Points[len(line.Points)-1][0] - g.XAxis.Start) * graphWidth / xRange)
			// For stacked lines, fill down to previous line; for non-stacked or first line, fill to graph bottom
			if g.Stacked && i > 0 {
				prevLine := &g.Lines[i-1]
				if len(prevLine.Points) > 0 {
					prevFirstY := graphTop + graphHeight - ((prevLine.Points[0][1] - g.YAxis.Start) * graphHeight / yRange)
					prevLastY := graphTop + graphHeight - ((prevLine.Points[len(prevLine.Points)-1][1] - g.YAxis.Start) * graphHeight / yRange)
					_, err = fmt.Fprintf(w, " L %d %d", lastX, prevLastY)
					if err != nil {
						return err
					}
					// Reverse through prev line points
					for j := len(prevLine.Points) - 2; j >= 0; j-- {
						px := leftPad + ((prevLine.Points[j][0] - g.XAxis.Start) * graphWidth / xRange)
						py := graphTop + graphHeight - ((prevLine.Points[j][1] - g.YAxis.Start) * graphHeight / yRange)
						_, err = fmt.Fprintf(w, " L %d %d", px, py)
						if err != nil {
							return err
						}
					}
					_, err = fmt.Fprintf(w, " L %d %d Z\" fill=\"%s\" fill-opacity=\"0.2\"/>\n", firstX, prevFirstY, line.Color)
					if err != nil {
						return err
					}
				}
			} else {
				// Non-stacked or first line: fill to graph bottom
				_, err = fmt.Fprintf(w, " L %d %d L %d %d Z\" fill=\"%s\" fill-opacity=\"0.2\"/>\n",
					lastX, graphBottom, firstX, graphBottom, line.Color)
				if err != nil {
					return err
				}
			}
		}

		// Draw line
		_, err = fmt.Fprintf(w, "  <path class=\"line\" stroke=\"%s\" stroke-width=\"2\" fill=\"none\" d=\"", line.Color)
		if err != nil {
			return err
		}
		for i, p := range line.Points {
			x := leftPad + ((p[0] - g.XAxis.Start) * graphWidth / xRange)
			y := graphTop + graphHeight - ((p[1] - g.YAxis.Start) * graphHeight / yRange)
			if i == 0 {
				_, err = fmt.Fprintf(w, "M %d %d", x, y)
			} else {
				_, err = fmt.Fprintf(w, " L %d %d", x, y)
			}
			if err != nil {
				return err
			}
		}
		_, err = fmt.Fprint(w, "\"/>\n")
		if err != nil {
			return err
		}
	}

	// Draw legend on the right
	legendBgHeight := g.Height - topPad - bottomPad
	legendY := graphTop + legendPadY
	_, err = fmt.Fprintf(w, `
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
	if err != nil {
		return err
	}
	for _, line := range g.Lines {
		if line.Label == "" {
			continue
		}
		_, err = fmt.Fprintf(w, `
    <g class="legend">
      <rect class="legend-color" x="%d" y="%d" rx="2" ry="2" fill="%s"/>
      <text x="%d" y="%d" style="text-anchor: start; dominant-baseline: central">%s</text>
    </g>`,
			legendX+10, legendY, line.Color,
			legendX+24, legendY+6, line.Label)
		if err != nil {
			return err
		}
		legendY += 20
	}
	_, err = fmt.Fprintf(w, "\n  </g>")
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(w, `</svg>`)
	return err
}
