package scale

import (
	"math"
	"strconv"
)

// NiceTicks returns pleasant tick values covering [min, max].
// Heckbert algorithm: step = 1 / 2 / 2.5 / 5 / 10 * 10^k.
func NiceTicks(min, max float64, maxTicks int) []float64 {
	if maxTicks < 2 {
		maxTicks = 2
	}
	if math.IsNaN(min) || math.IsNaN(max) {
		return []float64{0}
	}
	if min == max {
		min -= 1
		max += 1
	}
	if min > max {
		min, max = max, min
	}
	rawStep := (max - min) / float64(maxTicks)
	mag := math.Pow(10, math.Floor(math.Log10(rawStep)))
	norm := rawStep / mag
	var nice float64
	switch {
	case norm < 1.5:
		nice = 1
	case norm < 2.25:
		nice = 2
	case norm < 3.75:
		nice = 2.5
	case norm < 7.5:
		nice = 5
	default:
		nice = 10
	}
	step := nice * mag
	lo := math.Floor(min/step) * step
	hi := math.Ceil(max/step) * step
	var out []float64
	for v := lo; v <= hi+step*0.5; v += step {
		v = math.Round(v/step) * step
		out = append(out, v)
	}
	return out
}

// FormatTick formats a tick value concisely: up to 4 significant digits without trailing zeros.
func FormatTick(v float64) string {
	return strconv.FormatFloat(v, 'g', 4, 64)
}
