package promgrph

import (
	"testing"

	"github.com/shoenig/test/must"
)

func TestFormatValue(t *testing.T) {
	tests := []struct {
		v    float64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{999, "999"},
		{1000, "1.0K"},
		{1500, "1.5K"},
		{1000000, "1.0M"},
		{50000000, "50.0M"},
		{1000000000, "1.0B"},
		{-5000000, "-5.0M"},
		{0.5, "0.5"},
		{0.123, "0.123"},
		{0.001, "0.001"},
		{1.5, "1.5"},
		{1234.567, "1.2K"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatValue(tt.v)
			must.Eq(t, tt.want, got)
		})
	}
}

func TestNiceTicks(t *testing.T) {
	tests := []struct {
		name       string
		vmin, vmax float64
		maxTicks   int
		want       []float64 // just check the V values
	}{
		{
			name:     "zero range",
			vmin:     5,
			vmax:     5,
			maxTicks: 5,
			want:     []float64{4, 5, 6},
		},
		{
			name:     "small range 0-1",
			vmin:     0,
			vmax:     1,
			maxTicks: 5,
			want:     []float64{0, 0.2, 0.4, 0.6, 0.8, 1},
		},
		{
			name:     "0-5",
			vmin:     0,
			vmax:     5,
			maxTicks: 5,
			want:     []float64{0, 1, 2, 3, 4, 5},
		},
		{
			name:     "0-50M",
			vmin:     0,
			vmax:     50_000_000,
			maxTicks: 5,
			want:     []float64{0, 10_000_000, 20_000_000, 30_000_000, 40_000_000, 50_000_000},
		},
		{
			name:     "17-42",
			vmin:     17,
			vmax:     42,
			maxTicks: 5,
			want:     []float64{15, 20, 25, 30, 35, 40, 45},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticks := niceTicks(tt.vmin, tt.vmax, tt.maxTicks)
			var got []float64
			for _, tick := range ticks {
				got = append(got, tick.V)
			}
			must.Eq(t, tt.want, got)
		})
	}
}

func TestComputeYBounds(t *testing.T) {
	pt := func(x, y float64) [2]float64 { return [2]float64{x, y} }

	tests := []struct {
		name    string
		lines   []Line
		wantMin float64
		wantMax float64
	}{
		{
			name:    "empty lines",
			lines:   nil,
			wantMin: 0,
			wantMax: 1,
		},
		{
			name:    "single point",
			lines:   []Line{{Points: Section{pt(0, 5)}}},
			wantMin: 4.5,
			wantMax: 5.5,
		},
		{
			name:    "all same values",
			lines:   []Line{{Points: Section{pt(0, 3), pt(1, 3), pt(2, 3)}}},
			wantMin: 2.7,
			wantMax: 3.3,
		},
		{
			name:    "mixed values",
			lines:   []Line{{Points: Section{pt(0, 10), pt(1, 0), pt(2, 5)}}},
			wantMin: 0,
			wantMax: 10,
		},
		{
			name:    "negative values",
			lines:   []Line{{Points: Section{pt(0, -5), pt(1, -10), pt(2, 0)}}},
			wantMin: -10,
			wantMax: 0,
		},
		{
			name:    "multiple lines",
			lines:   []Line{{Points: Section{pt(0, 1), pt(1, 2)}}, {Points: Section{pt(0, 3), pt(1, 4)}}},
			wantMin: 1,
			wantMax: 4,
		},
		{
			name:    "float values",
			lines:   []Line{{Points: Section{pt(0, 0.5), pt(1, 0.75), pt(2, 0.25)}}},
			wantMin: 0.25,
			wantMax: 0.75,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()
			gotMin, gotMax := computeYBounds(tt.lines)
			// Use approximate comparison for float values
			if gotMin != tt.wantMin || gotMax != tt.wantMax {
				t.Errorf("computeYBounds() = (%v, %v), want (%v, %v)", gotMin, gotMax, tt.wantMin, tt.wantMax)
			}
		})
	}
}
