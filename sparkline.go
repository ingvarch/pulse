package pulse

import (
	"math"
	"slices"
	"strings"
)

var sparkBlocks = []rune("▁▂▃▄▅▆▇█")

// Sparkline renders values as one row of block characters (▁▂▃▄▅▆▇█), one value per cell,
// scaled between min and max. Min draws ▁ and max draws █; values outside are clamped.
// A gap (NaN, ±Inf) is a space. Min above max is swapped, and equal bounds work as min..min+1.
// A NaN or infinite bound makes every cell a gap. The result is plain text.
func Sparkline(values []float64, min, max float64) string {
	min, max = orderedRange(min, max)
	var b strings.Builder
	for _, v := range values {
		level, ok := levelOf(min, max, v, 7)
		if !ok {
			b.WriteByte(' ')
			continue
		}
		b.WriteRune(sparkBlocks[level])
	}
	return b.String()
}

// SparklineBraille renders values as braille dots, two values per cell, so it takes half the width of Sparkline.
// Each value is a bar from the bottom of its dot column, one to four dots high; min lights one dot and max four.
// Clamping, gaps, bounds and the range work as in Sparkline. A gap lights nothing, and a cell without dots is a space.
// The result is plain text.
func SparklineBraille(values []float64, min, max float64) string {
	min, max = orderedRange(min, max)
	var b strings.Builder
	for pair := range slices.Chunk(values, 2) {
		var bits uint8
		for dx, v := range pair {
			level, ok := levelOf(min, max, v, 3)
			if !ok {
				continue
			}
			for dy := 3 - level; dy < 4; dy++ {
				bits |= brailleBits[dy][dx]
			}
		}
		if bits == 0 {
			b.WriteByte(' ')
			continue
		}
		b.WriteRune(0x2800 + rune(bits))
	}
	return b.String()
}

// Sparkline renders the named series as one row of blocks in the series style.
// The name "" is the default series. An unknown name or a series without points gives "".
// It draws the points the chart shows, so it follows SetOffset, with the range View draws them in:
// the fitted range with auto range, the right axis range for a right-axis series, the log scale with SetLogScale.
// Inverted series are negated as in View. Thresholds, negative styles and the render mode do not apply.
// With auto range it updates Min and Max as View does.
func (m *Model) Sparkline(name string) string {
	names, seriesData, onRight, _ := m.plotSeries()
	idx := slices.Index(names, name)
	if idx < 0 || len(seriesData[idx]) == 0 {
		return ""
	}
	values := seriesData[idx]
	var lo, hi float64
	switch {
	case onRight[idx]:
		lo, hi = m.rightAxis.Min, m.rightAxis.Max
	case m.logScale:
		lo, hi = m.bounds()
		values = slices.Clone(values)
		for i, v := range values {
			values[i] = logPlot(lo, v)
		}
		hi = logSpan(lo, hi)
		lo = 0
	default:
		lo, hi = m.bounds()
	}
	return m.SeriesStyle(name).Render(Sparkline(values, lo, hi))
}

// levelOf returns the level of v from 0 to steps, and false when v or the bounds leave no level.
func levelOf(min, max, v float64, steps int) (int, bool) {
	norm := normOf(min, max, v)
	if isGap(v) || isGap(min) || isGap(max) || math.IsNaN(norm) {
		return 0, false
	}
	return int(math.Round(norm * float64(steps))), true
}

// orderedRange swaps min and max when min is above max.
func orderedRange(min, max float64) (float64, float64) {
	if min > max {
		return max, min
	}
	return min, max
}
