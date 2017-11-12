package movingstats_test

import (
	"testing"

	"fmt"

	"github.com/lagarciag/movingstats"
)

func TestTrueRangeSimple(t *testing.T) {

	ms := movingstats.NewMovingStats(10)

	floatSlice := []float64{10, 11, 12, 13, 14, 15, 16, 17, 16, 11,
		10, 11, 12, 13, 14, 15, 16, 17, 16, 11,
		10, 11, 12, 13, 14, 15, 16, 17, 16, 11,
		9, 11, 12, 13, 14, 15, 16, 17, 16, 11,
		10, 12, 13, 14, 13, 12, 11, 10, 9, 10}

	// WarmUp

	for i := 0; i < 39; i++ {
		ms.Add(floatSlice[i])
	}

	//ms.Add(floatSlice[10])

	currentHigh := ms.CurrentHigh()
	currentLow := ms.CurrentLow()

	trueRange := ms.TrueRange()

	dmi := ms.Adx()

	fmt.Println("adx ", dmi)

	t.Log("Current high: ", currentHigh)
	t.Log("Current Low: ", currentLow)
	t.Log("Current TR: ", trueRange)

}
