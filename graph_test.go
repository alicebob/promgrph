package promgrph

import (
	"testing"

	"github.com/shoenig/test/must"
)

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
