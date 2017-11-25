package multiema

import (
	"github.com/VividCortex/ewma"
)

type MultiEma struct {
	init       bool
	count      int
	periods    int
	periodSize int
	//emaSlice   []ewma.MovingAverage
	//intEma     ewma.MovingAverage
	emaSlice []dEma
	intEma   dEma
}

type dEma struct {
	emaEma ewma.MovingAverage
	ema    ewma.MovingAverage
}

func NewDema(periodSize int) dEma {
	dEma := dEma{}
	dEma.ema = ewma.NewMovingAverage(float64(periodSize))
	dEma.emaEma = ewma.NewMovingAverage(float64(periodSize))
	return dEma
}

func (dema *dEma) Add(value float64) {
	dema.ema.Add(value)
	dema.emaEma.Add(dema.ema.Value())
}

func (dema *dEma) Value() float64 {
	val := 2*dema.ema.Value() - dema.emaEma.Value()
	return val
}

func (dema *dEma) Set(value float64) {
	dema.ema.Set(value)
	dema.emaEma.Set(value)
}

func NewMultiEma(periods int, periodSize int) (mema *MultiEma) {

	mema = &MultiEma{}
	mema.init = false
	mema.count = 0
	mema.periods = periods
	mema.periodSize = periodSize
	//mema.emaSlice = make([]ewma.MovingAverage, periodSize)
	//mema.intEma = ewma.NewMovingAverage(float64(30))
	mema.emaSlice = make([]dEma, periodSize)
	mema.intEma = NewDema(30)

	for i := range mema.emaSlice {
		//mema.emaSlice[i] = ewma.NewMovingAverage(float64(periods))
		mema.emaSlice[i] = NewDema(periods)
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
