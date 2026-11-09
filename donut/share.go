package donut

// shares returns each section's share of the shape, in order. Values that
// cannot be drawn get 0. A drawn section gets at least minShare; the others
// shrink in proportion to their values to make room.
func shares(values []float64, minShare float64) []float64 {
	out := make([]float64, len(values))
	var drawn []int
	for i, v := range values {
		if drawable(v) {
			drawn = append(drawn, i)
		}
	}
	if len(drawn) == 0 {
		return out
	}
	if float64(len(drawn))*minShare >= 1 {
		for _, i := range drawn {
			out[i] = 1 / float64(len(drawn))
		}
		return out
	}
	pinned := map[int]bool{}
	for {
		room, free := 1-float64(len(pinned))*minShare, 0.0
		for _, i := range drawn {
			if !pinned[i] {
				free += values[i]
			}
		}
		again := false
		for _, i := range drawn {
			if pinned[i] {
				out[i] = minShare
				continue
			}
			out[i] = values[i] / free * room
			if out[i] < minShare {
				pinned[i] = true
				again = true
			}
		}
		if !again {
			return out
		}
	}
}
