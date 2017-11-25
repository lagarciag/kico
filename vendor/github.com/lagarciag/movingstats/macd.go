package movingstats

func (ms *MovingStats) macdCalc(value float64) {

	ms.ema12.Add(value)
	ms.ema26.Add(value)

	ms.macd = ms.ema12.Value() - ms.ema26.Value()
	ms.emaMacd9.Add(ms.macd)

	ms.macdDivergence = ms.macd - ms.emaMacd9.Value()

}
