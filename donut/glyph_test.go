package donut

import (
	"testing"
)

func TestOctantTableDistinctAndSpotChecks(t *testing.T) {
	seen := map[rune]int{}
	for p := 0; p < 256; p++ {
		r := octantRune(uint8(p))
		if q, dup := seen[r]; dup {
			t.Fatalf("patterns %d and %d share rune %U", q, p, r)
		}
		seen[r] = p
	}
	checks := map[uint8]rune{
		4: 0x1CD00, 254: 0x1CDE5, 6: 0x1CD01,
		0: ' ', 255: 0x2588, 15: 0x2580, 240: 0x2584, 85: 0x258C, 170: 0x2590,
		5: 0x2598, 10: 0x259D, 80: 0x2596, 160: 0x2597, 165: 0x259A, 90: 0x259E,
		245: 0x2599, 95: 0x259B, 175: 0x259C, 250: 0x259F,
		192: 0x2582, 252: 0x2586, 3: 0x1FB82, 63: 0x1FB85, 20: 0x1FBE6, 40: 0x1FBE7,
		1: 0x1CEA8, 2: 0x1CEAB, 64: 0x1CEA3, 128: 0x1CEA0,
	}
	for p, want := range checks {
		if got := octantRune(p); got != want {
			t.Errorf("octantRune(%d) = %U, want %U", p, got, want)
		}
	}
	for p := 0; p < 256; p++ {
		r := octantRune(uint8(p))
		if _, special := checks[uint8(p)]; !special && (r < 0x1CD00 || r > 0x1CDE5) {
			t.Errorf("octantRune(%d) = %U outside the octant block", p, r)
		}
	}
}

func TestBrailleBits(t *testing.T) {
	// sub-pixel k is row-major in a 2 x 4 cell
	want := map[int]rune{
		0: 0x2801, 2: 0x2802, 4: 0x2804, 6: 0x2840,
		1: 0x2808, 3: 0x2810, 5: 0x2820, 7: 0x2880,
	}
	for k, w := range want {
		if got := brailleRune(1 << k); got != w {
			t.Errorf("sub-pixel %d = %U, want %U", k, got, w)
		}
	}
	if brailleRune(0) != 0x2800 || brailleRune(255) != 0x28FF {
		t.Error("empty or full braille cell wrong")
	}
}

func TestHalfBlockGlyphs(t *testing.T) {
	want := map[uint8]rune{0: ' ', 1: '▀', 2: '▄', 3: '█'}
	for p, w := range want {
		if got := halfBlockRune(p); got != w {
			t.Errorf("halfBlockRune(%d) = %q, want %q", p, got, w)
		}
	}
}

func TestHalfBlocksTwoSectionsUseUpperHalf(t *testing.T) {
	r, p := glyphFor(HalfBlocks, pick{pattern: 2, fg: 0, bg: 1})
	if r != '▀' || p.fg != 1 || p.bg != 0 {
		t.Errorf("got %q fg=%d bg=%d, want upper half with fg=1 bg=0", r, p.fg, p.bg)
	}
}

func TestPickCellTwoColors(t *testing.T) {
	cases := []struct {
		name string
		sub  []int
		want pick
	}{
		{"two sections", []int{0, 0, 0, 0, 1, 1, 1, 1}, pick{pattern: 15, fg: 0, bg: 1}},
		{"two sections tie goes to section order", []int{1, 1, 1, 1, 0, 0, 0, 0}, pick{pattern: 240, fg: 0, bg: 1}},
		{"three sections, empty ranks above a tied section", []int{0, 0, 1, 1, 2, 0, -1, -1}, pick{pattern: 63, fg: 0, bg: -1}},
		{"three sections without empty", []int{0, 0, 0, 1, 1, 2, 2, 2}, pick{pattern: 31, fg: 0, bg: 2}},
		{"section on empty", []int{-1, -1, 1, 1, 1, 1, -1, -1}, pick{pattern: 60, fg: 1, bg: -1}},
		{"only empty", []int{-1, -1, -1, -1, -1, -1, -1, -1}, pick{pattern: 0, fg: -1, bg: -1}},
		{"one section fills the cell", []int{2, 2, 2, 2, 2, 2, 2, 2}, pick{pattern: 255, fg: 2, bg: -1}},
	}
	for _, c := range cases {
		if got := pickCell(c.sub, true); got != c.want {
			t.Errorf("%s: got %+v, want %+v", c.name, got, c.want)
		}
	}
}

func TestPickCellOneColor(t *testing.T) {
	cases := []struct {
		name string
		sub  []int
		want pick
	}{
		{"dots from every section, the larger wins", []int{0, 1, 1, -1, -1, -1, -1, 1}, pick{pattern: 1 | 2 | 4 | 128, fg: 1, bg: -1}},
		{"tie goes to section order", []int{1, 1, 0, 0, -1, -1, -1, -1}, pick{pattern: 15, fg: 0, bg: -1}},
		{"only empty", []int{-1, -1, -1, -1, -1, -1, -1, -1}, pick{pattern: 0, fg: -1, bg: -1}},
	}
	for _, c := range cases {
		if got := pickCell(c.sub, false); got != c.want {
			t.Errorf("%s: got %+v, want %+v", c.name, got, c.want)
		}
	}
}
