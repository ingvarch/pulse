package donut

import "math"

// The drawing lives in a space where a cell is 1 unit wide and 2 units tall,
// so a unit is square.

// shape is the outline and the clockwise position of a point on it.
type shape interface {
	cells() (w, h int)
	contains(x, y float64) bool
	// fraction is the clockwise distance of the point's nearest spot on the
	// middle line from the top center, as a share of the line length.
	fraction(x, y float64) float64
	midLength() float64
}

type ring struct{ h int }

func (r ring) cells() (int, int) { return 2 * r.h, r.h }

func (r ring) radii() (outer, inner float64) {
	outer = float64(r.h)
	return outer, 0.6 * outer
}

func (r ring) contains(x, y float64) bool {
	outer, inner := r.radii()
	d := math.Hypot(x-outer, y-outer)
	return d >= inner && d <= outer
}

func (r ring) fraction(x, y float64) float64 {
	c, _ := r.radii()
	theta := math.Atan2(x-c, -(y - c))
	if theta < 0 {
		theta += 2 * math.Pi
	}
	return theta / (2 * math.Pi)
}

func (r ring) midLength() float64 {
	outer, inner := r.radii()
	return math.Pi * (outer + inner)
}

type frame struct{ w, h, thickness, radius int }

// newFrame clamps the options to what fits in w columns and h rows.
func newFrame(w, h, thickness, radius int) frame {
	w, h = max(w, 1), max(h, 1)
	half := min(w, 2*h) / 2
	return frame{w, h, min(max(thickness, 1), max(half, 1)), min(max(radius, 0), half)}
}

func (f frame) cells() (int, int) { return f.w, f.h }

// roundRect reports whether the point is inside the rectangle with rounded corners.
func roundRect(x, y, x0, y0, x1, y1, rad float64) bool {
	dx := math.Max(math.Max(x0+rad-x, x-(x1-rad)), 0)
	dy := math.Max(math.Max(y0+rad-y, y-(y1-rad)), 0)
	return x >= x0 && x <= x1 && y >= y0 && y <= y1 && dx*dx+dy*dy <= rad*rad
}

func (f frame) contains(x, y float64) bool {
	w, h, t, r := float64(f.w), float64(2*f.h), float64(f.thickness), float64(f.radius)
	if !roundRect(x, y, 0, 0, w, h, r) {
		return false
	}
	if w <= 2*t || h <= 2*t {
		return true
	}
	return !roundRect(x, y, t, t, w-t, h-t, math.Max(r-t, 0))
}

// path is the middle line: its straight runs and corner radius.
type path struct {
	w, x0, y0, x1, y1, rad float64
}

func (f frame) path() path {
	w, h, t, r := float64(f.w), float64(2*f.h), float64(f.thickness), float64(f.radius)
	m, rad := t/2, math.Max(r-t/2, 0)
	return path{w: w, x0: m + rad, y0: m + rad, x1: w - m - rad, y1: h - m - rad, rad: rad}
}

func (p path) length() float64 {
	return 2*(p.x1-p.x0) + 2*(p.y1-p.y0) + 2*math.Pi*p.rad
}

func (f frame) midLength() float64 { return f.path().length() }

func (f frame) fraction(x, y float64) float64 {
	p := f.path()
	arc := p.arcAt(x, y)
	if arc < 0 {
		arc += p.length()
	}
	return arc / p.length()
}

// arcAt returns the clockwise arc length, from the top center, of the middle
// line point nearest to (x, y).
func (p path) arcAt(x, y float64) float64 {
	quarter := p.rad * math.Pi / 2
	cx := p.w / 2
	right := p.x1 - cx + quarter // end of the top-right corner
	bottom := right + p.y1 - p.y0 + quarter
	left := bottom + p.x1 - p.x0 + quarter
	cl := func(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }

	farX, farY := x < p.x0 || x > p.x1, y < p.y0 || y > p.y1
	if farX && farY {
		switch {
		case x > p.x1 && y < p.y0:
			return p.x1 - cx + p.rad*math.Atan2(x-p.x1, p.y0-y)
		case x > p.x1:
			return right + p.y1 - p.y0 + p.rad*math.Atan2(y-p.y1, x-p.x1)
		case y > p.y1:
			return bottom + p.x1 - p.x0 + p.rad*math.Atan2(p.x0-x, y-p.y1)
		default:
			return left + p.y1 - p.y0 + p.rad*math.Atan2(p.y0-y, p.x0-x)
		}
	}
	// nearest side by distance to its line, in the order top, right, bottom, left
	best, side := y-p.y0+p.rad, 0
	for i, d := range []float64{p.x1 + p.rad - x, p.y1 + p.rad - y, x - p.x0 + p.rad} {
		if d < best {
			best, side = d, i+1
		}
	}
	switch side {
	case 0:
		return cl(x, p.x0, p.x1) - cx
	case 1:
		return right + cl(y, p.y0, p.y1) - p.y0
	case 2:
		return bottom + p.x1 - cl(x, p.x0, p.x1)
	default:
		return left + p.y1 - cl(y, p.y0, p.y1)
	}
}
