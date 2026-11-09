package scale_test

import (
	"math"
	"reflect"
	"slices"
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

func TestLogTicks_PowersOfTen(t *testing.T) {
	got := scale.LogTicks(1, 1000, 5)
	want := []float64{1, 10, 100, 1000}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LogTicks(1,1000,5) = %v, want %v", got, want)
	}
}

func TestLogTicks_BoundsBetweenPowers(t *testing.T) {
	got := scale.LogTicks(3, 5000, 5)
	want := []float64{10, 100, 1000}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LogTicks(3,5000,5) = %v, want %v", got, want)
	}
}

func TestLogTicks_NarrowRangeAddsTwoAndFive(t *testing.T) {
	tests := []struct {
		min, max float64
		want     []float64
	}{
		{2, 8, []float64{2, 5}},
		{10, 30, []float64{10, 20}},
		{0.5, 1.5, []float64{0.5, 1}},
		{1, 10, []float64{1, 10}},
	}
	for _, tc := range tests {
		if got := scale.LogTicks(tc.min, tc.max, 5); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("LogTicks(%v,%v,5) = %v, want %v", tc.min, tc.max, got, tc.want)
		}
	}
}

func TestLogTicks_NarrowRangeWithoutNiceValueKeepsBounds(t *testing.T) {
	got := scale.LogTicks(3, 4, 5)
	want := []float64{3, 4}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LogTicks(3,4,5) = %v, want %v", got, want)
	}
}

func TestLogTicks_ThinsPowersFromTheLowest(t *testing.T) {
	got := scale.LogTicks(1, 1e9, 4)
	want := []float64{1, 1e3, 1e6, 1e9}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LogTicks(1,1e9,4) = %v, want %v", got, want)
	}
	got = scale.LogTicks(1, 1e9, 3)
	want = []float64{1, 1e4, 1e8}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LogTicks(1,1e9,3) = %v, want %v", got, want)
	}
}

func TestLogTicks_InvalidRange(t *testing.T) {
	for _, r := range [][2]float64{{0, 10}, {-1, 10}, {10, 1}, {math.NaN(), 10}, {1, math.Inf(1)}} {
		if got := scale.LogTicks(r[0], r[1], 5); got != nil {
			t.Errorf("LogTicks(%v,%v,5) = %v, want nil", r[0], r[1], got)
		}
	}
}

func TestLogRange_WidensToWholeDecades(t *testing.T) {
	tests := []struct {
		min, max, lo, hi float64
	}{
		{5, 300, 1, 1000},
		{0.03, 40, 0.01, 100},
		{10, 1000, 10, 1000},
		{1000, 1000, 1000, 10000},
		{20, 80, 10, 100},
		{300, 5, 1, 1000},
	}
	for _, tc := range tests {
		lo, hi := scale.LogRange(tc.min, tc.max)
		if lo != tc.lo || hi != tc.hi {
			t.Errorf("LogRange(%v,%v) = %v, %v, want %v, %v", tc.min, tc.max, lo, hi, tc.lo, tc.hi)
		}
	}
}

func TestLogTicks_KeepsThePowerWhenThinningHalfDecades(t *testing.T) {
	got := scale.LogTicks(3, 30, 2)
	if !slices.Contains(got, 10) {
		t.Fatalf("LogTicks(3, 30, 2) = %v, want it to keep 10", got)
	}
	if len(got) > 2 {
		t.Fatalf("LogTicks(3, 30, 2) = %v, want at most 2 ticks", got)
	}
}

func TestLogRange_ExtremeBoundsStayFinite(t *testing.T) {
	for _, tc := range []struct{ min, max, wantLo, wantHi float64 }{
		{5e-324, 5e-324, 5e-324, 5e-323},
		{1.5e308, 1.5e308, 1e308, 1.5e308},
	} {
		if lo, hi := scale.LogRange(tc.min, tc.max); lo != tc.wantLo || hi != tc.wantHi {
			t.Errorf("LogRange(%v, %v) = %v..%v, want %v..%v", tc.min, tc.max, lo, hi, tc.wantLo, tc.wantHi)
		}
	}
}

func TestLogTicks_KeepsTheOnlyNiceValueBetweenTheBounds(t *testing.T) {
	for _, tc := range []struct {
		min, max float64
		want     []float64
	}{
		{9, 11, []float64{9, 10, 11}},
		{15, 25, []float64{15, 20, 25}},
		{10, 10, []float64{10}},
	} {
		if got := scale.LogTicks(tc.min, tc.max, 5); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("LogTicks(%v,%v,5) = %v, want %v", tc.min, tc.max, got, tc.want)
		}
	}
}

func TestLogRange_ReturnsInvalidBoundsUnchanged(t *testing.T) {
	for _, tc := range [][2]float64{{0, 10}, {-5, 10}, {math.NaN(), 5}, {1, math.Inf(1)}} {
		lo, hi := scale.LogRange(tc[0], tc[1])
		same := func(a, b float64) bool { return a == b || (math.IsNaN(a) && math.IsNaN(b)) }
		if !same(lo, tc[0]) || !same(hi, tc[1]) {
			t.Errorf("LogRange(%v, %v) = %v, %v, want the input", tc[0], tc[1], lo, hi)
		}
	}
}
