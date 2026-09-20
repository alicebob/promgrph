package promgrph

import (
	"fmt"
	"math"
)

type (
	Graph struct {
		Width   int
		Height  int
		Title   string
		YAxis   Axis
		XAxis   Axis
		Stacked bool
		Step    int // query step in seconds, for gap detection
		Lines   []Line
	}
	Axis struct {
		Label string
		Start int
		End   int
		Ticks []AxisTick
	}
	AxisTick struct {
		V     int
		Label string
	}
	Line struct {
		Color  string
		Fill   bool // fill in the area under the line?
		Label  string
		Points Section
	}
	Section [][2]int
)

// formatValue formats an integer as a compact string.
// Uses suffixes K, M, B, T for large numbers.
func formatValue(v int) string {
	if v < 0 {
		return "-" + formatValue(-v)
	}

	const (
		K = 1000
		M = 1000 * K
		B = 1000 * M
		T = 1000 * B
	)

	switch {
	case v >= T:
		return fmt.Sprintf("%.1fT", float64(v)/T)
	case v >= B:
		return fmt.Sprintf("%.1fB", float64(v)/B)
	case v >= M:
		return fmt.Sprintf("%.1fM", float64(v)/M)
	case v >= K:
		return fmt.Sprintf("%.1fK", float64(v)/K)
	default:
		return fmt.Sprintf("%d", v)
	}
}

// niceNum returns the nearest "nice" number (1, 2, 5, 10, 20, 50, ...) for axis ticks.
// Part of Wilkinson's algorithm (Leland Wilkinson, "An Algorithm for Optimal
// Tick Marks on Continuous scales", 1999).
// If round is true, rounds to nearest; if false, rounds down.
func niceNum(n float64, round bool) float64 {
	if n == 0 {
		return 0
	}
	exponent := int(math.Floor(math.Log10(math.Abs(n))))
	base := math.Pow10(exponent)
	fraction := n / base

	if round {
		if fraction < 1.5 {
			return 1 * base
		} else if fraction < 3 {
			return 2 * base
		} else if fraction < 7 {
			return 5 * base
		}
		return 10 * base
	}
	// floor
	if fraction <= 1 {
		return 1 * base
	} else if fraction <= 2 {
		return 2 * base
	} else if fraction <= 5 {
		return 5 * base
	}
	return 10 * base
}

// niceTicks generates nice Y-axis ticks for a given range.
// Returns round-numbered ticks, evenly spaced, covering [vmin, vmax].
func niceTicks(vmin, vmax int, maxTicks int) []AxisTick {
	if vmin == vmax {
		return []AxisTick{
			{V: vmin - 1, Label: formatValue(vmin - 1)},
			{V: vmin, Label: formatValue(vmin)},
			{V: vmin + 1, Label: formatValue(vmin + 1)},
		}
	}

	minVal := float64(vmin)
	maxVal := float64(vmax)
	rng := maxVal - minVal

	// Target ~5-10 ticks
	step := niceNum(rng/float64(maxTicks-1), true)
	if step == 0 {
		step = niceNum(rng, false)
	}
	// Ensure step is at least 1 for integer values
	if step < 1 {
		step = 1
	}

	niceMin := math.Floor(minVal/step) * step
	niceMax := math.Ceil(maxVal/step) * step

	var ticks []AxisTick
	for v := niceMin; v <= niceMax+0.5; v += step {
		ticks = append(ticks, AxisTick{V: int(math.Round(v)), Label: formatValue(int(math.Round(v)))})
	}

	return ticks
}

func interp(v, inMin, inMax float64, outMin, outMax int) int {
	return outMin + int((v-inMin)*float64(outMax-outMin)/(inMax-inMin))
}

// stackLines pre-calculates stacked line values.
// It modifies the lines in place, adding each line's Y values to the previous lines.
// All lines must have the same number of points at the same X positions.
func stackLines(lines []Line) {
	for i := 1; i < len(lines); i++ {
		prev := &lines[i-1]
		curr := &lines[i]
		// Stack current line on top of previous line
		for j := range curr.Points {
			if j < len(prev.Points) {
				curr.Points[j][1] += prev.Points[j][1]
			}
		}
	}
}

// computeYBounds returns the minimum and maximum Y values from the lines.
// If all values are the same, it adds padding (±1) to avoid division by zero.
// If there are no points, it returns (0, 1).
func computeYBounds(lines []Line) (yMin, yMax int) {
	found := false
	for _, l := range lines {
		for _, p := range l.Points {
			if !found {
				yMin, yMax = p[1], p[1]
				found = true
			} else {
				yMin = min(yMin, p[1])
				yMax = max(yMax, p[1])
			}
		}
	}
	if !found {
		return 0, 1
	}
	if yMin == yMax {
		return yMin - 1, yMax + 1
	}
	return yMin, yMax
}
