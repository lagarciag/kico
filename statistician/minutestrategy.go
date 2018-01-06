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

	count uint64

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

	// --------------------
	// Indicators History
	// -------------------
	indicatorsToRetrieve   int
	indicatorsHistoryTotal []movingstats.Indicators
	indicatorsSaved        int
	latestIndicators       movingstats.Indicators
	previewIndicators      movingstats.Indicators
	indicatorsHistory0     []movingstats.Indicators
	indicatorsHistory1     []movingstats.Indicators
	dirtyHistory           bool
}

func NewMinuteStrategy(name string, minuteWindowSize int, stdLimit float64, doLog bool, kr *kredis.Kredis, sampleRate int) *MinuteStrategy {

	ID := fmt.Sprintf("%s_MS_%d", name, minuteWindowSize)

	// -------------------
	// Setup MinutStrategy
	// --------------------
	ps := &MinuteStrategy{}
	ps.dirtyHistory = false
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

	// ---------------------------
	// Get Indicators to produce
	// indicators history
	// ---------------------------
	ps.initIndicators()

	ps.movingStats = movingstats.NewMovingStats(int(ps.movingSampleWindowSize),
		ps.latestIndicators,
		ps.previewIndicators,
		ps.indicatorsHistory0,
		ps.indicatorsHistory1,
		ps.indicatorsHistoryTotal,
		ps.dirtyHistory,
		ID)

	ps.addChannel = make(chan float64, ps.movingSampleWindowSize)

	ps.stable = false

	ps.stDevBuyLimit = stdLimit
	ps.stableCount = ps.movingSampleWindowSize * 26

	go ps.indicatorsStorer()

	return ps

}

func (ps *MinuteStrategy) initIndicators() {

	ps.indicatorsToRetrieve = -1
	ps.indicatorsHistoryTotal = ps.indicatorsHistoryGetter(ps.indicatorsToRetrieve)
	ps.indicatorsSaved = len(ps.indicatorsHistoryTotal)

	if ps.indicatorsSaved < ps.movingSampleWindowSize {
		ps.dirtyHistory = true
		log.Warn("Indicators History is to shallow, setting dirty bit")
	}

	log.Infof("Indicators count: %d, sample window size: %d, max period count (26): %d", ps.indicatorsSaved, ps.movingSampleWindowSize)

	// ------------------
	// Get latest values
	// ------------------
	ps.latestIndicators = ps.indicatorsGetter(0)
	ps.previewIndicators = ps.indicatorsGetter(ps.movingSampleWindowSize)

	upperWindowSize := ps.movingSampleWindowSize * 2
	lowerWindowSize := ps.movingSampleWindowSize

	if ps.indicatorsSaved < (ps.movingSampleWindowSize * 2) {

		if ps.indicatorsSaved > ps.movingSampleWindowSize {
			upperWindowSize = 2*ps.movingSampleWindowSize - ps.indicatorsSaved
		} else {

			upperWindowSize = 0
		}
	}

	if ps.indicatorsSaved < ps.movingSampleWindowSize {
		if ps.indicatorsSaved > 0 {
			lowerWindowSize = ps.indicatorsSaved - 1
		} else {
			ps.dirtyHistory = true
			lowerWindowSize = 1
		}
	}

	log.Info("History Indicators retrieved: ", len(ps.indicatorsHistoryTotal))
	log.Info("History Upper Window Size :", upperWindowSize)
	log.Info("History Lower Window Size :", lowerWindowSize)

	ps.indicatorsHistory0 = ps.indicatorsHistoryTotal[0:lowerWindowSize]

	if upperWindowSize > ps.movingSampleWindowSize {

		ps.indicatorsHistory1 = ps.indicatorsHistoryTotal[ps.movingSampleWindowSize:upperWindowSize]

	} else {
		ps.indicatorsHistory1 = ps.indicatorsHistory0
	}

	log.Info("Updating data from indicators history: ", len(ps.indicatorsHistory0))

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
	ms.count++
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

	sma := ms.movingStats.SmaShort()
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

	if adx > float64(20) {
		adxBull = true
	}

	if pDI > float64(15) && pDI > mDI {
		pDIBull = true
	}

	if mDI > float64(20) {
		mDIBear = true
	}

	if pDIBull && adxBull {
		pDirectionalBull = true
	}

	if mDIBear || adxBull {
		mDirectionalBear = true
	}

	buyKey := fmt.Sprintf("%s_BUY", ms.ID)
	sellKey := fmt.Sprintf("%s_SELL", ms.ID)

	atrLimitOk := false

	if ms.movingStats.Atrp() > ms.movingStats.AtrLimit() {
		atrLimitOk = true
		//log.Debug("AtrLimit OK ", ms.movingStats.Atrp())
	} else {
		//log.Debug("AtrLimit NOT ok, ", ms.movingStats.Atrp())
	}

	if ms.doDbUpdate {
		if pDirectionalBull && ms.MacdBullish() && ms.EmaDirectionUp() && atrLimitOk {
			if ms.buy == false {
				log.Infof("BUY CHANGE for %s :%v", buyKey, true)
			}
			if err := ms.kr.Publish(buyKey, "true"); err != nil {
				log.Errorf("Publishing to: %s -> %s ", buyKey, "true")
			}
			ms.buy = true
		} else {
			if ms.buy == true {
				log.Infof("BUY CHANGE for %s :%v", buyKey, false)
			}
			if err := ms.kr.Publish(buyKey, "false"); err != nil {
				log.Errorf("Publishing to: %s -> %s ", buyKey, "false")
			}
			ms.buy = false
		}

		if mDirectionalBear && !ms.MacdBullish() && !ms.EmaDirectionUp() {
			if ms.sell == false {
				log.Infof("SELL CHANGE for %s : %v", sellKey, true)
			}
			if err := ms.kr.Publish(sellKey, "true"); err != nil {
				log.Errorf("Publishing to: %s -> %s ", sellKey, "true")
			}
			ms.sell = true
		} else {
			if ms.sell == true {
				log.Infof("SELL CHANGE for %s :%v", sellKey, false)
			}
			if err := ms.kr.Publish(sellKey, "false"); err != nil {
				log.Errorf("Publishing to: %s -> %s ", buyKey, "false")
			}
			ms.sell = false
		}

		if ms.count%6 == 0 {
			log.Infof("** BUY STATUS UPDATE for %s :%v", buyKey, ms.buy)
			log.Infof("** SEL STATUS UPDATE for %s :%v", sellKey, ms.sell)
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

		ms.indicators.Sma = ms.movingStats.SmaShort()

		ms.indicators.SmaLong = ms.movingStats.SmaLong()

		ms.indicators.Ema = ms.Ema()

		ms.indicators.Mema9 = ms.movingStats.Mema9()

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
		ms.indicators.ATRP = ms.movingStats.Atrp()

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

	if err = json.Unmarshal([]byte(indicatorsJson), &indicators); err != nil {
		log.Error("unmarshaling indicators json")
	}

	return indicators

}

func (ms *MinuteStrategy) indicatorsHistoryGetter(size int) (indicators []movingstats.Indicators) {

	key := fmt.Sprintf("%s_INDICATORS", ms.ID)
	indicatorsJson, err := ms.kr.GetRawStringList(key, size)

	if err != nil {
		log.Fatal("Fatal error getting indicators: ", err.Error())
	}

	if len(indicatorsJson) < size {
		size = len(indicatorsJson)
	}

	if size == -1 {
		size = len(indicatorsJson)
	}

	if size == 0 {
		size = 1
	}

	indicators = make([]movingstats.Indicators, size)

	for i, indicatorJson := range indicatorsJson {
		anIndicator := movingstats.Indicators{}
		if err := json.Unmarshal([]byte(indicatorJson), &anIndicator); err != nil {
			log.Error("unmarshaling indicators json")
		}

		indicators[i] = anIndicator
	}

	if len(indicatorsJson) == 0 {
		indicators[0].LastValue = 0
	}

	return indicators

}

func (ms *MinuteStrategy) indicatorsStorer() {
	for indicator := range ms.indicatorsChan {
		//log.Info("Store indicator: ", indicator, ms.ID, ms.)
		indicatorsJSON, err := json.Marshal(indicator)
		if err != nil {
			log.Error("indicators marshall: ", err.Error())
		}
		//log.Infof("STORE: %s , %f", ms.ID, indicator.LastValue)

		err = ms.kr.AddString(ms.ID, "INDICATORS", indicatorsJSON)

		if err != nil {
			log.Fatal("AddString :", err.Error())

		}

		//logrus.Infof("STORE DONE: %s , %f", ms.ID, indicator.LastValue)
	}

}
