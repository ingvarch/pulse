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
