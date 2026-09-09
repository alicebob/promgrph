package promgrph

import (
	"fmt"
	"io"
)

func renderSVG(w io.Writer, g Graph) error {
	const (
		leftPad   = 40
		topPad    = 30
		bottomPad = 25
		rightPad  = 10
	)
	graphWidth := g.Width - leftPad - rightPad
	graphHeight := g.Height - topPad - bottomPad
	graphTop := topPad
	graphBottom := g.Height - bottomPad

	_, err := fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">
  <style>
    .background { fill: #ffffff; }
    .title { font-family: sans-serif; font-size: 16px; text-anchor: middle; }
    .axis { stroke: #000000; stroke-width: 1px; }
    .tick { stroke: #cccccc; stroke-width: 1px; }
    .tick text { font-family: sans-serif; font-size: 12px; text-anchor: middle; }
  </style>
  <rect class="background" x="0" y="0" width="%d" height="%d"/>
  <text class="title" x="%d" y="20">%s</text>
`,
		g.Width, g.Height, g.Width, g.Height,
		g.Width, g.Height,
		g.Width/2, g.Title,
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
      <text x="%d" y="%d">%s</text>
    </g>`,
			leftPad-5, yPos, leftPad, yPos,
			leftPad-15, yPos+4, label)
		if err != nil {
			return err
		}
	}

	// Draw X axis
	_, err = fmt.Fprintf(w, `<g class="axis">
    <line x1="%d" y1="%d" x2="%d" y2="%d"/>
  </g>
`, leftPad, graphBottom, g.Width-rightPad, graphBottom)
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
	for _, line := range g.Lines {
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
			_, err = fmt.Fprintf(w, " L %d %d L %d %d Z\" fill=\"%s\" fill-opacity=\"0.2\"/>\n",
				lastX, graphBottom, firstX, graphBottom, line.Color)
			if err != nil {
				return err
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

	_, err = fmt.Fprint(w, `</svg>`)
	return err
}
