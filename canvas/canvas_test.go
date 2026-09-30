package canvas_test

import (
	"strings"
	"testing"

	"github.com/ingvarch/pulse/canvas"
)

func TestCanvas_SetAndView(t *testing.T) {
	c := canvas.New(5, 3)
	c.Set(0, 0, 'X')
	c.Set(4, 2, 'Y')
	view := c.View()
	if !strings.Contains(view, "X") {
		t.Fatalf("View() must contain X, got:\n%s", view)
	}
	if !strings.Contains(view, "Y") {
		t.Fatalf("View() must contain Y, got:\n%s", view)
	}
	lines := strings.Split(strings.TrimRight(view, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d:\n%s", len(lines), view)
	}
}

func TestCanvas_OutOfBoundsIsNoop(t *testing.T) {
	c := canvas.New(2, 2)
	c.Set(-1, 0, 'X')
	c.Set(0, 5, 'X')
	_ = c.View()
}
