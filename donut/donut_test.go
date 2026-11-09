package donut_test

import (
	"math"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/ingvarch/pulse/donut"
)

var (
	green  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
	yellow = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffff00"))
	red    = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
)

const redSeq = "38;2;255;0;0"

func pods() []donut.Section {
	return []donut.Section{
		{Name: "Running", Value: 52, Style: green},
		{Name: "Pending", Value: 8, Style: yellow},
		{Name: "Failed", Value: 4, Style: red},
	}
}

func plainRows(view string) [][]rune {
	var rows [][]rune
	for line := range strings.SplitSeq(ansi.Strip(view), "\n") {
		rows = append(rows, []rune(line))
	}
	return rows
}

var glyphSets = []donut.GlyphSet{donut.Braille, donut.Octants, donut.HalfBlocks}

func TestViewHeightAndWidth(t *testing.T) {
	for _, gs := range glyphSets {
		shapes := map[string]*donut.Model{
			"ring":  donut.New(12, donut.WithGlyphs(gs), donut.WithCenterLabel("Total")),
			"frame": donut.New(12, donut.WithGlyphs(gs), donut.WithFrame(40, 9, 3, 5), donut.WithCenterLabel("Total")),
		}
		wantW := map[string]int{"ring": 24, "frame": 40}
		wantH := map[string]int{"ring": 12, "frame": 9}
		for name, m := range shapes {
			for _, filled := range []bool{false, true} {
				if filled {
					m.SetSections(pods()...)
				}
				lines := strings.Split(m.View(), "\n")
				if len(lines) != wantH[name] {
					t.Errorf("%s glyphs=%d filled=%v: %d lines, want %d", name, gs, filled, len(lines), wantH[name])
				}
				for i, l := range lines {
					if w := lipgloss.Width(l); w != wantW[name] {
						t.Errorf("%s glyphs=%d filled=%v line %d: width %d, want %d", name, gs, filled, i, w, wantW[name])
					}
				}
			}
		}
	}
}

func TestNewClampsHeight(t *testing.T) {
	if got := strings.Count(donut.New(0).View(), "\n") + 1; got != 1 {
		t.Errorf("height 0 gives %d lines, want 1", got)
	}
}

func TestSetSectionsCopiesSlice(t *testing.T) {
	s := pods()
	m := donut.New(8)
	m.SetSections(s...)
	before := m.Legend()
	s[0].Value = 1000
	if m.Legend() != before {
		t.Error("changing the caller's slice changed the donut")
	}
}

func TestGlyphsGetterAndSetter(t *testing.T) {
	m := donut.New(4)
	if m.Glyphs() != donut.Braille {
		t.Errorf("default glyphs = %d, want Braille", m.Glyphs())
	}
	m.SetGlyphs(donut.Octants)
	if m.Glyphs() != donut.Octants {
		t.Error("SetGlyphs did not stick")
	}
}

func TestCenterTotalAndLabel(t *testing.T) {
	m := donut.New(12, donut.WithCenterLabel("Total"))
	m.SetSections(pods()...)
	rows := plainRows(m.View())
	if got := string(rows[5][11:13]); got != "64" {
		t.Errorf("total = %q at row 5, col 11", got)
	}
	if got := string(rows[6][9:14]); got != "Total" {
		t.Errorf("label = %q at row 6, col 9", got)
	}
	if !strings.Contains(m.View(), "\x1b[1m") {
		t.Error("total is not bold")
	}
}

func TestCenterTotalAloneIsCentered(t *testing.T) {
	m := donut.New(12)
	m.SetSections(pods()...)
	rows := plainRows(m.View())
	if got := string(rows[5][11:13]); got != "64" {
		t.Errorf("total = %q at row 5, col 11", got)
	}
}

func TestCenterDropsLabelOnly(t *testing.T) {
	m := donut.New(4, donut.WithCenterLabel("Total"))
	m.SetSections(donut.Section{Name: "a", Value: 60}, donut.Section{Name: "b", Value: 4})
	view := ansi.Strip(m.View())
	if strings.Contains(view, "Total") {
		t.Error("label should not fit in a ring of height 4")
	}
	rows := plainRows(m.View())
	if got := string(rows[1][3:5]); got != "64" {
		t.Errorf("total = %q at row 1, col 3", got)
	}
}

func TestCenterDropsBothWhenHoleIsTooSmall(t *testing.T) {
	m := donut.New(2, donut.WithCenterLabel("Total"))
	m.SetSections(donut.Section{Name: "a", Value: 60}, donut.Section{Name: "b", Value: 4})
	view := ansi.Strip(m.View())
	if strings.Contains(view, "64") || strings.Contains(view, "Total") {
		t.Errorf("center text should be dropped:\n%s", view)
	}
}

func TestCenterDropsLabelWhenTotalDoesNotFit(t *testing.T) {
	m := donut.New(8, donut.WithCenterLabel("T"))
	m.SetSections(donut.Section{Name: "a", Value: 0.1}, donut.Section{Name: "b", Value: 0.2}) // total 0.30000000000000004
	view := ansi.Strip(m.View())
	if strings.Contains(view, "0.3") || strings.Contains(view, "T") {
		t.Errorf("center text should be dropped:\n%s", view)
	}
}

func TestZeroTotalShowsZeroAndFaintShape(t *testing.T) {
	for _, secs := range [][]donut.Section{nil, {{Name: "a", Value: 0, Style: green}, {Name: "b", Value: -2, Style: red}}} {
		m := donut.New(12)
		m.SetSections(secs...)
		view := m.View()
		rows := plainRows(view)
		if rows[5][11] != '0' {
			t.Errorf("total 0 not shown: %q", string(rows[5]))
		}
		if !strings.Contains(view, "\x1b[2m") {
			t.Error("shape is not faint")
		}
		if strings.Contains(view, "38;2;") {
			t.Error("shape has a color")
		}
	}
}

func TestTinySectionIsVisibleInEveryGlyphSet(t *testing.T) {
	for _, gs := range glyphSets {
		shapes := map[string]*donut.Model{
			"ring":  donut.New(10, donut.WithGlyphs(gs)),
			"frame": donut.New(10, donut.WithGlyphs(gs), donut.WithFrame(40, 10, 3, 4)),
		}
		for name, m := range shapes {
			m.SetSections(
				donut.Section{Name: "ok", Value: 499, Style: green},
				donut.Section{Name: "bad", Value: 1, Style: red},
			)
			if !strings.Contains(m.View(), redSeq) {
				t.Errorf("%s glyphs=%d: section 1 of 500 has no cell of its color", name, gs)
			}
		}
	}
}

func TestZeroSectionDrawsNothingButIsInLegend(t *testing.T) {
	m := donut.New(10)
	m.SetSections(
		donut.Section{Name: "ok", Value: 5, Style: green},
		donut.Section{Name: "none", Value: 0, Style: red},
	)
	if strings.Contains(m.View(), redSeq) {
		t.Error("zero section was drawn")
	}
	if !strings.Contains(ansi.Strip(m.Legend()), "none 0") {
		t.Errorf("legend = %q", ansi.Strip(m.Legend()))
	}
}

func TestLegend(t *testing.T) {
	m := donut.New(8)
	if m.Legend() != "" {
		t.Errorf("empty legend = %q", m.Legend())
	}
	m.SetSections(pods()...)
	want := "● Running 52   ● Pending 8   ● Failed 4"
	if got := ansi.Strip(m.Legend()); got != want {
		t.Errorf("legend = %q, want %q", got, want)
	}
	if !strings.Contains(m.Legend(), "38;2;0;255;0") {
		t.Error("dot is not in the section style")
	}
	m.SetSections(donut.Section{Name: "x", Value: 1.5}, donut.Section{Name: "y", Value: math.NaN()}, donut.Section{Name: "z", Value: -2})
	if got, want := ansi.Strip(m.Legend()), "● x 1.5   ● y NaN   ● z -2"; got != want {
		t.Errorf("legend = %q, want %q", got, want)
	}
}

func TestEveryGlyphSetDrawsItsOwnCharacters(t *testing.T) {
	probes := map[donut.GlyphSet]func(r rune) bool{
		donut.Braille:    func(r rune) bool { return r >= 0x2800 && r <= 0x28FF },
		donut.HalfBlocks: func(r rune) bool { return r == '▀' || r == '▄' || r == '█' },
		donut.Octants:    func(r rune) bool { return r >= 0x1CD00 && r <= 0x1CDE5 },
	}
	for gs, ok := range probes {
		m := donut.New(8, donut.WithGlyphs(gs))
		m.SetSections(pods()...)
		found := false
		for _, r := range ansi.Strip(m.View()) {
			if ok(r) {
				found = true
			}
		}
		if !found {
			t.Errorf("glyph set %d drew none of its characters", gs)
		}
	}
}

func TestMultilineLabelKeepsViewShape(t *testing.T) {
	m := donut.New(12, donut.WithCenterLabel("a\nb"))
	m.SetSections(pods()...)
	rows := strings.Split(m.View(), "\n")
	if len(rows) != 12 {
		t.Fatalf("lines = %d, want 12", len(rows))
	}
	for i, r := range rows {
		if w := lipgloss.Width(r); w != 24 {
			t.Errorf("line %d width = %d, want 24", i, w)
		}
	}
	if !strings.Contains(ansi.Strip(m.View()), "64") {
		t.Error("total should stay")
	}
}

func TestTabLabelKeepsViewWidth(t *testing.T) {
	m := donut.New(12, donut.WithCenterLabel("a\tb"))
	m.SetSections(pods()...)
	for i, r := range strings.Split(m.View(), "\n") {
		if w := lipgloss.Width(r); w != 24 {
			t.Errorf("line %d width = %d, want 24", i, w)
		}
	}
}

func TestBackgroundOnlyInTwoColorGlyphSets(t *testing.T) {
	for _, gs := range glyphSets {
		m := donut.New(10, donut.WithGlyphs(gs))
		m.SetSections(pods()...)
		got := strings.Contains(m.View(), "48;2;")
		if want := gs != donut.Braille; got != want {
			t.Errorf("glyph set %d: background = %v, want %v", gs, got, want)
		}
	}
}

func TestUnknownGlyphSetIsBraille(t *testing.T) {
	m := donut.New(4, donut.WithGlyphs(donut.GlyphSet(7)))
	if m.Glyphs() != donut.Braille {
		t.Errorf("WithGlyphs(7) = %d, want Braille", m.Glyphs())
	}
	m.SetGlyphs(donut.GlyphSet(-1))
	if m.Glyphs() != donut.Braille {
		t.Errorf("SetGlyphs(-1) = %d, want Braille", m.Glyphs())
	}
}

func TestFrameCenterTotalAndLabel(t *testing.T) {
	m := donut.New(7, donut.WithFrame(30, 7, 2, 2), donut.WithCenterLabel("Total"))
	m.SetSections(pods()...)
	rows := plainRows(m.View())
	if got := string(rows[2][14:16]); got != "64" {
		t.Errorf("total = %q at row 2, col 14", got)
	}
	if got := string(rows[3][12:17]); got != "Total" {
		t.Errorf("label = %q at row 3, col 12", got)
	}
}
