package promgrph

import (
	"fmt"
	"math"
)

type (
	Graph struct {
		Width     int
		Height    int
		Title     string
		YAxis     Axis
		XAxis     Axis
		Stacked   bool
		FixedYMin *float64
		FixedYMax *float64
		Step      int // query step in seconds, for gap detection
		Lines     []Line
		Fill      int // 0..100, alpha value of any area fill.
	}
	Axis struct {
		Label string
		Start float64
		End   float64
		Ticks []AxisTick
	}
	AxisTick struct {
		V     float64
		Label string
	}
	Line struct {
		Color  string
		Label  string
		Points Section
	}
	Section [][2]float64
)

// formatValue formats a float64 as a compact string.
// Uses suffixes K, M, B, T for large numbers.
// For small numbers, formats with reasonable precision.
func formatValue(v float64) string {
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
		return fmt.Sprintf("%.1fT", v/T)
	case v >= B:
		return fmt.Sprintf("%.1fB", v/B)
	case v >= M:
		return fmt.Sprintf("%.1fM", v/M)
	case v >= K:
		return fmt.Sprintf("%.1fK", v/K)
	default:
		// For small numbers, format nicely
		// If it's a whole number, don't show decimal
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		// For numbers < 1, show up to 3 decimal places
		if v < 1 {
			return fmt.Sprintf("%.3g", v)
		}
		// For numbers between 1 and 1000, show up to 2 decimal places
		return fmt.Sprintf("%.2g", v)
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
func niceTicks(vmin, vmax float64, maxTicks int) []AxisTick {
	if vmin == vmax {
		return []AxisTick{
			{V: vmin - 1, Label: formatValue(vmin - 1)},
			{V: vmin, Label: formatValue(vmin)},
			{V: vmin + 1, Label: formatValue(vmin + 1)},
		}
	}

	minVal := vmin
	maxVal := vmax
	rng := maxVal - minVal

	// Target ~5-10 ticks
	step := niceNum(rng/float64(maxTicks-1), true)
	if step == 0 {
		step = niceNum(rng, false)
	}
	// Ensure step is at least a reasonable minimum for float values
	if step < 0.001 {
		step = 0.001
	}

	niceMin := math.Floor(minVal/step) * step
	niceMax := math.Ceil(maxVal/step) * step

	var ticks []AxisTick
	for v := niceMin; v <= niceMax+step*0.5; v += step {
		rounded := math.Round(v*1000) / 1000 // Round to 3 decimal places to avoid floating point artifacts
		ticks = append(ticks, AxisTick{V: rounded, Label: formatValue(rounded)})
	}

	return ticks
}

func interp(v, inMin, inMax float64, outMin, outMax float64) float64 {
	return outMin + (v-inMin)*float64(outMax-outMin)/(inMax-inMin)
}

// stackLines pre-calculates stacked line values.
// It modifies the lines in place, adding each line's Y values to the next lines.
// After stacking, line 0 will have the tallest (cumulative) values.
// All lines must have the same number of points at the same X positions.
func stackLines(lines []Line) {
	// Work from the end: each line accumulates all lines after it
	for i := len(lines) - 2; i >= 0; i-- {
		curr := &lines[i]
		next := &lines[i+1]
		for j := range curr.Points {
			if j < len(next.Points) {
				curr.Points[j][1] += next.Points[j][1]
			}
		}
	}
}

// computeYBounds returns the minimum and maximum Y values from the lines.
// If all values are the same, it adds padding (±epsilon) to avoid division by zero.
// If there are no points, it returns (0, 1).
func computeYBounds(lines []Line) (yMin, yMax float64) {
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
		// Add padding based on the value magnitude
		padding := 1.0
		if yMin != 0 {
			padding = math.Abs(yMin) * 0.1
			if padding < 0.01 {
				padding = 0.01
			}
		}
		return yMin - padding, yMax + padding
	}
	return yMin, yMax
}
