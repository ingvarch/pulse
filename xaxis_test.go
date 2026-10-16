package pulse_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/ingvarch/pulse"
)

// clockLabels labels each age with the time `age` minutes before newest.
func clockLabels(newest time.Time) func(int) string {
	return func(age int) string {
		return newest.Add(-time.Duration(age) * time.Minute).Format("15:04")
	}
}

func clock(h, m int) time.Time { return time.Date(2026, 10, 15, h, m, 0, 0, time.UTC) }

func newXAxisChart(w int, opts ...pulse.Option) *pulse.Model {
	return pulse.New(w, 5, append([]pulse.Option{pulse.WithRange(0, 100)}, opts...)...)
}

func axisRows(c *pulse.Model) (axis, labels string) {
	lines := strings.Split(ansi.Strip(c.View()), "\n")
	n := len(lines)
	return lines[n-2], lines[n-1]
}

func assertAxisRows(t *testing.T, c *pulse.Model, wantAxis, wantLabels string) {
	t.Helper()
	axis, labels := axisRows(c)
	if axis != wantAxis {
		t.Fatalf("axis row:\n got %q\nwant %q", axis, wantAxis)
	}
	if labels != wantLabels {
		t.Fatalf("label row:\n got %q\nwant %q", labels, wantLabels)
	}
}

func TestXAxis_Layouts(t *testing.T) {
	tests := []struct {
		name   string
		w      int
		newest time.Time
		axis   string
		labels string
	}{
		{
			"w40", 40, clock(14, 39),
			"    └──────────┴─────────┴─────────┴─────────",
			"     14:00   14:10     14:20     14:30  14:39",
		},
		{
			"w61", 61, clock(15, 0),
			"    └───────────────┴──────────────┴──────────────┴───────────────",
			"     14:00        14:15          14:30          14:45        15:00",
		},
		{
			"w12 only edges fit", 12, clock(15, 0),
			"    └────────────",
			"     14:49  15:00",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := newXAxisChart(tc.w, pulse.WithXAxis(clockLabels(tc.newest)))
			assertAxisRows(t, c, tc.axis, tc.labels)
		})
	}
}

func TestXAxis_HeightAndPlotUnchanged(t *testing.T) {
	label := clockLabels(clock(14, 39))
	with := newXAxisChart(40, pulse.WithXAxis(label))
	without := newXAxisChart(40)
	for i := 0; i < 40; i++ {
		with.Push(float64(i * 2))
		without.Push(float64(i * 2))
	}
	if with.Height() != 5 {
		t.Fatalf("Height() = %d, want 5", with.Height())
	}
	got := strings.Split(with.View(), "\n")
	want := strings.Split(without.View(), "\n")
	if len(got) != with.Height()+2 {
		t.Fatalf("View has %d lines, want %d", len(got), with.Height()+2)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("plot row %d differs:\n got %q\nwant %q", i, got[i], want[i])
		}
	}
}

func TestXAxis_EmptyLabels(t *testing.T) {
	t.Run("only newest", func(t *testing.T) {
		c := newXAxisChart(40, pulse.WithXAxis(func(age int) string {
			if age == 0 {
				return "now"
			}
			return ""
		}))
		assertAxisRows(t, c,
			"    └"+strings.Repeat("─", 40),
			strings.Repeat(" ", 42)+"now")
	})
	t.Run("all empty", func(t *testing.T) {
		c := newXAxisChart(40, pulse.WithXAxis(func(int) string { return "" }))
		assertAxisRows(t, c,
			"    └"+strings.Repeat("─", 40),
			strings.Repeat(" ", 45))
	})
}

func TestXAxis_NarrowWidths(t *testing.T) {
	tests := []struct {
		name   string
		w      int
		labels string
	}{
		{"w1 nothing fits", 1, strings.Repeat(" ", 6)},
		{"w4 nothing fits", 4, strings.Repeat(" ", 9)},
		{"w5 right edge only", 5, "     15:00"},
		{"w8 right edge wins over left", 8, "        15:00"},
		{"w11 both edges", 11, "     14:50 15:00"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := newXAxisChart(tc.w, pulse.WithXAxis(clockLabels(clock(15, 0))))
			assertAxisRows(t, c, "    └"+strings.Repeat("─", tc.w), tc.labels)
		})
	}
}

func TestXAxis_TooWideLabelDropped(t *testing.T) {
	c := newXAxisChart(10, pulse.WithXAxis(func(age int) string {
		if age == 9 {
			return "far too long label"
		}
		return ""
	}))
	assertAxisRows(t, c,
		"    └"+strings.Repeat("─", 10),
		strings.Repeat(" ", 15))
}

func TestXAxis_Spacing(t *testing.T) {
	tests := []struct {
		name   string
		labels map[int]string
		axis   string
		row    string
	}{
		{
			"one space between labels is kept",
			map[int]string{39: "1234567", 29: "abcde"},
			"    └" + strings.Repeat("─", 10) + "┴" + strings.Repeat("─", 29),
			"     1234567 abcde" + strings.Repeat(" ", 27),
		},
		{
			"touching label is dropped",
			map[int]string{39: "12345678", 29: "abcde"},
			"    └" + strings.Repeat("─", 40),
			"     12345678" + strings.Repeat(" ", 32),
		},
		{
			"even width centers left of the column",
			map[int]string{19: "ab"},
			"    └" + strings.Repeat("─", 20) + "┴" + strings.Repeat("─", 19),
			strings.Repeat(" ", 25) + "ab" + strings.Repeat(" ", 18),
		},
		{
			"w/4 is placed before 2w/4",
			map[int]string{29: "abcdefghijk", 19: "abcdefghijk"},
			"    └" + strings.Repeat("─", 10) + "┴" + strings.Repeat("─", 29),
			strings.Repeat(" ", 10) + "abcdefghijk" + strings.Repeat(" ", 24),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := newXAxisChart(40, pulse.WithXAxis(func(age int) string { return tc.labels[age] }))
			assertAxisRows(t, c, tc.axis, tc.row)
		})
	}
}

func TestXAxis_AxisStyle(t *testing.T) {
	st := lipgloss.NewStyle().Foreground(lipgloss.Color("#7aa2f7"))
	c := newXAxisChart(12, pulse.WithAxisStyle(st), pulse.WithXAxis(clockLabels(clock(15, 0))))
	lines := strings.Split(c.View(), "\n")
	n := len(lines)
	if want := "    " + st.Render("└────────────"); lines[n-2] != want {
		t.Fatalf("axis row:\n got %q\nwant %q", lines[n-2], want)
	}
	if want := "     " + st.Render("14:49") + "  " + st.Render("15:00"); lines[n-1] != want {
		t.Fatalf("label row:\n got %q\nwant %q", lines[n-1], want)
	}
}

func TestXAxis_IndependentOfGridModeEvents(t *testing.T) {
	label := clockLabels(clock(14, 39))
	wantAxis := "    └──────────┴─────────┴─────────┴─────────"
	wantLabels := "     14:00   14:10     14:20     14:30  14:39"
	for _, mode := range gapModes {
		for _, grid := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/grid=%v", mode.name, grid), func(t *testing.T) {
				c := newXAxisChart(40, pulse.WithXAxis(label),
					pulse.WithMode(mode.mode), pulse.WithGrid(grid),
					pulse.WithEvent(10, pulse.Event{Label: "deploy", Glyph: "*"}))
				for i := 0; i < 40; i++ {
					c.Push(float64(i))
				}
				assertAxisRows(t, c, wantAxis, wantLabels)
			})
		}
	}
}

func TestXAxis_SetXAxisNilTurnsOff(t *testing.T) {
	c := newXAxisChart(40)
	c.SetXAxis(clockLabels(clock(14, 39)))
	if n := len(strings.Split(c.View(), "\n")); n != 7 {
		t.Fatalf("axis on: %d lines, want 7", n)
	}
	c.SetXAxis(nil)
	if got, want := c.View(), newXAxisChart(40).View(); got != want {
		t.Fatalf("axis off differs from a chart that never had it:\n%s\n%s", got, want)
	}
}

func TestXAxis_LabelsFollowPushes(t *testing.T) {
	newest := clock(14, 39)
	c := newXAxisChart(40, pulse.WithXAxis(func(age int) string {
		return clockLabels(newest)(age)
	}))
	_, first := axisRows(c)
	newest = newest.Add(time.Minute)
	c.Push(1)
	_, second := axisRows(c)
	if first == second {
		t.Fatalf("label row did not change after a push: %q", first)
	}
	if !strings.HasSuffix(second, "14:40") {
		t.Fatalf("label row %q must end with the new time", second)
	}
}

func TestXAxis_EqualLineWidths(t *testing.T) {
	labels := map[string]func(int) string{
		"ascii": clockLabels(clock(14, 39)),
		"cjk":   func(age int) string { return fmt.Sprintf("%d分前", age) },
	}
	axes := map[string]func(func(int) string) pulse.Option{
		"WithXAxis":   pulse.WithXAxis,
		"WithXAxisAt": pulse.WithXAxisAt,
	}
	for axisName, axis := range axes {
		for name, label := range labels {
			t.Run(axisName+"/"+name, func(t *testing.T) {
				c := newXAxisChart(40, axis(label),
					pulse.WithAxisStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#444b6a"))))
				for _, line := range strings.Split(c.View(), "\n") {
					if w := lipgloss.Width(line); w != 4+1+40 {
						t.Fatalf("line %q has width %d, want %d", ansi.Strip(line), w, 45)
					}
				}
			})
		}
	}
}

func TestXAxis_CustomRendererGetsAxisRows(t *testing.T) {
	c := newXAxisChart(12, pulse.WithRenderer(staticRenderer{}), pulse.WithXAxis(clockLabels(clock(15, 0))))
	want := "STATIC_CHART\n    └────────────\n     14:49  15:00"
	if got := ansi.Strip(c.View()); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// sixHourLabels labels the newest column "now" and each column that holds a time of day divisible by 6 hours.
func sixHourLabels(newest time.Time, step time.Duration) func(int) string {
	return func(age int) string {
		if age == 0 {
			return "now"
		}
		t := newest.Add(-time.Duration(age) * step)
		if b := t.Truncate(6 * time.Hour); t.Sub(b) < step {
			return b.Format("15:04")
		}
		return ""
	}
}

func mapLabels(labels map[int]string) func(int) string {
	return func(age int) string { return labels[age] }
}

// vGridRow is a plot row without data or a horizontal grid line, with vertical grid lines at cols.
func vGridRow(w int, cols ...int) string {
	row := []rune(strings.Repeat(" ", w))
	for _, col := range cols {
		row[col] = '│'
	}
	return "    │" + string(row)
}

func TestXAxisAt_RoundTimes(t *testing.T) {
	c := newXAxisChart(48, pulse.WithXAxisAt(sixHourLabels(clock(15, 10), 30*time.Minute)))
	assertAxisRows(t, c,
		"    └─────┴───────────┴───────────┴───────────┴──────",
		"        18:00       00:00       06:00       12:00 now")
}

func TestXAxisAt_Placement(t *testing.T) {
	tests := []struct {
		name   string
		labels map[int]string
		axis   string
		row    string
	}{
		{
			"middle label is centered on its column",
			map[int]string{12: "abcde"},
			"    └" + strings.Repeat("─", 27) + "┴" + strings.Repeat("─", 12),
			strings.Repeat(" ", 30) + "abcde" + strings.Repeat(" ", 10),
		},
		{
			"edge labels align to the edges without a tick",
			map[int]string{0: "now", 39: "start"},
			"    └" + strings.Repeat("─", 40),
			"     start" + strings.Repeat(" ", 32) + "now",
		},
		{
			"label that runs past the left edge is dropped",
			map[int]string{38: "abcde"},
			"    └" + strings.Repeat("─", 40),
			strings.Repeat(" ", 45),
		},
		{
			"label that ends at the left edge fits",
			map[int]string{37: "abcde"},
			"    └──┴" + strings.Repeat("─", 37),
			"     abcde" + strings.Repeat(" ", 35),
		},
		{
			"label that runs past the right edge is dropped",
			map[int]string{1: "abcde"},
			"    └" + strings.Repeat("─", 40),
			strings.Repeat(" ", 45),
		},
		{
			"one space between labels is kept",
			map[int]string{10: "abcde", 16: "fghij"},
			"    └" + strings.Repeat("─", 23) + "┴─────┴" + strings.Repeat("─", 10),
			strings.Repeat(" ", 26) + "fghij abcde" + strings.Repeat(" ", 8),
		},
		{
			"touching label to the left is dropped",
			map[int]string{10: "abcde", 15: "fghij"},
			"    └" + strings.Repeat("─", 29) + "┴" + strings.Repeat("─", 10),
			strings.Repeat(" ", 32) + "abcde" + strings.Repeat(" ", 8),
		},
		{
			"right edge stays",
			map[int]string{0: "now", 2: "abcde"},
			"    └" + strings.Repeat("─", 40),
			strings.Repeat(" ", 42) + "now",
		},
		{
			"middle label wins over the left edge",
			map[int]string{39: "start", 35: "abcde"},
			"    └────┴" + strings.Repeat("─", 35),
			"       abcde" + strings.Repeat(" ", 33),
		},
		{
			"no candidates",
			map[int]string{},
			"    └" + strings.Repeat("─", 40),
			strings.Repeat(" ", 45),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := newXAxisChart(40, pulse.WithXAxisAt(mapLabels(tc.labels)))
			assertAxisRows(t, c, tc.axis, tc.row)
		})
	}
}

func TestXAxisAt_NarrowWidths(t *testing.T) {
	tests := []struct {
		name   string
		w      int
		labels map[int]string
		row    string
	}{
		{"w1 label at the only column", 1, map[int]string{0: "x"}, "     x"},
		{"w2 left edge touches the right one", 2, map[int]string{0: "a", 1: "b"}, "      a"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := newXAxisChart(tc.w, pulse.WithXAxisAt(mapLabels(tc.labels)))
			assertAxisRows(t, c, "    └"+strings.Repeat("─", tc.w), tc.row)
		})
	}
}

func TestXAxisAt_GridFollowsTicks(t *testing.T) {
	tests := []struct {
		name   string
		labels map[int]string
		cols   []int
	}{
		// "xyz" at age 2 touches "now" and is dropped; the edge labels "now" and "start" get no grid line.
		{"placed middle labels", map[int]string{0: "now", 2: "xyz", 9: "abcde", 15: "fghij", 30: "klmno", 39: "start"}, []int{9, 24, 30}},
		{"no candidates", map[int]string{}, nil},
	}
	for _, mode := range gapModes {
		for _, tc := range tests {
			t.Run(mode.name+"/"+tc.name, func(t *testing.T) {
				c := newXAxisChart(40, pulse.WithMode(mode.mode), pulse.WithTicks(0, 100),
					pulse.WithXAxisAt(mapLabels(tc.labels)))
				lines := strings.Split(ansi.Strip(c.View()), "\n")
				if want := vGridRow(40, tc.cols...); lines[1] != want {
					t.Fatalf("plot row 1:\n got %q\nwant %q", lines[1], want)
				}
			})
		}
	}
}

func TestXAxis_GridStaysAtQuarters(t *testing.T) {
	c := newXAxisChart(40, pulse.WithTicks(0, 100),
		pulse.WithXAxis(mapLabels(map[int]string{0: "now", 19: "abcde"})))
	lines := strings.Split(ansi.Strip(c.View()), "\n")
	if want := vGridRow(40, 10, 20, 30); lines[1] != want {
		t.Fatalf("plot row 1:\n got %q\nwant %q", lines[1], want)
	}
}

func TestXAxisAt_AsksEveryColumnOnce(t *testing.T) {
	for _, w := range []int{1, 2, 40} {
		t.Run(fmt.Sprintf("w%d", w), func(t *testing.T) {
			asked := map[int]int{}
			c := newXAxisChart(w, pulse.WithXAxisAt(func(age int) string {
				asked[age]++
				return ""
			}))
			c.View()
			if len(asked) != w {
				t.Fatalf("asked ages %v, want 0..%d", asked, w-1)
			}
			for age := 0; age < w; age++ {
				if asked[age] != 1 {
					t.Fatalf("age %d asked %d times, want 1", age, asked[age])
				}
			}
		})
	}
}

func TestXAxisAt_SetXAxisAtNilTurnsOff(t *testing.T) {
	c := newXAxisChart(40)
	c.SetXAxisAt(sixHourLabels(clock(15, 10), 30*time.Minute))
	if n := len(strings.Split(c.View(), "\n")); n != 7 {
		t.Fatalf("axis on: %d lines, want 7", n)
	}
	c.SetXAxisAt(nil)
	if got, want := c.View(), newXAxisChart(40).View(); got != want {
		t.Fatalf("axis off differs from a chart that never had it:\n%s\n%s", got, want)
	}
}

func TestXAxisAt_LastSetterWins(t *testing.T) {
	label := sixHourLabels(clock(15, 10), 30*time.Minute)
	t.Run("SetXAxis after WithXAxisAt", func(t *testing.T) {
		c := newXAxisChart(48, pulse.WithXAxisAt(label))
		c.SetXAxis(label)
		if got, want := c.View(), newXAxisChart(48, pulse.WithXAxis(label)).View(); got != want {
			t.Fatalf("got\n%s\nwant\n%s", got, want)
		}
	})
	t.Run("SetXAxisAt after WithXAxis", func(t *testing.T) {
		c := newXAxisChart(48, pulse.WithXAxis(label))
		c.SetXAxisAt(label)
		if got, want := c.View(), newXAxisChart(48, pulse.WithXAxisAt(label)).View(); got != want {
			t.Fatalf("got\n%s\nwant\n%s", got, want)
		}
	})
}
