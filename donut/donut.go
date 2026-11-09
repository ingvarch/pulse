// Package donut draws a ring or a rectangular frame split into sections, to
// show the parts of a whole in the terminal.
package donut

import (
	"math"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
)

// Section is one part of the whole.
type Section struct {
	// Name is shown in the legend.
	Name string
	// Value is the size of the section. A value of 0 or less, NaN or Inf is
	// not drawn but stays in the legend.
	Value float64
	// Style supplies the foreground color of the section.
	Style lipgloss.Style
}

// Model is a donut chart. Create it with New.
type Model struct {
	h        int
	label    string
	glyphs   GlyphSet
	frame    *frame
	sections []Section
}

// Option configures a Model in New.
type Option func(*Model)

// New returns a donut h rows tall, which makes a ring 2h columns wide.
// h below 1 is raised to 1.
func New(h int, opts ...Option) *Model {
	m := &Model{h: max(h, 1)}
	for _, o := range opts {
		o(m)
	}
	return m
}

// WithCenterLabel sets the text under the total in the middle. It is dropped
// when it does not fit in the hole or has more than one line.
func WithCenterLabel(s string) Option { return func(m *Model) { m.label = s } }

// WithGlyphs sets the glyph set. The default is Braille.
func WithGlyphs(g GlyphSet) Option { return func(m *Model) { m.SetGlyphs(g) } }

// WithFrame replaces the ring by a frame w columns wide and h rows tall.
// The frame's w and h set the size, and the h given to New is not used.
// thickness and radius are in columns: the border is thickness columns wide
// at the sides and half as many rows tall at the top and bottom. A radius of
// 0 gives square corners. Both are clamped to what fits.
func WithFrame(w, h, thickness, radius int) Option {
	return func(m *Model) {
		f := newFrame(w, h, thickness, radius)
		m.frame = &f
	}
}

// SetSections replaces the sections. It copies the slice.
func (m *Model) SetSections(s ...Section) { m.sections = append([]Section(nil), s...) }

// SetGlyphs changes the glyph set. An unknown value gives Braille.
func (m *Model) SetGlyphs(g GlyphSet) {
	m.glyphs = Braille
	if g == Octants || g == HalfBlocks {
		m.glyphs = g
	}
}

// Glyphs returns the glyph set.
func (m *Model) Glyphs() GlyphSet { return m.glyphs }

// Legend returns one line with a colored dot, the name and the value of
// every section, zero and negative ones included.
func (m *Model) Legend() string {
	entries := make([]string, len(m.sections))
	for i, s := range m.sections {
		entries[i] = s.Style.Render("●") + " " + s.Name + " " + formatValue(s.Value)
	}
	return strings.Join(entries, "   ")
}

func formatValue(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

func (m *Model) shape() shape {
	if m.frame != nil {
		return *m.frame
	}
	return ring{m.h}
}

func drawable(v float64) bool { return v > 0 && !math.IsInf(v, 0) }

// total is the sum of the drawn values.
func (m *Model) total() float64 {
	sum := 0.0
	for _, s := range m.sections {
		if drawable(s.Value) {
			sum += s.Value
		}
	}
	return sum
}
