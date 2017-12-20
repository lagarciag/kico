package multiema

import "github.com/VividCortex/ewma"

type SimpleEma struct {
	ema ewma.MovingAverage
}

func NewSema(periodSize int, initVal float64) SimpleEma {
	dEma := SimpleEma{}
	dEma.ema = ewma.NewMovingAverage(float64(periodSize))
	dEma.Set(initVal)
	return dEma
}

func (sema *SimpleEma) Add(value float64) {
	sema.ema.Add(value)
}

func (sema *SimpleEma) Value() float64 {
	val := sema.ema.Value()
	return val
}

func (sema *SimpleEma) Set(value float64) {
	sema.ema.Set(value)
}
