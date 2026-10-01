package pulse_test

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/ingvarch/pulse"
)

func TestZeroBaseline_AxisAndRulings(t *testing.T) {
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
	m := pulse.New(20, 9,
		pulse.WithRange(-50, 50),
		pulse.WithZeroBaseline(true),
		pulse.WithLineStyle(green),
		pulse.WithFill(true),
	)

	if !m.ZeroBaseline() {
		t.Fatal("expected ZeroBaseline() to be true")
	}

	// 1. Initial empty view: should contain ├ at the zero row
	view := m.View()
	if !strings.Contains(view, "├") {
		t.Fatalf("expected view to contain '├' on axis at zero baseline, got:\n%s", view)
	}

	// 2. Push positive values: area fill must produce ┴ on zero baseline row
	for i := 0; i < 5; i++ {
		m.Push(35)
	}
	posView := m.View()
	if !strings.Contains(posView, "┴") {
		t.Fatalf("expected view with positive fill to contain '┴' on zero baseline, got:\n%s", posView)
	}

	// 3. Clear and push negative values: area fill must produce ┬ on zero baseline row
	m2 := pulse.New(20, 9,
		pulse.WithRange(-50, 50),
		pulse.WithZeroBaseline(true),
		pulse.WithLineStyle(green),
		pulse.WithFill(true),
	)
	for i := 0; i < 5; i++ {
		m2.Push(-35)
	}
	negView := m2.View()
	if !strings.Contains(negView, "┬") {
		t.Fatalf("expected view with negative fill to contain '┬' on zero baseline, got:\n%s", negView)
	}
}

func TestZeroBaseline_OutsideRangeDoesNotRender(t *testing.T) {
	// Range strictly above zero: zero baseline should not activate
	mPos := pulse.New(20, 8,
		pulse.WithRange(20, 100),
		pulse.WithZeroBaseline(true),
	)
	if strings.Contains(mPos.View(), "├") {
		t.Fatalf("expected no '├' when 0 is outside positive range [20, 100], got:\n%s", mPos.View())
	}

	// Range strictly below zero: zero baseline should not activate
	mNeg := pulse.New(20, 8,
		pulse.WithRange(-100, -20),
		pulse.WithZeroBaseline(true),
	)
	if strings.Contains(mNeg.View(), "├") {
		t.Fatalf("expected no '├' when 0 is outside negative range [-100, -20], got:\n%s", mNeg.View())
	}
}

func TestNegativeStyle_LinesRenderer(t *testing.T) {
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))

	m := pulse.New(15, 9,
		pulse.WithLineWidth(1),
		pulse.WithRange(-50, 50),
		pulse.WithLineStyle(green),
		pulse.WithNegativeStyle(red),
		pulse.WithZeroBaseline(true),
		pulse.WithFill(true),
	)

	if m.NegativeStyle().GetForeground() != red.GetForeground() {
		t.Fatal("NegativeStyle() did not retain style")
	}

	// Push sequence: positive then negative
	m.Push(30)
	m.Push(30)
	m.Push(-30)
	m.Push(-30)

	view := m.View()

	// Both green (positive) and red (negative) must be present in the output
	if !strings.Contains(view, green.Render("─")) && !strings.Contains(view, green.Render("╭")) {
		t.Fatalf("expected positive line styled in green, got:\n%s", view)
	}
	if !strings.Contains(view, red.Render("─")) && !strings.Contains(view, red.Render("╰")) {
		t.Fatalf("expected negative line styled in red, got:\n%s", view)
	}
}

func TestSeriesNegativeStyle(t *testing.T) {
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))

	m := pulse.New(15, 9,
		pulse.WithLineWidth(1),
		pulse.WithRange(-50, 50),
		pulse.WithSeriesStyle("pnl", green),
		pulse.WithSeriesNegativeStyle("pnl", red),
		pulse.WithFill(true),
	)

	if m.SeriesNegativeStyle("pnl").GetForeground() != red.GetForeground() {
		t.Fatal("SeriesNegativeStyle(\"pnl\") did not retain style")
	}

	m.PushSeries("pnl", 25)
	m.PushSeries("pnl", 25)
	m.PushSeries("pnl", -25)
	m.PushSeries("pnl", -25)

	view := m.View()
	if !strings.Contains(view, red.Render("─")) && !strings.Contains(view, red.Render("╰")) {
		t.Fatalf("expected negative line in series 'pnl' styled in red, got:\n%s", view)
	}
}

func TestSeriesInverted_RXTX(t *testing.T) {
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a"))
	purple := lipgloss.NewStyle().Foreground(lipgloss.Color("#bb9af7"))

	m := pulse.New(20, 9,
		pulse.WithRange(-100, 100),
		pulse.WithZeroBaseline(true),
		pulse.WithSeriesStyle("rx", green),
		pulse.WithSeriesStyle("tx", purple),
		pulse.WithSeriesInverted("tx", true),
		pulse.WithFill(true),
	)

	if !m.SeriesInverted("tx") {
		t.Fatal("expected SeriesInverted(\"tx\") to be true")
	}
	if m.SeriesInverted("rx") {
		t.Fatal("expected SeriesInverted(\"rx\") to be false")
	}

	// Push raw positive values for both
	m.PushSeries("rx", 60)
	m.PushSeries("tx", 40)

	// Last must preserve raw pushed value (40, not -40)
	if lastTX, ok := m.Last("tx"); !ok || lastTX != 40 {
		t.Fatalf("m.Last(\"tx\") = %v, want 40", lastTX)
	}
	if lastRX, ok := m.Last("rx"); !ok || lastRX != 60 {
		t.Fatalf("m.Last(\"rx\") = %v, want 60", lastRX)
	}

	view := m.View()
	// Both colors must be present: green above, purple below
	lines := strings.Split(strings.TrimRight(view, "\n"), "\n")
	// Row 0..3 are positive rows (rx), row 4 is zero, row 5..8 are negative rows (tx)
	posHalf := strings.Join(lines[:4], "\n")
	negHalf := strings.Join(lines[5:], "\n")

	if !strings.Contains(posHalf, "\x1b[38;2;158;206;106m") && !strings.Contains(posHalf, "9ece6a") {
		// check if green is in positive half
		t.Fatalf("positive half must contain rx green style, got:\n%s", posHalf)
	}
	if !strings.Contains(negHalf, "\x1b[38;2;187;154;247m") && !strings.Contains(negHalf, "bb9af7") {
		// check if purple is in negative half
		t.Fatalf("negative half must contain tx purple style, got:\n%s", negHalf)
	}
}

func TestBrailleRenderer_BidirectionalFill(t *testing.T) {
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("#7dcfff"))

	withFill := pulse.New(20, 8,
		pulse.WithMode(pulse.ModeBraille),
		pulse.WithRange(-50, 50),
		pulse.WithLineStyle(cyan),
		pulse.WithFill(true),
	)
	noFill := pulse.New(20, 8,
		pulse.WithMode(pulse.ModeBraille),
		pulse.WithRange(-50, 50),
		pulse.WithLineStyle(cyan),
		pulse.WithFill(false),
	)

	for i := 0; i < 5; i++ {
		withFill.Push(-30)
		noFill.Push(-30)
	}

	viewWith := withFill.View()
	viewNo := noFill.View()

	// When fill is enabled for negative values in Braille, viewWith must differ from viewNo
	if viewWith == viewNo {
		t.Fatalf("expected Braille view with negative fill to differ from no-fill view, but they are identical:\n%s", viewWith)
	}
}

func TestBrailleRenderer_NegativeStyle(t *testing.T) {
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))

	m := pulse.New(15, 8,
		pulse.WithMode(pulse.ModeBraille),
		pulse.WithRange(-50, 50),
		pulse.WithLineStyle(green),
		pulse.WithNegativeStyle(red),
	)

	m.Push(30)
	m.Push(-30)

	view := m.View()
	lines := strings.Split(strings.TrimRight(view, "\n"), "\n")
	negHalf := strings.Join(lines[len(lines)/2:], "\n")

	if !strings.Contains(negHalf, "\x1b[38;2;255;0;0m") && !strings.Contains(negHalf, "ff0000") {
		t.Fatalf("expected red style in negative braille rows, got:\n%s", negHalf)
	}
}

func TestZeroCrossing_RoundedCorners(t *testing.T) {
	// A chart with thin lines (lineWidth = 1) and smooth = true (default)
	m := pulse.New(15, 9,
		pulse.WithLineWidth(1),
		pulse.WithRange(-50, 50),
		pulse.WithZeroBaseline(true),
		pulse.WithFill(false),
	)

	// Stream positive crest: 0 -> 30 -> 30 -> 0 => rounded corners '╭' and '╮'
	m.Push(0)
	m.Push(30)
	m.Push(30)
	m.Push(0)

	viewPos := m.View()
	if !strings.Contains(viewPos, "╭") {
		t.Fatalf("expected positive crest to contain rounded corner '╭', got:\n%s", viewPos)
	}
	if !strings.Contains(viewPos, "╮") {
		t.Fatalf("expected positive crest to contain rounded corner '╮', got:\n%s", viewPos)
	}

	// Stream negative trough: 0 -> -30 -> -30 -> 0 => rounded corners '╰' and '╯'
	m2 := pulse.New(15, 9,
		pulse.WithLineWidth(1),
		pulse.WithRange(-50, 50),
		pulse.WithZeroBaseline(true),
		pulse.WithFill(false),
	)
	m2.Push(0)
	m2.Push(-30)
	m2.Push(-30)
	m2.Push(0)

	viewNeg := m2.View()
	if !strings.Contains(viewNeg, "╰") {
		t.Fatalf("expected negative trough to contain rounded corner '╰', got:\n%s", viewNeg)
	}
	if !strings.Contains(viewNeg, "╯") {
		t.Fatalf("expected negative trough to contain rounded corner '╯', got:\n%s", viewNeg)
	}

	// Sharp mode: WithSmooth(false) => '┌', '┐', '└', '┘' instead of '╭', '╮', '╰', '╯'
	mSharp := pulse.New(15, 9,
		pulse.WithLineWidth(1),
		pulse.WithSmooth(false),
		pulse.WithRange(-50, 50),
		pulse.WithZeroBaseline(true),
		pulse.WithFill(false),
	)
	mSharp.Push(0)
	mSharp.Push(30)
	mSharp.Push(30)
	mSharp.Push(-30)
	mSharp.Push(-30)
	mSharp.Push(0)

	viewSharp := mSharp.View()
	if strings.Contains(viewSharp, "╭") || strings.Contains(viewSharp, "╰") {
		t.Fatalf("sharp mode must not contain rounded corners, got:\n%s", viewSharp)
	}
	if !strings.Contains(viewSharp, "┌") || !strings.Contains(viewSharp, "┘") {
		t.Fatalf("sharp mode must contain sharp corners '┌' and '┘', got:\n%s", viewSharp)
	}
}

func TestSeriesInverted_RXTX_RoundedCorners(t *testing.T) {
	m := pulse.New(15, 9,
		pulse.WithLineWidth(1),
		pulse.WithRange(-50, 50),
		pulse.WithZeroBaseline(true),
		pulse.WithSeriesInverted("tx", true),
	)

	// Stream RX peak (positive crest)
	m.PushSeries("rx", 10)
	m.PushSeries("rx", 35)
	m.PushSeries("rx", 35)
	m.PushSeries("rx", 10)

	// Stream TX peak (raw positive 10 -> 35 -> 35 -> 10, inverted to negative trough)
	m.PushSeries("tx", 10)
	m.PushSeries("tx", 35)
	m.PushSeries("tx", 35)
	m.PushSeries("tx", 10)

	view := m.View()
	t.Logf("RX/TX View:\n%s", view)
	if !strings.Contains(view, "╭") || !strings.Contains(view, "╮") {
		t.Fatalf("expected RX to produce rounded top corners '╭' and '╮', got:\n%s", view)
	}
	if !strings.Contains(view, "╰") || !strings.Contains(view, "╯") {
		t.Fatalf("expected TX to produce rounded bottom corners '╰' and '╯', got:\n%s", view)
	}
}

