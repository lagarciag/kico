package movingaverage

import (
	"fmt"
	"math"

	"github.com/lagarciag/ringbuffer"
	log "github.com/sirupsen/logrus"
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

	//init bool
}

func New(period int, abs bool) *MovingAverage {

	avg := &MovingAverage{}
	avg.count = 0
	avg.abs = abs
	//avg.init = false
	avg.period = period
	avg.avgHistBuff = ringbuffer.NewBuffer(period, false, 0, 0)
	avg.varHistBuff = ringbuffer.NewBuffer(period, false, 0, 0)
	return avg
}

func (avg *MovingAverage) Init(initVal float64, historyValues []float64) {
	//avg.init = true
	avg.avgSum = 0
	var realhistory []float64
	if len(historyValues) > 1 {

		avg.average = initVal
		avg.avgSum = 0

		log.Info("MovingAverage Init: Avg Init val: ", avg.average)

		if len(historyValues) > avg.period {
			realhistory = historyValues[:avg.period]
		} else {
			realhistory = historyValues
		}

		if len(realhistory) > avg.period {
			log.Error("realhistory still > avgperiod")
		}

		for _, value := range realhistory {
			avg.avgHistBuff.Push(value)
			avg.count++
			avg.avgSum = avg.avgSum + value
		}

		if avg.count <= avg.period {
			dbgMsg := `
			avg.period : %d
			avg.avgSum : %f
			avg.count  : %d
		`
			log.Infof("History is <= than period: %s", fmt.Sprintf(dbgMsg, avg.period, avg.avgSum, avg.count))

			avg.average = avg.avgSum / float64(avg.count)

		}
		if avg.count > avg.period {
			//avg.avgSum = initVal * float64(avg.period)
			avg.average = avg.avgSum / float64(avg.period)
			dbgMsg := `
			avg.period : %d
			avg.avgSum : %f
			avg.count  : %d
		`
			log.Infof("History is >>> than period: %s", fmt.Sprintf(dbgMsg, avg.period, avg.avgSum, avg.count))

		} else {
			//avg.avgSum = initVal * float64(avg.count)
		}

	}

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

func (avg *MovingAverage) Add(value float64) {
	avg.count++
	if avg.abs {
		value = math.Abs(value)
	}

	lastAvgValue := avg.avgHistBuff.Oldest()

	if avg.count <= avg.period {
		avg.avgSum = avg.avgSum + value
		avg.average = avg.avgSum / float64(avg.count)
	} else {
		avg.avgSum = avg.avgSum - lastAvgValue + value
		avg.average = avg.avgSum / float64(avg.period)
	}

	//log.Info("lastVal :", lastAvgValue)
	//log.Info("buf: ", avg.avgHistBuff.GetBuff())
	//log.Info("avg:", avg.average)

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
