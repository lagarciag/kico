package statistician

import (
	"fmt"
	"os"

	"sync"

	"time"

	"github.com/lagarciag/movingstats"
	"github.com/sirupsen/logrus"
)

type MinuteStrategy struct {
	mu *sync.Mutex

	latestValue float64

	sampleRate int

	// Moving windows size in minutes
	minuteWindowSize uint

	// This is the windows size in number of samples
	movingSampleWindowSize uint

	// True when the samples complete the window size
	stable bool

	// Number of samples to be taken to consider the strategy stable.
	// This number is not necesary the movingSampleWindowSize as there could
	// be data that requires more samples for a correct calculation
	stableCount uint

	// Holds the current count of samples
	currentSampleCount uint

	// This is the object that holds and does the statistical math
	movingStats *movingstats.MovingStats

	// Standard Deviation limit for making volatility decisions
	stDevBuyLimit float64

	// Simple Exponential moving average Slope
	sEmaSlop float64

	// ------------
	// Buy Signals
	// ------------
	stDevBuy bool
	macdBuy  bool

	//Logging

	log *logrus.Logger
	fh  *os.File
}

func NewMinuteStrategy(minuteWindowSize uint, stdLimit float64, doLog bool) *MinuteStrategy {

	// --------------
	// Setup logging
	// --------------
	logName := fmt.Sprintf("/tmp/mintue_strategy_%d.log", minuteWindowSize)
	csvFile := fmt.Sprintf("/tmp/mintue_strategy_%d.csv", minuteWindowSize)

	if _, err := os.Stat(logName); !os.IsNotExist(err) {
		err := os.Remove(logName)
		if err != nil {
			fmt.Println(err)
			panic("Could not delete file")
		}
	}

	if _, err := os.Stat(csvFile); !os.IsNotExist(err) {
		err := os.Remove(csvFile)
		if err != nil {
			fmt.Println(err)
			panic("Could not delete file")
		}
	}

	f, err := os.Create(csvFile)

	if err != nil {
		panic(err)
	}

	f.WriteString("time,Count,VAL,SMA,STD,STDP,EMA,ELP,N,H,T,EMUP,MAC,Md9,MDI,MBUL,ADX,mDI,pDI,BUY\n")

	log := logrus.New()
	formatter := &logrus.TextFormatter{}
	formatter.FullTimestamp = true
	formatter.ForceColors = true
	log.Level = logrus.DebugLevel
	log.Formatter = formatter

	if doLog {
		filePath := logName
		log.Info("Log file path:", filePath)
		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			log.Out = file
		} else {
			log.Info("Failed to log to file, using default stderr")
		}
	}

	// -------------------
	// Setup MinutStrategy
	// --------------------
	ps := &MinuteStrategy{}
	ps.fh = f
	ps.mu = &sync.Mutex{}
	ps.log = log
	ps.sampleRate = sampleRate

	ps.minuteWindowSize = minuteWindowSize
	ps.movingSampleWindowSize = minuteWindowSize * sampleRate
	ps.movingStats = movingstats.NewMovingStats(int(ps.movingSampleWindowSize))

	ps.stable = false

	ps.stDevBuyLimit = stdLimit
	ps.stableCount = sampleRate * minuteWindowSize * 26
	return ps

}

func (ms *MinuteStrategy) WarmUp(value float64) {
	for n := 0; n < int(ms.stableCount); n++ {
		ms.Add(value)
	}
}

func (ms *MinuteStrategy) Add(value float64) {
	ms.mu.Lock()
	ms.latestValue = value
	ms.movingStats.Add(value)

	if ms.currentSampleCount == ms.stableCount {
		ms.stable = true
	}
	ms.currentSampleCount++

	if ms.currentSampleCount%(30) == 0 {
		ms.log.Info(ms.Print(), ms.minuteWindowSize, ms.currentSampleCount)
		ms.fh.WriteString(ms.PrintCsv() + "\n")
	}

	ms.mu.Unlock()
}

func (ms *MinuteStrategy) StdDevPercentage() float64 {
	stDev := ms.movingStats.StdDevLong()
	sma := ms.movingStats.SMA1()
	return (stDev / sma) * 100
}

func (ms *MinuteStrategy) StdDev() float64 {
	return ms.movingStats.StdDevLong()
}

func (ms *MinuteStrategy) Ema() float64 {
	return ms.movingStats.Ema1()
}

func (ms *MinuteStrategy) EmaSlope() float64 {

	return ms.movingStats.Ema1Slope()
}

func (ms *MinuteStrategy) Madc() float64 {
	return ms.movingStats.Macd()
}

func (ms *MinuteStrategy) MadcDiv() float64 {
	return ms.movingStats.MacdDiv()
}

// --------------
// Buy Signals
// --------------

func (ms *MinuteStrategy) StdDevBuy() bool {
	if ms.StdDevPercentage() >= ms.stDevBuyLimit {
		return true
	} else {
		return false
	}

}

func (ms *MinuteStrategy) MacdBullish() bool {

	macdDiv := ms.movingStats.MacdDiv()

	if macdDiv > 0 {
		return true
	} else {
		return false
	}

}

func (ms *MinuteStrategy) EmaDirectionUp() bool {
	return ms.movingStats.Ema1Up()
}

func (ms *MinuteStrategy) Buy() bool {
	if ms.StdDevBuy() && ms.MacdBullish() && ms.EmaDirectionUp() {
		return true
	}
	return false
}

// --------------
// Utilities
// --------------

func (ms *MinuteStrategy) Stable() bool {
	return ms.stable
}

func (ms *MinuteStrategy) Print() string {
	lastValue := ms.latestValue
	sma := ms.movingStats.SMA1()
	ema := ms.Ema()
	emaUup := ms.EmaDirectionUp()
	slope := ms.EmaSlope()
	macdDiv := ms.movingStats.MacdDiv()
	macdBull := ms.MacdBullish()
	macd := ms.movingStats.Macd()

	stdDev := ms.StdDev()
	stdDevPercentage := ms.StdDevPercentage()
	//stdDevBuy := ms.StdDevBuy()

	md9 := ms.movingStats.EmaMacd9()
	buy := ms.Buy()

	encode := `%-10d - VAL : %-4.2f - SMA: %-4.2f - STD: %-4.2f - STDP: %-4.2f - EMA: %-4.2f - ELP: %+4.2f N: %4.2f - H: %4.2f - T: %4.2F - EMUP: %-4v -MAC: %4.2f -Md9: %4.2f MDI: %-4.2f - MdBUL:%t - BUY:%t- `

	toPrint := fmt.Sprintf(encode,
		ms.currentSampleCount, lastValue,
		sma, stdDev, stdDevPercentage,
		ema, slope, ms.movingStats.HistNow, ms.movingStats.HistMostRecent,
		ms.movingStats.HistOldest, emaUup, macd,
		md9, macdDiv, macdBull, buy)

	return toPrint

}

func (ms *MinuteStrategy) PrintCsv() string {
	lastValue := ms.latestValue
	sma := ms.movingStats.SMA1()
	ema := ms.Ema()
	emaUup := ms.EmaDirectionUp()
	slope := ms.EmaSlope()
	macdDiv := ms.movingStats.MacdDiv()
	macdBull := ms.MacdBullish()
	macd := ms.movingStats.Macd()

	stdDev := ms.StdDev()
	stdDevPercentage := ms.StdDevPercentage()
	//stdDevBuy := ms.StdDevBuy()

	adx := ms.movingStats.Adx()
	mDI := ms.movingStats.MinusDI()
	pDI := ms.movingStats.PlusDI()

	md9 := ms.movingStats.EmaMacd9()
	buy := ms.Buy()
	encode := `%s,%-10d,%-4.2f,%-4.2f,%-4.2f,%-4.2f,%-4.2f,%+4.2f,%4.2f,%4.2f,%4.2f,%-4v,%4.2f,%4.2f,%-4.2f,%t,%4.2f,%4.2f,%4.2f,%t`

	toPrint := fmt.Sprintf(encode, time.Now().String(),
		ms.currentSampleCount, lastValue,
		sma, stdDev, stdDevPercentage,
		ema, slope, ms.movingStats.HistNow, ms.movingStats.HistMostRecent,
		ms.movingStats.HistOldest, emaUup, macd,
		md9, macdDiv, macdBull, adx, mDI, pDI, buy)

	return toPrint

}
