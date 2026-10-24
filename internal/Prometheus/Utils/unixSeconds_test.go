package Utils

import (
	"math"
	"testing"
)

// The values below are the shape the controller actually returns (see the
// device list fixtures in internal/Omada/Model/Devices); MACs, names and IPs
// are anonymised, the timestamps are kept because that is what this test is
// about.
func TestUnixSeconds(t *testing.T) {
	tests := []struct {
		name     string
		lastSeen float64
		want     float64
	}{
		{
			name:     "milliseconds as returned by the Open API device list",
			lastSeen: 1792833739702,
			want:     1792833739.702,
		},
		{
			name:     "milliseconds with sub-millisecond precision",
			lastSeen: 1792833739702.75,
			want:     1792833739.70275,
		},
		{
			name:     "value at the threshold is milliseconds",
			lastSeen: millisecondThreshold,
			want:     millisecondThreshold / 1000,
		},
		{
			name:     "seconds are passed through unchanged",
			lastSeen: 1792833739,
			want:     1792833739,
		},
		{
			name:     "zero is left alone",
			lastSeen: 0,
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UnixSeconds(tt.lastSeen)
			if math.Abs(got-tt.want) > 1e-6 {
				t.Fatalf("UnixSeconds(%v) = %v, want %v", tt.lastSeen, got, tt.want)
			}
		})
	}
}
