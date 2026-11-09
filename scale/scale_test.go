package scale_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/ingvarch/pulse/scale"
)

func TestNiceTicks_CPUPercent(t *testing.T) {
	got := scale.NiceTicks(0, 100, 5)
	want := []float64{0, 20, 40, 60, 80, 100}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("NiceTicks(0,100,5) = %v, want %v", got, want)
	}
}

func TestNiceTicks_SmallRange(t *testing.T) {
	got := scale.NiceTicks(0, 1, 5)
	if len(got) < 2 {
		t.Fatalf("expected at least 2 ticks, got %v", got)
	}
	if got[0] > 0 || got[len(got)-1] < 1 {
		t.Fatalf("ticks %v must cover [0,1]", got)
	}
}

func TestFormatTick_Integer(t *testing.T) {
	if got := scale.FormatTick(40); got != "40" {
		t.Fatalf("FormatTick(40) = %q, want %q", got, "40")
	}
}

func TestFormatTick_RoundsLongFloat(t *testing.T) {
	if got := scale.FormatTick(88.66666666666666); got != "88.67" {
		t.Fatalf("FormatTick(88.666..) = %q, want %q", got, "88.67")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1 KB"},
		{1536, "1.5 KB"},
		{1048576, "1 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1 GB"},
		{-1048576, "-1 MB"},
		{-1536, "-1.5 KB"},
	}
	for _, tc := range tests {
		if got := scale.FormatBytes(tc.input); got != tc.want {
			t.Errorf("FormatBytes(%v) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestFormatBytesRate(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0, "0 B/s"},
		{1024, "1 KB/s"},
		{104857600, "100 MB/s"},
		{-104857600, "-100 MB/s"},
	}
	for _, tc := range tests {
		if got := scale.FormatBytesRate(tc.input); got != tc.want {
			t.Errorf("FormatBytesRate(%v) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestFormatBytesRateAbs(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0, "0 B/s"},
		{52428800, "50 MB/s"},
		{-52428800, "50 MB/s"},
	}
	for _, tc := range tests {
		if got := scale.FormatBytesRateAbs(tc.input); got != tc.want {
			t.Errorf("FormatBytesRateAbs(%v) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestBytesRateFormatter(t *testing.T) {
	fmtAbs := scale.BytesRateFormatter(true)
	if got := fmtAbs(-1048576); got != "1 MB/s" {
		t.Errorf("BytesRateFormatter(true)(-1MB) = %q, want %q", got, "1 MB/s")
	}

	fmtSigned := scale.BytesRateFormatter(false)
	if got := fmtSigned(-1048576); got != "-1 MB/s" {
		t.Errorf("BytesRateFormatter(false)(-1MB) = %q, want %q", got, "-1 MB/s")
	}
}

func TestNiceRange_RoundsOutToTicks(t *testing.T) {
	lo, hi := scale.NiceRange(0, 730, 3)
	if lo != 0 || hi < 730 || hi >= 1000 {
		t.Fatalf("NiceRange(0,730,3) = %v..%v, want 0..(750 or 800)", lo, hi)
	}
}

func TestNiceRange_BoundsAreTicksOfTheirOwnRange(t *testing.T) {
	for _, tc := range []struct{ min, max float64 }{
		{0, 730}, {0, 1}, {-3, 40}, {-730, 5}, {0, 0.37}, {0, 12345}, {-1e6, 2.5e6},
		{0, 2.6e15}, {-4e14, 1.05e15}, {0, 2.6e29},
	} {
		for n := 2; n <= 5; n++ {
			lo, hi := scale.NiceRange(tc.min, tc.max, n)
			if lo > tc.min || hi < tc.max {
				t.Fatalf("NiceRange(%v,%v,%d) = %v..%v does not cover the input", tc.min, tc.max, n, lo, hi)
			}
			ticks := scale.NiceTicks(lo, hi, n)
			if ticks[0] != lo || ticks[len(ticks)-1] != hi {
				t.Fatalf("NiceRange(%v,%v,%d) = %v..%v, but its ticks are %v", tc.min, tc.max, n, lo, hi, ticks)
			}
		}
	}
}

func TestNiceRange_PicksTheSmallestStableRange(t *testing.T) {
	for _, tc := range []struct {
		min, max, wantLo, wantHi float64
		ticks                    int
	}{
		{0, 426, 0, 500, 2},
		{-90, 232, -250, 250, 2},
		{0, 2.6, 0, 4, 2},
		{0, 2.6e15, 0, 4e15, 2},
	} {
		lo, hi := scale.NiceRange(tc.min, tc.max, tc.ticks)
		if lo != tc.wantLo || hi != tc.wantHi {
			t.Errorf("NiceRange(%v,%v,%d) = %v..%v, want %v..%v", tc.min, tc.max, tc.ticks, lo, hi, tc.wantLo, tc.wantHi)
		}
	}
}

func TestNiceTicks_HugeBoundsTerminate(t *testing.T) {
	ticks := scale.NiceTicks(0, 1.7e308, 5)
	if len(ticks) == 0 || len(ticks) > 100 {
		t.Fatalf("NiceTicks(0,1.7e308,5) returned %d ticks", len(ticks))
	}
	lo, hi := scale.NiceRange(0, 1.7e308, 5)
	if math.IsInf(lo, 0) || math.IsInf(hi, 0) || hi < 1.7e308 {
		t.Fatalf("NiceRange(0,1.7e308,5) = %v..%v", lo, hi)
	}
}
