package promgrph

import (
	"testing"

	"github.com/shoenig/test/must"
)

func TestFormatValue(t *testing.T) {
	tests := []struct {
		v    int
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
		vmin, vmax int
		maxTicks   int
		want       []int // just check the V values
	}{
		{
			name:     "zero range",
			vmin:     5,
			vmax:     5,
			maxTicks: 5,
			want:     []int{4, 5, 6},
		},
		{
			name:     "small range 0-1",
			vmin:     0,
			vmax:     1,
			maxTicks: 5,
			want:     []int{0, 1},
		},
		{
			name:     "0-5",
			vmin:     0,
			vmax:     5,
			maxTicks: 5,
			want:     []int{0, 1, 2, 3, 4, 5},
		},
		{
			name:     "0-50M",
			vmin:     0,
			vmax:     50_000_000,
			maxTicks: 5,
			want:     []int{0, 10_000_000, 20_000_000, 30_000_000, 40_000_000, 50_000_000},
		},
		{
			name:     "17-42",
			vmin:     17,
			vmax:     42,
			maxTicks: 5,
			want:     []int{15, 20, 25, 30, 35, 40, 45},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticks := niceTicks(tt.vmin, tt.vmax, tt.maxTicks)
			var got []int
			for _, tick := range ticks {
				got = append(got, tick.V)
			}
			must.Eq(t, tt.want, got)
		})
	}
}

func TestComputeYBounds(t *testing.T) {
	tests := []struct {
		name    string
		lines   []Line
		wantMin int
		wantMax int
	}{
		{
			name:    "empty lines",
			lines:   nil,
			wantMin: 0,
			wantMax: 1,
		},
		{
			name:    "single point",
			lines:   []Line{{Points: [][2]int{{0, 5}}}},
			wantMin: 4,
			wantMax: 6,
		},
		{
			name:    "all same values",
			lines:   []Line{{Points: [][2]int{{0, 3}, {1, 3}, {2, 3}}}},
			wantMin: 2,
			wantMax: 4,
		},
		{
			name:    "mixed values",
			lines:   []Line{{Points: [][2]int{{0, 10}, {1, 0}, {2, 5}}}},
			wantMin: 0,
			wantMax: 10,
		},
		{
			name:    "negative values",
			lines:   []Line{{Points: [][2]int{{0, -5}, {1, -10}, {2, 0}}}},
			wantMin: -10,
			wantMax: 0,
		},
		{
			name:    "multiple lines",
			lines:   []Line{{Points: [][2]int{{0, 1}, {1, 2}}}, {Points: [][2]int{{0, 3}, {1, 4}}}},
			wantMin: 1,
			wantMax: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()
			gotMin, gotMax := computeYBounds(tt.lines)
			must.Eq(t, tt.wantMin, gotMin)
			must.Eq(t, tt.wantMax, gotMax)
		})
	}
}
