package multiema

import (
	"github.com/VividCortex/ewma"
	"github.com/sirupsen/logrus"
)

type MultiEma struct {
	init       bool
	count      int
	periods    int
	periodSize int
	//emaSlice   []ewma.MovingAverage
	//intEma     ewma.MovingAverage
	emaSlice []DoubleEma
	intEma   DoubleEma
}

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

func NewMultiEma(periods int, periodSize int, initValues []float64) (mema *MultiEma) {

	mema = &MultiEma{}
	mema.init = false
	if initValues[0] != 0 {
		mema.init = true
	} else {
		logrus.Debug("NewMultiEma initval :", initValues)
	}

	mema.count = 0
	mema.periods = periods
	mema.periodSize = periodSize
	//mema.emaSlice = make([]ewma.MovingAverage, periodSize)
	//mema.intEma = ewma.NewMovingAverage(float64(30))
	mema.emaSlice = make([]DoubleEma, periodSize)
	mema.intEma = NewDema(30, initValues[0])

	for i := range mema.emaSlice {
		//mema.emaSlice[i] = ewma.NewMovingAverage(float64(periods))
		mema.emaSlice[i] = NewDema(periods, initValues[i])
	}
	return mema
}

func (mema *MultiEma) Add(valule float64) {
	if !mema.init {
		mema.emaSlice[mema.count].Set(valule)
	} else {
		mema.emaSlice[mema.count].Add(valule)
	}
	mema.count++
	if mema.count%mema.periodSize == 0 {
		mema.count = 0
		if !mema.init {
			mema.init = true
		}
	}

	val := mema.inVal()
	mema.intEma.Add(val)

}

func (mema *MultiEma) inVal() (val float64) {
	valueCount := mema.count - 1
	if mema.count == 0 {
		valueCount = mema.periodSize - 1
	}
	val = mema.emaSlice[valueCount].Value()
	return val
}

func (mema *MultiEma) Value() (val float64) {
	return mema.intEma.Value()
}
