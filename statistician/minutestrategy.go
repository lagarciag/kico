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
	"github.com/metakeule/fmtdate"
	"github.com/sirupsen/logrus"
)

type Indicators struct {
	Name      string  `json:"name"`
	Date      string  `json:"date"`
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

	CHigh float64 `json:"c_high"`
	CLow  float64 `json:"c_low"`
	PHigh float64 `json:"p_high"`
	PLow  float64 `json:"p_low"`
	MDM   float64 `json:"mdm"`
	PDM   float64 `json:"pdm"`
	Adx   float64 `json:"adx"`
	MDI   float64 `json:"m_di"`
	PDI   float64 `json:"p_di"`

	TR  float64 `json:"tr"`
	ATR float64 `json:"atr"`

	Md9  float64 `json:"md_9"`
	Buy  bool    `json:"buy"`
	Sell bool    `json:"sell"`
}

type MinuteStrategy struct {
	ID string

	init bool

	addChannel chan float64

	warmAppLock *sync.Cond

	warmUpComplete bool

	indicators Indicators

	kr *kredis.Kredis

	mu *sync.Mutex

	LatestValue float64

	sampleRate int

	multiplier int

	// Moving windows size in minutes
	minuteWindowSize int

	// This is the windows size in number of samples
	movingSampleWindowSize int

	// True when the samples complete the window size
	stable bool

	// Number of samples to be taken to consider the strategy stable.
	// This number is not necesary the movingSampleWindowSize as there could
	// be data that requires more samples for a correct calculation
	stableCount int

	// Holds the current count of samples
	currentSampleCount int

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

	buy  bool
	sell bool

	doDbUpdate bool
}

func NewMinuteStrategy(name string, minuteWindowSize int, stdLimit float64, doLog bool, kr *kredis.Kredis, sampleRate int) *MinuteStrategy {

	ID := fmt.Sprintf("%s_MS_%d", name, minuteWindowSize)

	// -------------------
	// Setup MinutStrategy
	// --------------------
	ps := &MinuteStrategy{}
	ps.ID = ID

	ps.init = true
	ps.indicatorsChan = make(chan Indicators, 1300000)
	ps.indicators = Indicators{}
	ps.doDbUpdate = true
	ps.kr = kr
	//ps.fh = f
	ps.mu = &sync.Mutex{}
	ps.warmAppLock = sync.NewCond(&sync.Mutex{})
	//ps.log = log
	ps.sampleRate = sampleRate
	ps.multiplier = 60 / sampleRate

	ps.minuteWindowSize = minuteWindowSize
	ps.movingSampleWindowSize = minuteWindowSize * ps.multiplier

	ps.movingStats = movingstats.NewMovingStats(int(ps.movingSampleWindowSize))
	ps.addChannel = make(chan float64, ps.movingSampleWindowSize)

	ps.stable = false

	ps.stDevBuyLimit = stdLimit
	ps.stableCount = ps.movingSampleWindowSize * 26

	go ps.indicatorsStorer()

	return ps

}

func (ms *MinuteStrategy) SetDbUpdate(do bool) {
	ms.doDbUpdate = do
}

func (ms *MinuteStrategy) WarmUp(value float64) {
	for n := 0; n < int(ms.stableCount*2); n++ {
		ms.add(value)
	}
	ms.init = false
	ms.warmUpComplete = true
	ms.warmAppLock.Broadcast()
	logrus.Info("Warm up Complete -> ", ms.ID)
}

func (ms *MinuteStrategy) Add(value float64) {

	if ms.init {
		ms.init = false
		go ms.addWorker()
		time.Sleep(time.Millisecond * 500)
		go ms.WarmUp(value)

	} else {
		ms.addChannel <- value
	}

}

func (ms *MinuteStrategy) add(value float64) {

	ms.mu.Lock()
	ms.LatestValue = value
	ms.movingStats.Add(value)

	if ms.currentSampleCount == ms.stableCount {
		ms.stable = true
	}
	ms.currentSampleCount++

	ms.buySellUpdate()

	if ms.warmUpComplete {
		ms.updateIndicators()
	}

	ms.mu.Unlock()

	if ms.warmUpComplete {
		ms.storeIndicators()
	}
}

func (ms *MinuteStrategy) addWorker() {

	ms.warmAppLock.L.Lock()
	ms.warmAppLock.Wait()
	ms.warmAppLock.L.Unlock()

	logrus.Info("addWorker waken up -> ", ms.ID)

	for value := range ms.addChannel {
		ms.add(value)
	}
}

func (ms *MinuteStrategy) StdDevPercentage() float64 {
	stDev := ms.movingStats.StdDevLong()
	stDev100 := stDev * float64(100)

	sma := ms.movingStats.SMA1()
	//logrus.Debugf("STDEV: %f, STDEV100: %f , SMA: %f , PER: %f", stDev, stDev100, sma, stDev100/sma)

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

func (ms *MinuteStrategy) buySellUpdate() {

	ms.buy = false
	ms.sell = false

	adx := ms.movingStats.Adx()
	mDI := ms.movingStats.MinusDI()
	pDI := ms.movingStats.PlusDI()

	pDirectionalBull := false
	mDirectionalBear := false
	adxBull := false
	pDIBull := false
	mDIBear := false
	//diBull := false

	if adx > float64(25) {
		adxBull = true
	}

	if pDI > float64(25) {
		pDIBull = true
	}

	if mDI > float64(25) {
		mDIBear = true
	}

	if pDIBull || adxBull {
		pDirectionalBull = true
	}

	if mDIBear {
		mDirectionalBear = true
	}

	if pDirectionalBull && ms.MacdBullish() && ms.EmaDirectionUp() {
		ms.buy = true
	}

	if mDirectionalBear && !ms.MacdBullish() && !ms.EmaDirectionUp() {
		ms.sell = true
	}

}

func (ms *MinuteStrategy) Buy() bool {
	return ms.buy
}

func (ms *MinuteStrategy) Sell() bool {
	return ms.sell
}

// --------------
// Utilities
// --------------

func (ms *MinuteStrategy) Stable() bool {
	return ms.stable
}

func (ms *MinuteStrategy) updateIndicators() {

	if ms.doDbUpdate {

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
			ms.indicators.StdDev = stDev
		}

		stDevP := ms.StdDevPercentage()

		if math.IsNaN(stDevP) {
			ms.indicators.StdDevPercentage = 0
		} else {
			ms.indicators.StdDevPercentage = stDevP
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
		ms.indicators.Buy = ms.buy
		ms.indicators.Sell = ms.sell

		ms.indicators.CHigh = ms.movingStats.CHigh()
		ms.indicators.PHigh = ms.movingStats.PHigh()
		ms.indicators.CLow = ms.movingStats.CLow()
		ms.indicators.PLow = ms.movingStats.PLow()

		ms.indicators.MDM = ms.movingStats.MinusDM()
		ms.indicators.PDM = ms.movingStats.PlusDM()

		ms.indicators.TR = ms.movingStats.TrueRange()
		ms.indicators.ATR = ms.movingStats.Atr()

		//--------------------
		//Calculate UTC time
		//--------------------

		stringOffset := "+06.00h"

		offSet, err := time.ParseDuration(stringOffset)
		if err != nil {
			panic(err)
		}
		ms.indicators.Date = fmtdate.Format("MM/DD/YYYY hh:mm:ss", time.Now().Add(offSet))
	}
}

func (ms *MinuteStrategy) storeIndicators() {
	if ms.doDbUpdate {
		ms.indicatorsChan <- ms.indicators
	}
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
