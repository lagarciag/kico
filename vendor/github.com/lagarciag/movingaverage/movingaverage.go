package movingaverage

import (
	"math"

	"github.com/lagarciag/ringbuffer"
)

type MovingAverage struct {
	count  int
	period int
	abs    bool

	avgSum      float64
	average     float64
	avgHistBuff *ringbuffer.RingBuffer

	avg2Sum     float64
	variance    float64
	varHistBuff *ringbuffer.RingBuffer

	init bool
}

func New(period int, abs bool) *MovingAverage {

	avg := &MovingAverage{}
	avg.abs = abs
	avg.init = false
	avg.period = period
	avg.avgHistBuff = ringbuffer.NewBuffer(period, false, 0, 0)
	avg.varHistBuff = ringbuffer.NewBuffer(period, false, 0, 0)
	return avg
}

func (avg *MovingAverage) Init(initVal float64, historyValues []float64) {
	avg.init = true

	if avg.abs {
		initVal = math.Abs(initVal)
	}

	avg.average = initVal
	for _, value := range historyValues {
		if avg.abs {
			value = math.Abs(value)
		}
		avg.avgHistBuff.Push(value)
		avg.count++
	}
	if avg.count >= avg.period {
		avg.avgSum = initVal * float64(avg.period)
		//for _, val := range historyValues[0 : avg.period-1] {
		//	avg.avgSum = avg.avgSum + val
		//}
	} else {
		avg.avgSum = initVal * float64(avg.count)
		//for _, val := range historyValues {
		//	avg.avgSum = avg.avgSum + val
		//}
	}

}

func (avg *MovingAverage) Add(value float64) {
	avg.avg(value)

}

func (avg *MovingAverage) SimpleMovingAverage() float64 {
	return avg.average
}

func (avg *MovingAverage) Value() float64 {
	return avg.average
}

func (avg *MovingAverage) MovingStandardDeviation() float64 {
	return math.Sqrt(avg.variance)
}

func (avg *MovingAverage) avg(value float64) {

	if avg.abs {
		value = math.Abs(value)
	}

	avg.count++

	lastAvgValue := avg.avgHistBuff.Oldest()

	//logrus.Info("lastVal :", lastAvgValue)
	//logrus.Info("buf: ", avg.avgHistBuff.GetBuff())

	if avg.abs {
		avg.avgSum = math.Abs((avg.avgSum - lastAvgValue) + value)
	} else {
		avg.avgSum = (avg.avgSum - lastAvgValue) + value
	}

	if avg.count < avg.period {
		avg.average = avg.avgSum / float64(avg.count)
	} else {
		avg.average = avg.avgSum / float64(avg.period)
	}

	avg.avgHistBuff.Push(value)

	value2 := float64(value * value)

	last2AvgValue := avg.varHistBuff.Tail()
	avg.avg2Sum = (avg.avg2Sum - last2AvgValue) + value2

	n := float64(avg.period)
	if avg.count < avg.period {
		n = float64(avg.count)
	}

	avg.variance = math.Abs(((n * avg.avg2Sum) - (avg.avgSum * avg.avgSum)) / (n * (n - 1)))

	if math.IsNaN(avg.variance) {
		avg.variance = float64(0)
	}

	avg.varHistBuff.Push(value2)
}

func (avg *MovingAverage) TestCount() int {
	return avg.count
}
