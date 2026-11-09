package donut

import (
	"math"
	"strings"
	"testing"
)

func gridText(g [][]int) []string {
	out := make([]string, len(g))
	for i, row := range g {
		var b strings.Builder
		for _, s := range row {
			if s < 0 {
				b.WriteByte('.')
			} else {
				b.WriteByte(byte('A' + s))
			}
		}
		out[i] = b.String()
	}
	return out
}

func count(g [][]int) map[int]int {
	c := map[int]int{}
	for _, row := range g {
		for _, s := range row {
			c[s]++
		}
	}
	return c
}

func sectionsOf(vals ...float64) []Section {
	s := make([]Section, len(vals))
	for i, v := range vals {
		s[i] = Section{Name: string(rune('A' + i)), Value: v}
	}
	return s
}

func TestFrameGolden(t *testing.T) {
	m := New(3, WithFrame(6, 3, 1, 0))
	m.SetSections(sectionsOf(1, 1)...)
	want := []string{
		"BBBBBBAAAAAA",
		"BBBBBBAAAAAA",
		"BB........AA",
		"BB........AA",
		"BB........AA",
		"BB........AA",
		"BB........AA",
		"BB........AA",
		"BB........AA",
		"BB........AA",
		"BBBBBBAAAAAA",
		"BBBBBBAAAAAA",
	}
	got := gridText(m.grid())
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("frame grid:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestRingGolden(t *testing.T) {
	m := New(3)
	m.SetSections(sectionsOf(2, 1, 1)...)
	want := []string{
		"....CCAA....",
		"..CCCCAAAA..",
		".CCCC..AAAA.",
		".CC......AA.",
		"CCC......AAA",
		"CC........AA",
		"BB........AA",
		"BBB......AAA",
		".BB......AA.",
		".BBBB..AAAA.",
		"..BBBBAAAA..",
		"....BBAA....",
	}
	got := gridText(m.grid())
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("ring grid:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestRingClockwiseFromTop(t *testing.T) {
	m := New(8)
	m.SetSections(sectionsOf(1, 1, 1)...)
	g := m.grid()
	mid := len(g[0]) / 2
	if g[0][mid] != 0 {
		t.Errorf("sub-pixel right of the top center = %d, want section 0", g[0][mid])
	}
	if g[0][mid-1] != 2 {
		t.Errorf("sub-pixel left of the top center = %d, want the last section", g[0][mid-1])
	}
	row := len(g) / 2
	if g[row][len(g[row])-1] != 0 {
		t.Errorf("3 o'clock = %d, want section 0 (a third is 120 degrees)", g[row][len(g[row])-1])
	}
}

func TestFrameClockwiseFromTop(t *testing.T) {
	m := New(4, WithFrame(20, 4, 2, 3))
	m.SetSections(sectionsOf(1, 1, 1, 1)...)
	g := m.grid()
	mid := len(g[0]) / 2
	if g[0][mid] != 0 || g[0][mid-1] != 3 {
		t.Errorf("top center neighbours = %d, %d; want 0 and 3", g[0][mid], g[0][mid-1])
	}
}

func shapeCount(g [][]int) int {
	n := 0
	for _, row := range g {
		for _, s := range row {
			if s >= 0 {
				n++
			}
		}
	}
	return n
}

func checkProportions(t *testing.T, m *Model, vals []float64) {
	t.Helper()
	m.SetSections(sectionsOf(vals...)...)
	g := m.grid()
	total, sum := float64(shapeCount(g)), 0.0
	for _, v := range vals {
		sum += v
	}
	c := count(g)
	for i, v := range vals {
		got, want := float64(c[i])/total, v/sum
		if math.Abs(got-want) > 0.015 {
			t.Errorf("section %d share = %.4f, want %.4f", i, got, want)
		}
	}
}

func TestProportions(t *testing.T) {
	for _, gs := range []GlyphSet{Braille, Octants, HalfBlocks} {
		checkProportions(t, New(20, WithGlyphs(gs)), []float64{52, 8, 4})
		checkProportions(t, New(20, WithGlyphs(gs), WithFrame(60, 16, 4, 6)), []float64{52, 8, 4, 30})
	}
}

func TestMinimumArc(t *testing.T) {
	m := New(10)
	m.SetSections(sectionsOf(499, 1)...)
	g := m.grid()
	c := count(g)
	// 2 units of the middle line of length pi*(R+r) = pi*16
	want := 2 / (math.Pi * 16) * float64(shapeCount(g))
	if got := float64(c[1]); math.Abs(got-want) > want*0.25 {
		t.Errorf("tiny section has %v sub-pixels, want about %.0f", got, want)
	}
}

func TestZeroAndInvalidSectionsDrawNothing(t *testing.T) {
	m := New(6)
	m.SetSections(sectionsOf(5, 0, -3, math.NaN(), math.Inf(1), 5)...)
	c := count(m.grid())
	for i := 1; i <= 4; i++ {
		if c[i] != 0 {
			t.Errorf("section %d drew %d sub-pixels", i, c[i])
		}
	}
	if c[0] == 0 || c[5] == 0 {
		t.Errorf("valid sections missing: %v", c)
	}
}

func TestMinimumsOverflowKeepEverySectionVisible(t *testing.T) {
	m := New(1) // middle line is about 5 units: three minimums of 2 units exceed it
	m.SetSections(sectionsOf(1000, 1, 1)...)
	c := count(m.grid())
	if c[1] == 0 || c[2] == 0 {
		t.Errorf("small sections vanished: %v", c)
	}
}

func sideThickness(g [][]int) (top, bottom, left, right int) {
	h, w := len(g), len(g[0])
	for r := 0; r < h && g[r][w/2] >= 0; r++ {
		top++
	}
	for r := h - 1; r >= 0 && g[r][w/2] >= 0; r-- {
		bottom++
	}
	for c := 0; c < w && g[h/2][c] >= 0; c++ {
		left++
	}
	for c := w - 1; c >= 0 && g[h/2][c] >= 0; c-- {
		right++
	}
	return
}

func TestFrameThicknessEqualOnAllSides(t *testing.T) {
	cols := map[GlyphSet]int{Braille: 2, Octants: 2, HalfBlocks: 1}
	for gs, sx := range cols {
		m := New(8, WithGlyphs(gs), WithFrame(30, 8, 3, 0))
		m.SetSections(sectionsOf(1)...)
		top, bottom, left, right := sideThickness(m.grid())
		for _, got := range []int{top, bottom, left, right} {
			if got != 3*sx {
				t.Errorf("glyph set %d: thickness %d,%d,%d,%d, want %d on every side", gs, top, bottom, left, right, 3*sx)
				break
			}
		}
	}
}

func TestFrameCorners(t *testing.T) {
	square := New(8, WithFrame(30, 8, 3, 0))
	square.SetSections(sectionsOf(1)...)
	if square.grid()[0][0] < 0 {
		t.Error("radius 0: corner sub-pixel is empty")
	}
	round := New(8, WithFrame(30, 8, 3, 8))
	round.SetSections(sectionsOf(1)...)
	g := round.grid()
	w, h := len(g[0]), len(g)
	for _, p := range [][2]int{{0, 0}, {0, w - 1}, {h - 1, 0}, {h - 1, w - 1}} {
		if g[p[0]][p[1]] >= 0 {
			t.Errorf("radius 8: corner sub-pixel %v is filled", p)
		}
	}
}

func TestFrameClampsOptions(t *testing.T) {
	m := New(4, WithFrame(0, 0, 0, -5))
	m.SetSections(sectionsOf(1)...)
	if g := m.grid(); len(g) == 0 || len(g[0]) == 0 {
		t.Error("empty grid for a degenerate frame")
	}
	big := New(4, WithFrame(10, 4, 99, 99))
	big.SetSections(sectionsOf(1, 2)...)
	if shapeCount(big.grid()) == 0 {
		t.Error("oversized thickness and radius draw nothing")
	}
}

func TestFrameMiddleLineOrder(t *testing.T) {
	m := New(6, WithFrame(20, 6, 2, 4))
	m.SetSections(sectionsOf(1, 1, 1, 1)...)
	g := m.grid()
	// walking clockwise from the top center the section never goes backwards
	h, w := len(g), len(g[0])
	var seq []int
	for c := w / 2; c < w; c++ {
		seq = append(seq, g[0][c])
	}
	for r := 1; r < h; r++ {
		seq = append(seq, g[r][w-1])
	}
	for c := w - 2; c >= 0; c-- {
		seq = append(seq, g[h-1][c])
	}
	for r := h - 2; r >= 0; r-- {
		seq = append(seq, g[r][0])
	}
	for c := 1; c < w/2; c++ {
		seq = append(seq, g[0][c])
	}
	last := 0
	for _, s := range seq {
		if s < 0 {
			continue
		}
		if s < last {
			t.Fatalf("section order goes backwards: %v", seq)
		}
		last = s
	}
	if last != 3 {
		t.Errorf("walk ended in section %d, want 3", last)
	}
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestSharesEqualWhenMinimumsOverflow(t *testing.T) {
	got := shares([]float64{1000, 1, 1}, 0.4)
	for i, s := range got {
		if !near(s, 1.0/3) {
			t.Errorf("share %d = %v, want 1/3", i, got[i])
		}
	}
}

func TestSharesCascadePinsLaterSections(t *testing.T) {
	got := shares([]float64{100, 7, 1, 1}, 0.06)
	want := []float64{0.82, 0.06, 0.06, 0.06}
	sum := 0.0
	for i := range want {
		sum += got[i]
		if !near(got[i], want[i]) {
			t.Errorf("shares = %v, want %v", got, want)
			break
		}
	}
	if !near(sum, 1) {
		t.Errorf("shares sum to %v", sum)
	}
}

func TestSharesSkipUndrawable(t *testing.T) {
	got := shares([]float64{0, 5, math.NaN(), 5}, 0.1)
	if got[0] != 0 || got[2] != 0 || !near(got[1], 0.5) || !near(got[3], 0.5) {
		t.Errorf("shares = %v", got)
	}
}
