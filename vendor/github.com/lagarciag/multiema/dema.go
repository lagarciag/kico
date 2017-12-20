package multiema

import "github.com/VividCortex/ewma"

type DoubleEma struct {
	emaEma ewma.MovingAverage
	ema    ewma.MovingAverage
}

func NewDema(periodSize int, initVal float64) DoubleEma {
	dEma := DoubleEma{}
	dEma.ema = ewma.NewMovingAverage(float64(periodSize))
	dEma.emaEma = ewma.NewMovingAverage(float64(periodSize))
	dEma.Set(initVal)
	dEma.emaEma.Set(initVal)
	return dEma
}

func (dema *DoubleEma) Add(value float64) {
	dema.ema.Add(value)
	dema.emaEma.Add(dema.ema.Value())
}

func (dema *DoubleEma) Value() float64 {
	val := 2*dema.ema.Value() - dema.emaEma.Value()
	return val
}

func (dema *DoubleEma) Set(value float64) {
	dema.ema.Set(value)
	dema.emaEma.Set(value)
}
