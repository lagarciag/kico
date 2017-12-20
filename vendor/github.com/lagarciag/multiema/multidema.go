package multiema

import "github.com/sirupsen/logrus"

type MultiDEma struct {
	initCount  int
	init       bool
	count      int
	periods    int
	periodSize int
	//emaSlice   []ewma.MovingAverage
	//intEma     ewma.MovingAverage
	emaSlice []DoubleEma
	intEma   DoubleEma
}

func NewMultiDEma(periods int, periodSize int, initValue float64) (mema *MultiDEma) {

	mema = &MultiDEma{}
	mema.init = false
	if initValue != float64(0) {
		mema.init = true
	} else {
		logrus.Debug("NewMultiDEma initval :", initValue)
	}

	mema.count = 0
	mema.periods = periods
	mema.periodSize = periodSize
	//mema.emaSlice = make([]ewma.MovingAverage, periodSize)
	//mema.intEma = ewma.NewMovingAverage(float64(30))
	mema.emaSlice = make([]DoubleEma, periodSize)
	mema.intEma = NewDema(30, initValue)

	//logrus.Info("NewEma Init: ", len(initValues), len(mema.emaSlice))

	for i := range mema.emaSlice {
		mema.emaSlice[i] = NewDema(periods, initValue)
	}
	return mema
}

func (mema *MultiDEma) Add(valule float64) {
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
	mema.initCount++
}

func (mema *MultiDEma) inVal() (val float64) {
	valueCount := mema.count - 1
	if mema.count == 0 {
		valueCount = mema.periodSize - 1
	}
	val = mema.emaSlice[valueCount].Value()
	return val
}

func (mema *MultiDEma) Value() (val float64) {
	return mema.intEma.Value()
}
