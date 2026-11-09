package pulse

import (
	"math"

	"github.com/ingvarch/pulse/scale"
)

// SetLogScale maps values to rows by their base-10 logarithm.
// Values at or below zero are drawn as gaps. A non-positive minimum is drawn as max/10, and a range
// with both bounds non-positive as 1..10; Min and Max keep the stored range.
func (m *Model) SetLogScale(on bool) { m.logScale = on }

// LogScale returns whether the Y axis is logarithmic.
func (m *Model) LogScale() bool { return m.logScale }

// bounds returns the range used for drawing. On a log scale it is the range as set, so the symmetric
// rewrite does not apply. A non-positive minimum becomes max/10, and bounds that are both non-positive become 1..10.
func (m *Model) bounds() (lo, hi float64) {
	if !m.logScale {
		return m.min, m.max
	}
	if m.rangeMin > 0 {
		return m.rangeMin, m.rangeMax
	}
	if m.rangeMax <= 0 {
		return 1, 10
	}
	return m.rangeMax / 10, m.rangeMax
}

// logSpan returns the height of the log plot space: log10(hi/lo).
func logSpan(lo, hi float64) float64 { return math.Log10(hi) - math.Log10(lo) }

// logPlot maps v into the log plot space, where lo is 0. Values at or below zero and gaps become NaN.
func logPlot(lo, v float64) float64 {
	if isGap(v) || v <= 0 {
		return math.NaN()
	}
	return math.Log10(v) - math.Log10(lo)
}

// fitLogRange returns the auto range on a log scale: the positive finite values, widened to whole decades.
func fitLogRange(seriesData [][]float64) (lo, hi float64) {
	lo, hi = math.Inf(1), 0.0
	for _, data := range seriesData {
		for _, v := range data {
			if !isGap(v) && v > 0 {
				lo, hi = math.Min(lo, v), math.Max(hi, v)
			}
		}
	}
	if hi == 0 {
		return 1, 10
	}
	return scale.LogRange(lo, hi)
}

// logChart is the Chart a renderer sees on a log scale: the Y range is the plot space 0..span,
// so row 0 of the plot is the bottom row and nothing is negative.
type logChart struct {
	*Model
	span float64
	data map[string][]float64
}

func (c logChart) Min() float64               { return 0 }
func (c logChart) Max() float64               { return c.span }
func (c logChart) ZeroBaseline() bool         { return false }
func (c logChart) Symmetric() bool            { return false }
func (c logChart) SeriesInverted(string) bool { return false }

func (c logChart) SeriesData(name string) []float64 { return c.data[name] }

// toLogPlot converts the series to the log plot space and returns it with the Chart for the renderer.
func (m *Model) toLogPlot(names []string, seriesData [][]float64) ([][]float64, Chart) {
	lo, hi := m.bounds()
	plot := make([][]float64, len(seriesData))
	byName := make(map[string][]float64, len(names))
	for i, data := range seriesData {
		if data == nil {
			continue
		}
		plot[i] = make([]float64, len(data))
		for j, v := range data {
			plot[i][j] = logPlot(lo, v)
		}
		byName[names[i]] = plot[i]
	}
	return plot, logChart{Model: m, span: logSpan(lo, hi), data: byName}
}
