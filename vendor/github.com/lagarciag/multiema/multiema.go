package multiema

import (
	"github.com/VividCortex/ewma"
)

type MultiEma struct {
	init       bool
	count      int
	periods    int
	periodSize int
	emaSlice   []ewma.MovingAverage
}

func NewMultiEma(periods int, periodSize int) (mema *MultiEma) {

	mema = &MultiEma{}
	mema.init = false
	mema.count = 0
	mema.periods = periods
	mema.periodSize = periodSize
	mema.emaSlice = make([]ewma.MovingAverage, periodSize)

	for i := range mema.emaSlice {
		mema.emaSlice[i] = ewma.NewMovingAverage(float64(periods))
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
}

func (mema *MultiEma) Value() (val float64) {
	valueCount := mema.count - 1
	if mema.count == 0 {
		valueCount = mema.periodSize - 1
	}
	val = mema.emaSlice[valueCount].Value()
	return val
}
