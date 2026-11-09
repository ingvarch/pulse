package pulse_test

import (
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/ingvarch/pulse"
)

var historyModes = []struct {
	name string
	mode pulse.RenderMode
}{
	{"lines", pulse.ModeLines},
	{"braille", pulse.ModeBraille},
}

func wave(i int) float64 { return 50 + 40*math.Sin(float64(i)*0.4) }

func pushWave(m *pulse.Model, from, to int) {
	for i := from; i < to; i++ {
		m.Push(wave(i))
	}
}

func TestSetSize_NarrowThenWidenWithHistoryRestoresColumns(t *testing.T) {
	for _, tc := range historyModes {
		t.Run(tc.name, func(t *testing.T) {
			m := pulse.New(20, 6, pulse.WithMode(tc.mode), pulse.WithHistory(40))
			pushWave(m, 0, 40)

			m.SetSize(10, 6)
			want := pulse.New(10, 6, pulse.WithMode(tc.mode))
			pushWave(want, 30, 40)
			if got := m.View(); got != want.View() {
				t.Fatalf("narrow view differs from a chart that holds the last 10 points\ngot:\n%s\nwant:\n%s", got, want.View())
			}

			pushWave(m, 40, 55)
			m.SetSize(20, 6)
			wide := pulse.New(20, 6, pulse.WithMode(tc.mode))
			pushWave(wide, 0, 55)
			if got := m.View(); got != wide.View() {
				t.Fatalf("widened view differs from an always wide chart\ngot:\n%s\nwant:\n%s", got, wide.View())
			}
		})
	}
}

func TestSetSize_WidenWithoutHistoryLeavesLeftColumnsEmpty(t *testing.T) {
	for _, tc := range historyModes {
		t.Run(tc.name, func(t *testing.T) {
			m := pulse.New(10, 6, pulse.WithMode(tc.mode))
			pushWave(m, 0, 30)

			m.SetSize(20, 6)
			want := pulse.New(20, 6, pulse.WithMode(tc.mode))
			pushWave(want, 20, 30)
			if got := m.View(); got != want.View() {
				t.Fatalf("widened view differs from a chart with only the kept points\ngot:\n%s\nwant:\n%s", got, want.View())
			}
			if n := m.Len(); n != 10 {
				t.Fatalf("Len = %d, want 10", n)
			}
		})
	}
}

func TestSetSize_HeightChangeMovesTicksAndKeepsData(t *testing.T) {
	for _, tc := range historyModes {
		t.Run(tc.name, func(t *testing.T) {
			m := pulse.New(20, 4, pulse.WithMode(tc.mode))
			pushWave(m, 0, 20)
			before := m.View()

			m.SetSize(20, 10)
			want := pulse.New(20, 10, pulse.WithMode(tc.mode))
			pushWave(want, 0, 20)
			got := m.View()
			if got != want.View() {
				t.Fatalf("resized view differs from a fresh chart\ngot:\n%s\nwant:\n%s", got, want.View())
			}
			if got == before || strings.Count(got, "\n") <= strings.Count(before, "\n") {
				t.Fatal("height change did not add rows")
			}
			if m.Height() != 10 || m.Len() != 20 {
				t.Fatalf("Height = %d, Len = %d, want 10 and 20", m.Height(), m.Len())
			}
		})
	}
}

func TestSetSize_ClampsToOne(t *testing.T) {
	m := pulse.New(5, 5)
	m.SetSize(0, 0)
	if m.Width() != 1 || m.Height() != 1 {
		t.Fatalf("size = %dx%d, want 1x1", m.Width(), m.Height())
	}
	m.SetSize(-4, -2)
	if m.Width() != 1 || m.Height() != 1 {
		t.Fatalf("size = %dx%d, want 1x1", m.Width(), m.Height())
	}
	m.Push(10)
	_ = m.View()
}

func TestSetSize_EventComesBackAfterWidening(t *testing.T) {
	for _, tc := range historyModes {
		t.Run(tc.name, func(t *testing.T) {
			m := pulse.New(20, 6, pulse.WithMode(tc.mode), pulse.WithHistory(40))
			pushWave(m, 0, 40)
			m.AddEventAt(15, pulse.Event{Label: "Deploy", Glyph: "X"})

			m.SetSize(10, 6)
			if ev := m.VisibleEvents(); len(ev) != 0 {
				t.Fatalf("narrow window shows %d events, want 0", len(ev))
			}

			pushWave(m, 40, 43)
			m.SetSize(20, 6)
			ev := m.VisibleEvents()
			if len(ev) != 1 || ev[0].Col != 1 || ev[0].Age != 18 {
				t.Fatalf("events after widening = %+v, want one at col 1, age 18", ev)
			}
			if !strings.Contains(m.View(), "X") {
				t.Fatal("event pin is not drawn after widening")
			}
		})
	}
}

func TestSetSize_EventLeavingHistoryIsPruned(t *testing.T) {
	m := pulse.New(20, 6, pulse.WithHistory(40))
	pushWave(m, 0, 40)
	m.AddEvent(pulse.Event{Label: "Old"})
	pushWave(m, 40, 80)
	if ev := m.VisibleEvents(); len(ev) != 0 {
		t.Fatalf("event outside the window is visible: %+v", ev)
	}

	m.SetSize(60, 6)
	if ev := m.VisibleEvents(); len(ev) != 0 {
		t.Fatalf("event older than the history is visible after widening: %+v", ev)
	}
}

func TestSetSize_EventPrunedWithoutRenderBeforeWidening(t *testing.T) {
	cases := []struct {
		name            string
		w, history      int
		pushes, newSize int
	}{
		{"no history", 20, 0, 30, 40},
		{"history", 20, 40, 50, 60},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := pulse.New(tc.w, 4, pulse.WithHistory(tc.history))
			m.AddEvent(pulse.Event{Label: "E"})
			pushWave(m, 0, tc.pushes)

			m.SetSize(tc.newSize, 4)
			if ev := m.VisibleEvents(); len(ev) != 0 {
				t.Fatalf("event whose points were dropped is visible: %+v", ev)
			}
		})
	}
}

func TestSetHistory_RaiseDoesNotRestoreDroppedEvent(t *testing.T) {
	m := pulse.New(20, 4)
	m.AddEvent(pulse.Event{Label: "E"})
	pushWave(m, 0, 30)

	m.SetHistory(60)
	m.SetSize(40, 4)
	if ev := m.VisibleEvents(); len(ev) != 0 {
		t.Fatalf("event whose points were dropped is visible: %+v", ev)
	}
}

func TestHistory_SeriesNeverHoldsMoreThanCapacity(t *testing.T) {
	cases := []struct {
		name    string
		w       int
		history int
		want    int
	}{
		{"no history", 5, 0, 5},
		{"history above width", 5, 8, 8},
		{"history below width", 5, 3, 5},
		{"negative history", 5, -2, 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := pulse.New(tc.w, 4, pulse.WithHistory(tc.history))
			pushWave(m, 0, 50)
			m.PushSeries("a", 1)
			if m.Len() != tc.w {
				t.Fatalf("visible Len = %d, want %d", m.Len(), tc.w)
			}
			m.SetSize(100, 4)
			if m.Len() != tc.want {
				t.Fatalf("stored Len = %d, want %d", m.Len(), tc.want)
			}
		})
	}
}

func TestSetHistory_ShrinkTrimsStoredData(t *testing.T) {
	m := pulse.New(5, 4, pulse.WithHistory(30))
	pushWave(m, 0, 50)

	m.SetHistory(8)
	m.SetSize(100, 4)
	if m.Len() != 8 {
		t.Fatalf("Len = %d, want 8", m.Len())
	}
}

func TestSetSize_ShrinkWithoutHistoryTrimsStoredData(t *testing.T) {
	m := pulse.New(10, 4)
	pushWave(m, 0, 10)

	m.SetSize(3, 4)
	m.SetSize(10, 4)
	if m.Len() != 3 {
		t.Fatalf("Len = %d, want 3", m.Len())
	}
}

func TestHistory_AccessorsUseVisibleWindow(t *testing.T) {
	m := pulse.New(10, 4, pulse.WithHistory(30))
	pushWave(m, 0, 30)
	m.SetSize(4, 4)

	data := m.SeriesData("")
	if len(data) != 4 || data[3] != wave(29) || data[0] != wave(26) {
		t.Fatalf("SeriesData = %v, want the last 4 points", data)
	}
	if v, ok := m.Last(""); !ok || v != wave(29) {
		t.Fatalf("Last = %v, %v", v, ok)
	}
}

func TestHistory_LastUsesVisibleWindow(t *testing.T) {
	m := pulse.New(3, 4, pulse.WithHistory(30))
	m.Push(7)
	for range 5 {
		m.Push(math.NaN())
	}
	if v, ok := m.Last(""); ok {
		t.Fatalf("Last = %v, true; the visible window holds only gaps", v)
	}

	m.SetSize(10, 4)
	if v, ok := m.Last(""); !ok || v != 7 {
		t.Fatalf("Last after widening = %v, %v, want 7, true", v, ok)
	}
}

type lenRecorder struct{ lens []int }

func (r *lenRecorder) Render(_ pulse.Chart, ctx pulse.RenderContext) string {
	r.lens = r.lens[:0]
	for _, d := range ctx.SeriesData {
		r.lens = append(r.lens, len(d))
	}
	return ""
}

func TestHistory_RendererGetsOnlyVisibleWindow(t *testing.T) {
	rec := &lenRecorder{}
	m := pulse.New(20, 4, pulse.WithRenderer(rec), pulse.WithHistory(40))
	m.SetSeriesInverted("b", true)
	for i := range 40 {
		m.Push(wave(i))
		m.PushSeries("b", wave(i))
	}
	m.SetSize(10, 4)
	_ = m.View()
	if len(rec.lens) != 2 || rec.lens[0] != 10 || rec.lens[1] != 10 {
		t.Fatalf("ctx.SeriesData lengths = %v, want [10 10]", rec.lens)
	}
}

func TestHistory_InvertedSeriesUsesVisibleWindow(t *testing.T) {
	for _, tc := range historyModes {
		t.Run(tc.name, func(t *testing.T) {
			m := pulse.New(20, 6, pulse.WithMode(tc.mode), pulse.WithRange(-100, 100), pulse.WithHistory(40), pulse.WithInverted(true))
			pushWave(m, 0, 40)
			m.SetSize(10, 6)

			want := pulse.New(10, 6, pulse.WithMode(tc.mode), pulse.WithRange(-100, 100), pulse.WithInverted(true))
			pushWave(want, 30, 40)
			if m.View() != want.View() {
				t.Fatalf("inverted narrow view differs\ngot:\n%s\nwant:\n%s", m.View(), want.View())
			}
		})
	}
}

func TestSetSize_XAxisFollowsWidth(t *testing.T) {
	label := func(age int) string {
		if age%5 == 0 {
			return strconv.Itoa(age)
		}
		return ""
	}
	axes := map[string]func(func(int) string) pulse.Option{
		"WithXAxis":   pulse.WithXAxis,
		"WithXAxisAt": pulse.WithXAxisAt,
	}
	for name, axis := range axes {
		t.Run(name, func(t *testing.T) {
			m := pulse.New(20, 4, axis(label), pulse.WithHistory(40))
			pushWave(m, 0, 40)
			m.SetSize(10, 4)

			want := pulse.New(10, 4, axis(label))
			pushWave(want, 30, 40)
			if m.View() != want.View() {
				t.Fatalf("x axis view differs\ngot:\n%s\nwant:\n%s", m.View(), want.View())
			}
		})
	}
}

func TestHistory_EventKeptWhateverTheOptionOrder(t *testing.T) {
	ev := pulse.Event{Label: "E"}
	orders := map[string][]pulse.Option{
		"event first":   {pulse.WithEvent(30, ev), pulse.WithHistory(100)},
		"history first": {pulse.WithHistory(100), pulse.WithEvent(30, ev)},
	}
	for name, opts := range orders {
		t.Run(name, func(t *testing.T) {
			m := pulse.New(20, 4, opts...)
			m.SetSize(60, 4)
			if got := len(m.VisibleEvents()); got != 1 {
				t.Fatalf("got %d visible events after widening, want 1", got)
			}
		})
	}
}
