package movingstats_test

import (
	"testing"

	"fmt"

	"github.com/lagarciag/movingstats"
)

func TestTrueRangeSimple(t *testing.T) {
	ind1 := movingstats.Indicators{}
	ind2 := movingstats.Indicators{}
	arg1 := make([]movingstats.Indicators, 1)
	arg2 := make([]movingstats.Indicators, 1)
	ms := movingstats.NewMovingStats(10, ind1, ind2, arg1, arg2, arg2, false, "test")

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
