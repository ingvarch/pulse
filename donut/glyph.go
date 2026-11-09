package donut

import "slices"

// GlyphSet chooses the characters that draw the cells.
type GlyphSet int

const (
	// Braille draws 2x4 dots with one color per cell. It works in nearly
	// every terminal. It is the default.
	Braille GlyphSet = iota
	// Octants draws 2x4 sub-pixels with two colors per cell, from the
	// Unicode 16 octant block. The terminal or its font must have it.
	Octants
	// HalfBlocks draws 1x2 sub-pixels with two colors per cell, with the
	// characters ▀ ▄ █. It works everywhere.
	HalfBlocks
)

// grid returns the sub-pixel columns and rows of one cell.
func (g GlyphSet) grid() (cols, rows int) {
	if g == HalfBlocks {
		return 1, 2
	}
	return 2, 4
}

// brailleBit is the dot bit of each sub-pixel, numbered row by row.
var brailleBit = [8]rune{0x01, 0x08, 0x02, 0x10, 0x04, 0x20, 0x40, 0x80}

func brailleRune(pattern uint8) rune {
	r := rune(0x2800)
	for k, bit := range brailleBit {
		if pattern&(1<<k) != 0 {
			r |= bit
		}
	}
	return r
}

func halfBlockRune(pattern uint8) rune {
	return [4]rune{' ', '▀', '▄', '█'}[pattern&3]
}

// oldOctants are the octant patterns that already have a block character.
var oldOctants = map[uint8]rune{
	0: ' ', 255: 0x2588, 15: 0x2580, 240: 0x2584, 85: 0x258C, 170: 0x2590,
	5: 0x2598, 10: 0x259D, 80: 0x2596, 160: 0x2597, 165: 0x259A, 90: 0x259E,
	245: 0x2599, 95: 0x259B, 175: 0x259C, 250: 0x259F,
	192: 0x2582, 252: 0x2586, 3: 0x1FB82, 63: 0x1FB85, 20: 0x1FBE6, 40: 0x1FBE7,
	1: 0x1CEA8, 2: 0x1CEAB, 64: 0x1CEA3, 128: 0x1CEA0,
}

// octantRunes maps a pattern, with bit n-1 set for octant n, to its character.
var octantRunes = func() (t [256]rune) {
	next := rune(0x1CD00)
	for p := range t {
		if r, ok := oldOctants[uint8(p)]; ok {
			t[p] = r
			continue
		}
		t[p] = next
		next++
	}
	return t
}()

func octantRune(pattern uint8) rune { return octantRunes[pattern] }

// pick is the drawing of one cell: the sub-pixels in the pattern (bit k is
// sub-pixel k, row by row) in the fg section's color on the bg section's
// color. A section of -1 means no color.
type pick struct {
	pattern uint8
	fg, bg  int
}

// candidate is a section, or the empty sub-pixels (id -1), and its sub-pixel count.
type candidate struct{ id, n int }

// pickCell chooses the pattern and colors for the sub-pixels of a cell. Each
// entry is a section index or -1 for empty.
func pickCell(sub []int, twoColor bool) pick {
	counts := map[int]int{}
	for _, s := range sub {
		counts[s]++
	}
	var sections []int
	for s := range counts {
		if s >= 0 {
			sections = append(sections, s)
		}
	}
	slices.Sort(sections)
	if len(sections) == 0 {
		return pick{fg: -1, bg: -1}
	}
	drawn := func(s int) bool { return s >= 0 }

	// Candidates in tie order: empty first, then the sections.
	cands := make([]candidate, 0, len(sections)+1)
	if counts[-1] > 0 {
		cands = append(cands, candidate{-1, counts[-1]})
	}
	for _, s := range sections {
		cands = append(cands, candidate{s, counts[s]})
	}
	slices.SortStableFunc(cands, func(a, b candidate) int { return b.n - a.n })

	if !twoColor {
		return pick{pattern: patternOf(sub, drawn), fg: dominant(cands), bg: -1}
	}
	first, second := cands[0].id, -1
	if len(cands) > 1 {
		second = cands[1].id
	}
	switch {
	case first < 0:
		return pick{pattern: patternOf(sub, drawn), fg: second, bg: -1}
	case second < 0:
		return pick{pattern: patternOf(sub, drawn), fg: first, bg: -1}
	}
	// The other candidates join the first one.
	return pick{pattern: patternOf(sub, func(s int) bool { return s != second }), fg: first, bg: second}
}

// dominant returns the section with the most sub-pixels; candidates are in tie order.
func dominant(cands []candidate) int {
	for _, c := range cands {
		if c.id >= 0 {
			return c.id
		}
	}
	return -1
}

func patternOf(sub []int, in func(int) bool) uint8 {
	var p uint8
	for k, s := range sub {
		if in(s) {
			p |= 1 << k
		}
	}
	return p
}

// glyphFor returns the character for a cell. Two sections on a half-block
// cell always use the upper half, with the top section as the foreground.
func glyphFor(g GlyphSet, p pick) (rune, pick) {
	switch g {
	case Octants:
		return octantRune(p.pattern), p
	case HalfBlocks:
		if p.bg >= 0 && p.pattern == 2 {
			p.pattern, p.fg, p.bg = 1, p.bg, p.fg
		}
		return halfBlockRune(p.pattern), p
	}
	return brailleRune(p.pattern), p
}
