package movingstats

import (
	"math"
)

func (ms *MovingStats) atrCalc(value float64) {
	// --------------
	// Update History
	// --------------
	lastVal := ms.currentWindowHistory.Oldest()
	ms.currentWindowHistory.Push(value)

	// lastWindowHistory is necesary for calculatin True Range
	ms.lastWindowHistory.Push(lastVal)

	// Add data to Average True Range Calculation
	ms.atr.Add(ms.TrueRange())

}

func (ms *MovingStats) CurrentHigh() float64 {
	return ms.currentWindowHistory.High()
}

func (ms *MovingStats) CurrentLow() float64 {
	return ms.currentWindowHistory.Low()
}

func (ms *MovingStats) trueRangeCurrentHighCurrentLow() float64 {
	currentHigh := ms.currentWindowHistory.High()
	currentLow := ms.currentWindowHistory.Low()
	return currentHigh - currentLow
}

func (ms *MovingStats) trueRangeCurrentHighPreviousClose() float64 {
	currentHigh := ms.currentWindowHistory.High()
	previousClose := ms.lastWindowHistory.MostRecent()

	return math.Abs(currentHigh - previousClose)
}

func (ms *MovingStats) PreviousClose() float64 {
	return ms.lastWindowHistory.MostRecent()
}

func (ms *MovingStats) trueRangeCurrentLowPreviousClose() float64 {
	currentLow := ms.currentWindowHistory.Low()
	previousClose := ms.lastWindowHistory.MostRecent()

	return math.Abs(currentLow - previousClose)
}

func (ms *MovingStats) TrueRange() float64 {

	//trSlice := make([]float64, 3)
	//trSlice[0] = ms.trueRangeCurrentHighCurrentLow()
	//trSlice[1] = ms.trueRangeCurrentHighPreviousClose()
	//trSlice[2] = ms.trueRangeCurrentLowPreviousClose()

	//sort.Float64s(trSlice)

	//return trSlice[0]
	return ms.trueRangeCurrentHighCurrentLow()
}
