package main

import (
	"fmt"
	"math"

	"github.com/ingvarch/pulse"
	"github.com/ingvarch/pulse/theme"
)

func main() {
	preset := theme.TokyoNight()
	chartW, chartH := 80, 13

	lc := pulse.New(chartW, chartH,
		pulse.WithRange(0, 100),
		pulse.WithTicks(0, 25, 50, 75, 100),
		pulse.WithLineWidth(1),
		pulse.WithAxisStyle(preset.Axis),
		pulse.WithLineStyle(preset.LineFor(60)),
	)

	for i := 0; i < chartW; i++ {
		v := 50 + 38*math.Sin(float64(i)*0.16)
		lc.Push(v)
	}

	fmt.Println(lc.View())
}
