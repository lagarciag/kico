package movingstats

import (
	"math"

	"fmt"

	"github.com/lagarciag/ringbuffer"
)

type MovingAverage struct {
	count  int
	period int

	avgSum  float64
	average float64
	//avgHistBuff *historyBuffer
	avgHistBuff *ringbuffer.RingBuffer

	avg2Sum       float64
	variance      float64
	varHistBuff   *ringbuffer.RingBuffer
	lastAverage   float64
	last2AvgValue float64
}

func NewAverage(period int) *MovingAverage {

	avg := &MovingAverage{}
	avg.period = period
	avg.avgHistBuff = ringbuffer.NewBuffer(period, false)
	avg.varHistBuff = ringbuffer.NewBuffer(period, false)
	return avg
}

func (avg *MovingAverage) Add(value float64) {
	avg.avg(value)
}

func (avg *MovingAverage) SimpleMovingAverage() float64 {
	return avg.average
}

func (avg *MovingAverage) MovingStandardDeviation() float64 {
	return math.Sqrt(avg.variance)
}

func (avg *MovingAverage) avg(value float64) {
	avg.count++

	lastAvgValue := avg.avgHistBuff.Oldest()

	fmt.Println(avg.avgHistBuff, lastAvgValue)

	avg.avgSum = (avg.avgSum - lastAvgValue) + value

	if avg.count < avg.period {
		avg.average = avg.avgSum / float64(avg.count)
	} else {
		avg.average = avg.avgSum / float64(avg.period)
	}

	avg.avgHistBuff.Push(value)

	value2 := value * value
	last2AvgValue := avg.varHistBuff.Oldest()
	avg.avg2Sum = (avg.avg2Sum - last2AvgValue) + value2

	n := float64(avg.period)
	if avg.count < avg.period {
		n = float64(avg.count)
	}

	avg.variance = ((n * avg.avg2Sum) - (avg.avgSum * avg.avgSum)) / (n * (n - 1))

	avg.varHistBuff.Push(value2)
}
