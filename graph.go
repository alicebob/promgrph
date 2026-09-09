package promgrph

type (
	Graph struct {
		Width  int
		Height int
		Title  string
		YAxis  Axis
		XAxis  Axis
		Lines  []Line
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
		Points [][2]int
	}
)

func interp(v, inMin, inMax float64, outMin, outMax int) int {
	return outMin + int((v-inMin)*float64(outMax-outMin)/(inMax-inMin))
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
