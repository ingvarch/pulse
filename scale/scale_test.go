package scale_test

import (
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

