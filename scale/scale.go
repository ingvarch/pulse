package scale

import (
	"fmt"
	"math"
	"slices"
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

// LogTicks returns tick values for a logarithmic axis over [min, max]: the powers of ten inside it.
// With fewer than two powers it adds the 2x and 5x values of each decade, and if that still
// gives fewer than two ticks it adds min and max. More than maxTicks ticks are thinned to
// every k-th one, counted from the lowest power of ten (or the lowest tick when there is none).
// It returns nil unless 0 < min <= max and both are finite.
func LogTicks(min, max float64, maxTicks int) []float64 {
	if !(min > 0 && max >= min) || math.IsInf(max, 0) {
		return nil
	}
	if maxTicks < 2 {
		maxTicks = 2
	}
	first, last := decadeOf(min), decadeOf(max)
	ticks := logSteps(min, max, first, last, 1)
	if len(ticks) < 2 {
		ticks = logSteps(min, max, first, last, 1, 2, 5)
	}
	if len(ticks) < 2 {
		ticks = append(ticks, min, max)
		slices.Sort(ticks)
		return slices.Compact(ticks)
	}
	return thin(ticks, maxTicks)
}

// thin keeps every k-th tick so at most maxTicks remain. The steps start at the first power of ten,
// or at the lowest tick when there is none, and the result keeps two ticks when it can.
func thin(ticks []float64, maxTicks int) []float64 {
	k := (len(ticks) + maxTicks - 1) / maxTicks
	anchor := slices.IndexFunc(ticks, func(v float64) bool { return v == math.Pow10(decadeOf(v)) })
	if anchor < 0 {
		anchor = 0
	}
	var out []float64
	for i := anchor % k; i < len(ticks); i += k {
		out = append(out, ticks[i])
	}
	if len(out) < 2 && len(ticks) >= 2 {
		return ticks[min(anchor, len(ticks)-2):][:2]
	}
	return out
}

// LogRange rounds [min, max] out to whole decades: 10^floor(log10(min)) to 10^ceil(log10(max)).
// When both land on the same power, hi is ten times lo. A bound that has no finite power of ten
// (below the smallest or above the largest) stays as it is. The bounds are swapped when min > max.
// Unless both are positive and finite, it returns the input.
func LogRange(min, max float64) (lo, hi float64) {
	if !(min > 0 && max > 0) || math.IsInf(min, 0) || math.IsInf(max, 0) {
		return min, max
	}
	if min > max {
		min, max = max, min
	}
	lo = math.Pow10(decadeOf(min))
	if lo == 0 {
		lo = min
	}
	hi = math.Pow10(decadeOf(max))
	if hi < max {
		hi *= 10
	}
	if hi < max || math.IsInf(hi, 0) {
		hi = max
	}
	if hi == lo && !math.IsInf(lo*10, 0) {
		hi = lo * 10
	}
	return lo, hi
}

// logSteps returns mult*10^e inside [min, max] for each decade e in [first, last], ascending.
func logSteps(min, max float64, first, last int, mult ...float64) []float64 {
	var out []float64
	for e := first; e <= last; e++ {
		for _, m := range mult {
			if v := m * math.Pow10(e); v >= min && v <= max {
				out = append(out, v)
			}
		}
	}
	return out
}

// decadeOf returns floor(log10(v)), corrected for rounding in Log10.
func decadeOf(v float64) int {
	e := int(math.Floor(math.Log10(v)))
	if math.Pow10(e+1) <= v {
		e++
	} else if math.Pow10(e) > v {
		e--
	}
	return e
}
