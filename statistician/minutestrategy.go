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
	log "github.com/sirupsen/logrus"
)

type MinuteStrategy struct {
	ID string

	init bool

	addChannel chan float64

	warmAppLock *sync.Cond

	warmUpComplete bool

	indicators movingstats.Indicators

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

	//log *logrus.Logger
	fh *os.File

	indicatorsChan chan movingstats.Indicators

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
	ps.indicatorsChan = make(chan movingstats.Indicators, 1300000)
	ps.indicators = movingstats.Indicators{}
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

	keyCounter := fmt.Sprintf("%s_INDICATORS", ps.ID)
	indicatorsSaved, err := ps.kr.GetCounterRaw(keyCounter)

	if err != nil {
		log.Fatal("Error reading indicators count")
	}

	oldestIndicator := ps.indicatorsGetter(indicatorsSaved - 1)

	log.Info("Oldest Values: ", oldestIndicator)

	maxPeriodCount := ps.movingSampleWindowSize * 26
	log.Infof("Indicators count: %d, sample window size: %d, max period count (26): %d", indicatorsSaved, ps.movingSampleWindowSize, maxPeriodCount)

	//os.Exit(1)

	// ------------------
	// Get latest values
	// ------------------
	latestIndicators := ps.indicatorsGetter(0)
	previewIndicators := ps.indicatorsGetter(ps.movingSampleWindowSize)

	indicatorsToRetrieve := ps.movingSampleWindowSize * 2

	upperWindowSize := ps.movingSampleWindowSize * 2
	lowerWindowSize := ps.movingSampleWindowSize

	if indicatorsSaved < (ps.movingSampleWindowSize * 2) {

		if indicatorsSaved > ps.movingSampleWindowSize {
			upperWindowSize = 2*ps.movingSampleWindowSize - indicatorsSaved
		} else {
			upperWindowSize = 0
		}
	}

	if indicatorsSaved < ps.movingSampleWindowSize {

		if indicatorsSaved > 0 {
			lowerWindowSize = indicatorsSaved
		} else {
			lowerWindowSize = 0
		}
	}

	indicatorsHistoryTotal := ps.indicatorsHistoryGetter(indicatorsToRetrieve)
	log.Info("History Indicators retrieved: ", len(indicatorsHistoryTotal))
	log.Info("History Upper Window Size :", upperWindowSize)
	log.Info("History Lower Window Size :", lowerWindowSize)

	indicatorsHistory0 := indicatorsHistoryTotal[0:lowerWindowSize]
	indicatorsHistory1 := indicatorsHistoryTotal[ps.movingSampleWindowSize:upperWindowSize]

	log.Info("Updating data from indicators history: ", len(indicatorsHistory0))

	ps.movingStats = movingstats.NewMovingStats(int(ps.movingSampleWindowSize),
		latestIndicators,
		previewIndicators,
		indicatorsHistory0,
		indicatorsHistory1)

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

	//TODO: HACK
	/*
		for n := 0; n < int(ms.stableCount*2); n++ {
			ms.add(value)
		}
	*/
	ms.init = false
	ms.warmUpComplete = true
	ms.warmAppLock.Broadcast()
	log.Info("Warm up Complete -> ", ms.ID)
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

	log.Info("addWorker waken up -> ", ms.ID)

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

	//ms.buy = false
	//ms.sell = false

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

	buyKey := fmt.Sprintf("%s_BUY", ms.ID)
	sellKey := fmt.Sprintf("%s_SELL", ms.ID)

	if ms.doDbUpdate {
		if pDirectionalBull && ms.MacdBullish() && ms.EmaDirectionUp() {
			if ms.buy != true {
				log.Infof("BUY UPDATE for %s :%v", buyKey, true)
				ms.kr.Publish(buyKey, "true")
			}
			ms.buy = true
		} else {
			if ms.buy != false {
				log.Infof("BUY UPDATE for %s :%v", buyKey, false)
				ms.kr.Publish(buyKey, "false")
			}
			ms.buy = false
		}

		if mDirectionalBear && !ms.MacdBullish() && !ms.EmaDirectionUp() {
			if ms.sell != true {
				log.Infof("SELL UPDATE for %s : %v", sellKey, true)
				ms.kr.Publish(sellKey, "true")
			}
			ms.sell = true
		} else {
			if ms.sell != false {
				log.Infof("SELL UPDATE for %s :%v", sellKey, false)
				ms.kr.Publish(sellKey, "false")
			}
			ms.sell = false
		}
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

		ms.indicators.Sema = ms.movingStats.SimpleEma()

		ms.indicators.EmaUp = ms.EmaDirectionUp()

		ms.indicators.Slope = ms.EmaSlope()

		ms.indicators.MacdDiv = ms.movingStats.MacdDiv()

		ms.indicators.Macd12 = ms.movingStats.MacdEma12()
		ms.indicators.Macd26 = ms.movingStats.MacdEma26()

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

func (ms *MinuteStrategy) indicatorsGetter(index int) (indicators movingstats.Indicators) {

	key := fmt.Sprintf("%s_INDICATORS", ms.ID)
	indicatorsJson, err := ms.kr.GetRawString(key, index)

	if err != nil {
		log.Fatal("Fatal error getting indicators: ", err.Error())
	}

	json.Unmarshal([]byte(indicatorsJson), &indicators)

	return indicators

}

func (ms *MinuteStrategy) indicatorsHistoryGetter(size int) (indicators []movingstats.Indicators) {

	key := fmt.Sprintf("%s_INDICATORS", ms.ID)
	indicatorsJson, err := ms.kr.GetRawStringList(key, size)

	if err != nil {
		log.Fatal("Fatal error getting indicators: ", err.Error())
	}

	indicators = make([]movingstats.Indicators, size)

	fmt.Println("SIZE: ", len(indicatorsJson), size)

	for i, indicatorJson := range indicatorsJson {
		anIndicator := movingstats.Indicators{}
		json.Unmarshal([]byte(indicatorJson), &anIndicator)
		indicators[i] = anIndicator
	}

	return indicators

}

func (ms *MinuteStrategy) indicatorsStorer() {
	for indicator := range ms.indicatorsChan {
		//logrus.Info("Store indicator: ", indicator)
		indicatorsJSON, err := json.Marshal(indicator)
		if err != nil {
			log.Error("indicators marshall: ", err.Error())
		}
		//logrus.Infof("STORE: %s , %f", ms.ID, indicator.LastValue)

		err = ms.kr.AddString(ms.ID, "INDICATORS", indicatorsJSON)

		if err != nil {
			log.Fatal("AddString :", err.Error())

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
