package ringbuffer

import "github.com/sirupsen/logrus"

type RingBuffer struct {
	buff          []float64
	head          int
	tail          int
	high          int
	low           int
	size          int
	recordHighLow bool
	init          bool
	counter       int
	initHighSet   bool
	initLowSet    bool
	initHighValue float64
	initLowValue  float64
}

func NewBuffer(size int, recordHighLow bool, initHigh, initLow float64) *RingBuffer {

	rb := &RingBuffer{}

	if initHigh != 0 {
		rb.initHighSet = true
		rb.initHighValue = initHigh
	}

	if initLow != 0 {
		rb.initLowSet = true
		rb.initLowValue = initLow
	}

	rb.recordHighLow = recordHighLow
	rb.size = size
	rb.buff = make([]float64, rb.size)

	rb.head = -1
	rb.tail = rb.head + 1
	rb.high = 0
	rb.low = 0
	rb.counter = 0

	rb.buff[len(rb.buff)-1] = initHigh
	rb.high = len(rb.buff) - 1

	rb.buff[len(rb.buff)-2] = initLow
	rb.low = len(rb.buff) - 2

	return rb
}

func (rb *RingBuffer) PushBuffer(values []float64) {
	for _, value := range values {
		rb.Push(value)
	}
}

func (rb *RingBuffer) SetInitHigh(highValue float64) {
	if highValue != 0 {
		logrus.Info("RingBuffer Init High: ", highValue)
		rb.initHighValue = highValue
		rb.initHighSet = true
		rb.counter = 0
	} else {
		logrus.Info("Skip ringbuffer init high")
		rb.initHighSet = false
	}
}

func (rb *RingBuffer) SetInitLow(value float64) {
	if value != 0 {
		logrus.Info("RingBuffer Init Low: ", value)
		rb.initLowValue = value
		rb.initLowSet = true
		rb.counter = 0
	} else {
		rb.initLowSet = false
		logrus.Info("Skip ringbuffer init low")
	}
}

//Push adds a new element to the buffer
func (rb *RingBuffer) Push(value float64) {

	if rb.counter == 0 && !rb.initLowSet {
		rb.SetInitLow(value)
		for i := range rb.buff {
			rb.buff[i] = value
		}
	}

	if rb.counter == 0 && !rb.initHighSet {
		rb.SetInitHigh(value)
	}

	if value > rb.initHighValue || rb.counter >= rb.size {
		rb.initHighSet = false
	}

	if value < rb.initLowValue || rb.counter >= rb.size {
		rb.initLowSet = false
	}

	highAtTail := false
	lowAtTail := false

	//if (rb.high == rb.tail) || (rb.high == rb.head) {
	if rb.high == rb.tail {
		highAtTail = true
	}

	//if (rb.low == rb.tail) || (rb.low == rb.head) {
	if rb.low == rb.tail {
		lowAtTail = true
	}

	// -----------------------
	// Increase ring pointers
	// -----------------------
	rb.head++
	rb.tail++

	if rb.tail%(rb.size) == 0 {
		rb.tail = 0
	}

	if rb.head%(rb.size) == 0 {
		rb.head = 0
	}

	// ----------------------
	// Put new value in head
	// ----------------------
	rb.buff[rb.head] = value

	if rb.recordHighLow == true {
		// --------------------
		// rb.high end of life,
		// --------------------
		if highAtTail {
			hVal := float64(0)

			for i, val := range rb.buff {
				if val > hVal {
					hVal = val
					rb.high = i
				}

			}
		}

		// --------------------
		// rb.low end of life
		// --------------------
		if lowAtTail {
			lVal := float64(0Xffffffff)
			//rb.buff[rb.tail] = lVal
			for i, val := range rb.buff {
				if val < lVal {
					lVal = val
					rb.low = i
				}

			}
		}

		if value >= rb.buff[rb.high] {
			rb.high = rb.head
		}

		if value <= rb.buff[rb.low] {
			rb.low = rb.head
		}
	}

	if !rb.init {
		rb.init = true
	}
	rb.counter++

}

//Tail returns the element at the buffer tail
func (rb *RingBuffer) Tail() float64 {
	return rb.buff[rb.tail]
}

//Head returns the element at the buffer tail
func (rb *RingBuffer) Head() float64 {
	return rb.buff[rb.head]
}

//MostRecent returns the element at the head - 1
func (rb *RingBuffer) MostRecent() float64 {
	return rb.buff[rb.head]

}

//Oldest returns the element at the head - 1
func (rb *RingBuffer) Oldest() float64 {
	oldest := rb.buff[rb.tail]
	if rb.counter < rb.size {
		oldest = rb.buff[0]
	}
	return oldest
}

//Tail returns the element at the buffer tail
func (rb *RingBuffer) High() float64 {

	value := rb.buff[rb.high]

	if rb.initHighSet {
		value = rb.initHighValue
	}

	return value
}

//Head returns the element at the buffer tail
func (rb *RingBuffer) Low() float64 {

	value := rb.buff[rb.low]

	if rb.initLowSet {
		value = rb.initLowValue
	}

	return value
}

func (rb *RingBuffer) tailNext() int {
	tailNext := rb.tail + 1

	if tailNext%(rb.size+1) == 0 {
		return 0
	}

	return tailNext

}

func (rb *RingBuffer) GetBuff() []float64 {

	return rb.buff
}
