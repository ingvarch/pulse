package scale

import (
	"fmt"
	"math"
	"strconv"
)

// NiceTicks returns pleasant tick values covering [min, max].
// Heckbert algorithm: step = 1 / 2 / 2.5 / 5 / 10 * 10^k.
// If the ticks would overflow float64, it returns [min, max].
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
	if math.IsInf(step, 0) || math.IsInf(lo, 0) || math.IsInf(hi, 0) {
		return []float64{min, max}
	}
	var out []float64
	for v := lo; v <= hi+step*0.5 && !math.IsInf(v, 0); v += step {
		v = math.Round(v/step) * step
		out = append(out, v)
	}
	return out
}

// NiceRange rounds [min, max] out to tick values, so that NiceTicks(lo, hi, maxTicks)
// starts at lo and ends at hi. It returns the narrowest such range.
// If no finite step fits, it returns the input.
func NiceRange(min, max float64, maxTicks int) (lo, hi float64) {
	ticks := NiceTicks(min, max, maxTicks)
	if len(ticks) < 2 {
		return min, max
	}
	step := ticks[1] - ticks[0]
	for range 64 {
		lo, hi = math.Floor(min/step)*step, math.Ceil(max/step)*step
		if math.IsInf(hi, 0) || math.IsInf(lo, 0) {
			return min, max
		}
		t := NiceTicks(lo, hi, maxTicks)
		if t[0] == lo && t[len(t)-1] == hi {
			return lo, hi
		}
		step = nextNiceStep(step)
	}
	return min, max
}

// nextNiceStep returns the next value in 1, 2, 2.5, 5 times a power of ten.
func nextNiceStep(step float64) float64 {
	mag := math.Pow(10, math.Floor(math.Log10(step)))
	switch norm := step / mag; {
	case norm < 1.5:
		return 2 * mag
	case norm < 2.25:
		return 2.5 * mag
	case norm < 3.75:
		return 5 * mag
	case norm < 7.5:
		return 10 * mag
	default:
		// Log10 rounded down a power of ten, such as Log10(1e15): step is 10*mag.
		return 20 * mag
	}
}

// FormatTick formats a tick value concisely: up to 4 significant digits without trailing zeros.
func FormatTick(v float64) string {
	return strconv.FormatFloat(v, 'g', 4, 64)
}

func formatBytesWithSuffix(v float64, suffix string, abs bool) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "0 B" + suffix
	}
	neg := v < 0
	if abs {
		neg = false
	}
	val := math.Abs(v)
	if val == 0 {
		return "0 B" + suffix
	}

	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	unitIdx := 0
	for val >= 1024 && unitIdx < len(units)-1 {
		val /= 1024
		unitIdx++
	}

	var formatted string
	if unitIdx == 0 {
		formatted = fmt.Sprintf("%.0f %s%s", val, units[unitIdx], suffix)
	} else {
		rounded := math.Round(val*10) / 10
		if rounded == math.Floor(rounded) {
			formatted = fmt.Sprintf("%.0f %s%s", rounded, units[unitIdx], suffix)
		} else {
			formatted = fmt.Sprintf("%.1f %s%s", rounded, units[unitIdx], suffix)
		}
	}

	if neg {
		return "-" + formatted
	}
	return formatted
}

// FormatBytes formats a byte count into a human-readable string (e.g. "1 KB", "1.5 MB").
func FormatBytes(v float64) string {
	return formatBytesWithSuffix(v, "", false)
}

// FormatBytesRate formats a throughput value in bytes per second (e.g. "1 KB/s", "100 MB/s").
func FormatBytesRate(v float64) string {
	return formatBytesWithSuffix(v, "/s", false)
}

// FormatBytesRateAbs formats a throughput value in bytes per second using its absolute value (e.g. "50 MB/s").
// Ideal for mirror/symmetric Y-axes (RX/TX, Read/Write).
func FormatBytesRateAbs(v float64) string {
	return formatBytesWithSuffix(v, "/s", true)
}

// BytesRateFormatter returns a Y-axis label formatter function for bytes per second.
// If abs is true, negative values are formatted without a minus sign (ideal for TX / Write).
func BytesRateFormatter(abs bool) func(float64) string {
	return func(v float64) string {
		return formatBytesWithSuffix(v, "/s", abs)
	}
}
