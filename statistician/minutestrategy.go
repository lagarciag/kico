package statistician

import (
	"fmt"
	"os"

	"sync"

	"time"

	"encoding/json"

	"math"

	"github.com/lagarciag/movingstats"
	"github.com/lagarciag/tayni/kredis"
	"github.com/sirupsen/logrus"
)

type Indicators struct {
	name      string  `json:"name"`
	LastValue float64 `json:"last_value"`
	Sma       float64 `json:"sma"`
	Ema       float64 `json:"ema"`
	EmaUp     bool    `json:"ema_up"`
	Slope     float64 `json:"slope"`
	MacdDiv   float64 `json:"macd_div"`
	MacdBull  bool    `json:"macd_bull"`
	Macd      float64 `json:"macd"`

	StdDev           float64 `json:"std_dev"`
	StdDevPercentage float64 `json:"std_dev_percentage"`
	//stdDevBuy := ms.StdDevBuy()

	Adx float64 `json:"adx"`
	MDI float64 `json:"m_di"`
	PDI float64 `json:"p_di"`

	Md9 float64 `json:"md_9"`
	Buy bool    `json:"buy"`
}

type MinuteStrategy struct {
	ID string

	indicators Indicators

	kr *kredis.Kredis

	mu *sync.Mutex

	LatestValue float64

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

	indicatorsChan chan Indicators
}

func NewMinuteStrategy(name string, minuteWindowSize uint, stdLimit float64, doLog bool, kr *kredis.Kredis) *MinuteStrategy {

	ID := fmt.Sprintf("%s_MS_%d", name, minuteWindowSize)

	// --------------
	// Setup logging
	// --------------
	logName := fmt.Sprintf("/tmp/%s.log", ID)
	csvFile := fmt.Sprintf("/tmp/%s.csv", ID)

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
	ps.ID = ID
	ps.indicatorsChan = make(chan Indicators, 1300000)
	ps.indicators = Indicators{}
	ps.kr = kr
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

	go ps.indicatorsStorer()

	return ps

}

func (ms *MinuteStrategy) WarmUp(value float64) {
	for n := 0; n < int(ms.stableCount); n++ {
		ms.Add(value)
	}
}

func (ms *MinuteStrategy) Add(value float64) {
	ms.mu.Lock()
	ms.LatestValue = value
	ms.movingStats.Add(value)

	if ms.currentSampleCount == ms.stableCount {
		ms.stable = true
	}
	ms.currentSampleCount++

	ms.updateIndicators()
	ms.mu.Unlock()
	ms.storeIndicators()

}

func (ms *MinuteStrategy) StdDevPercentage() float64 {
	stDev := ms.movingStats.StdDevLong()
	stDev100 := stDev * float64(100)

	sma := ms.movingStats.SMA1()
	//logrus.Debugf("STDEV * 100: %f , SMA: %f , PER: %f", stDev100, sma, stDev100/sma)

	return stDev100 / sma
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

func (ms *MinuteStrategy) updateIndicators() {
	ms.indicators.LastValue = ms.LatestValue

	ms.indicators.Sma = ms.movingStats.SMA1()

	ms.indicators.Ema = ms.Ema()

	ms.indicators.EmaUp = ms.EmaDirectionUp()

	ms.indicators.Slope = ms.EmaSlope()

	ms.indicators.MacdDiv = ms.movingStats.MacdDiv()

	ms.indicators.MacdBull = ms.MacdBullish()
	ms.indicators.Macd = ms.movingStats.Macd()

	stDev := ms.StdDev()
	if math.IsNaN(stDev) {
		ms.indicators.StdDev = 0
	} else {
		ms.indicators.StdDev = ms.StdDev()
	}

	stDevP := ms.StdDevPercentage()
	if math.IsNaN(stDevP) {
		ms.indicators.StdDevPercentage = 0
	} else {
		ms.indicators.StdDev = ms.StdDevPercentage()
	}

	//stdDevBuy := ms.StdDevBuy()
	adx := ms.movingStats.Adx()

	if math.IsNaN(adx) {
		ms.indicators.Adx = 0
	} else {
		ms.indicators.Adx = adx
	}

	MDI := ms.movingStats.MinusDI()

	if math.IsNaN(MDI) {
		ms.indicators.MDI = 0
	} else {
		ms.indicators.MDI = MDI
	}

	PDI := ms.movingStats.PlusDI()

	if math.IsNaN(PDI) {
		ms.indicators.PDI = 0
	} else {
		ms.indicators.PDI = PDI
	}

	ms.indicators.Md9 = ms.movingStats.EmaMacd9()
	ms.indicators.Buy = ms.Buy()

}

/*
func (ms *MinuteStrategy) Print() string {

	encode := `%-10d - VAL : %-4.2f - SMA: %-4.2f - STD: %-4.2f - STDP: %-4.2f - EMA: %-4.2f - ELP: %+4.2f N: %4.2f - H: %4.2f - T: %4.2F - EMUP: %-4v -MAC: %4.2f -Md9: %4.2f MDI: %-4.2f - MdBUL:%t - BUY:%t- `

	toPrint := fmt.Sprintf(encode,
		ms.currentSampleCount, ms.indicators.LastValue,
		ms.indicators.Sma, ms.indicators.StdDev, ms.indicators.StdDevPercentage,
		ms.indicators.Ema, ms.indicators.Slope, ms.movingStats.HistNow, ms.movingStats.HistMostRecent,
		ms.movingStats.HistOldest, emaUup, macd,
		md9, macdDiv, macdBull, buy)

	return toPrint

}
*/

func (ms *MinuteStrategy) storeIndicators() {

	ms.indicatorsChan <- ms.indicators

}

func (ms *MinuteStrategy) indicatorsStorer() {
	for indicator := range ms.indicatorsChan {
		//logrus.Info("Store indicator: ", indicator)
		indicatorsJSON, err := json.Marshal(indicator)
		if err != nil {
			logrus.Error("indicators marshall: ", err.Error())
		}
		//logrus.Infof("STORE: %s , %f", ms.ID, indicator.LastValue)

		err = ms.kr.AddString(ms.ID, "INDICATORS", indicatorsJSON)

		if err != nil {
			logrus.Fatal("AddString :", err.Error())

		}

		//logrus.Infof("STORE DONE: %s , %f", ms.ID, indicator.LastValue)
	}

}

func (ms *MinuteStrategy) PrintCsv() string {
	lastValue := ms.LatestValue
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
