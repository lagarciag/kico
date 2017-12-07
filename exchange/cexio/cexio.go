package cexio

import (
	"github.com/VividCortex/ewma"
	cexioapi "github.com/lagarciag/cexioapi"
	log "github.com/sirupsen/logrus"

	"time"

	"sync"

	"fmt"

	"strconv"

	"github.com/coreos/go-systemd/daemon"
	"github.com/lagarciag/tayni/kredis"
	"github.com/lagarciag/tayni/statistician"
)

type CollectorConfig struct {
	CexioKey     string
	CexioSecret  string
	Pairs        []string
	SampleRate   int
	HistoryCount int
}

type Bot struct {
	name         string
	key          string
	secret       string
	pairs        []string
	sampleRate   int
	historyCount int
	api          *cexioapi.API

	apiLock *sync.Mutex
	kr      *kredis.Kredis

	ticksPerMinute int
	btcUsdBase     float64

	tickerSub chan cexioapi.ResponseTickerSubData

	stats map[string]*statistician.Statistician

	// --------------------
	// Control Structures
	// --------------------
	priceUpdateTimer *time.Ticker
	statsUpdateTimer *time.Ticker
	shutdownCond     *sync.Cond
	priceUpdaterCond *sync.Cond
	apiError         chan error
	fullStop         chan bool
	apiStop          chan bool
	priceAdderChan   map[string]chan float64
	apiOnline        bool
}

func NewBot(config CollectorConfig, kr *kredis.Kredis) (bot *Bot) {

	//--------------------------------
	//Move this to a secure location
	//--------------------------------
	bot = &Bot{}
	bot.historyCount = config.HistoryCount
	bot.name = "CEXIO"
	bot.key = config.CexioKey
	bot.pairs = config.Pairs
	bot.sampleRate = config.SampleRate
	bot.secret = config.CexioSecret
	bot.kr = kr
	bot.kr.Start()
	bot.api, bot.apiError = cexioapi.NewPublicAPI()

	bot.tickerSub = make(chan cexioapi.ResponseTickerSubData)

	bot.shutdownCond = sync.NewCond(&sync.Mutex{})
	bot.priceUpdaterCond = sync.NewCond(&sync.Mutex{})
	bot.apiStop = make(chan bool)
	bot.fullStop = make(chan bool)
	//bot.priceAdderChan = make(chan float64, 300000)

	bot.priceAdderChan = make(map[string]chan float64)

	for _, pair := range bot.pairs {
		bot.priceAdderChan[pair] = make(chan float64, 300000)
	}

	// -----------------------
	// Start Error monitoring
	// -----------------------
	go bot.errorMonitor()

	bot.stats = make(map[string]*statistician.Statistician)

	for _, pairName := range bot.pairs {
		bot.stats[pairName] = statistician.NewStatistician(bot.name, pairName, bot.kr, false, bot.sampleRate)
	}

	return bot
}

func (bot *Bot) errorMonitor() {
	log.Info("Starting error monitor service for exchange : ", bot.name)
	for {
		select {
		case err := <-bot.apiError:
			{
				if bot.apiOnline {
					log.Error("API Error detected: ", err.Error())
					bot.PublicRestart()
				}
				continue
			}

		case <-bot.fullStop:
			{
				log.Info("Exiting errorMonitor()")
				return
			}
		}
	}
}

//exchangeStart are commong Start functionality
func (bot *Bot) exchangeStart() {
	log.Info("Starting Public CEXIO collector: ", bot.pairs)
	//bot.kr.Start()

	priceUdateTimer := (time.Second * time.Duration(bot.sampleRate))

	log.Info("Price Update timer set to : ", priceUdateTimer)

	bot.priceUpdateTimer = time.NewTicker(priceUdateTimer)
	go bot.exchangeConnect()

}

func (bot *Bot) statsStart() {
	log.Info("Starting stats calculator: ", bot.pairs)

	statsUpdateTimer := (time.Second * time.Duration(bot.sampleRate))

	bot.statsUpdateTimer = time.NewTicker(statsUpdateTimer)
	go bot.statsCollector()

}

func (bot *Bot) PublicStart() {
	go bot.exchangeStart()
	go bot.statsStart()
}

func (bot *Bot) PublicRestart() {
	log.Info("Restarting public api connection...")
	bot.api.Close("BOT")
	bot.apiOnline = false

	//--------------------------
	// Stop price update timer
	//--------------------------
	bot.priceUpdateTimer.Stop()
	close(bot.apiStop)
	time.Sleep(time.Second * 2)
	log.Info("All conections closed, restarting...")

	bot.apiStop = make(chan bool)
	log.Info("Restarting exchangeConnect...")

	// ------------------------
	// Start common routines
	// ------------------------
	go bot.exchangeStart()

}

//Deprecated
func (bot *Bot) Start() {
	log.Info("Starting CEXIO collector")

	bot.kr.Start()
	bot.api.Connect()

	bot.apiOnline = true

	go bot.api.ResponseCollector()

	for _, pair := range bot.pairs {

		statusCount, err := bot.kr.GetCounter(bot.name, pair)

		if err != nil {
			log.Fatal("Could not get count value for pair :", pair)
		}

		log.Infof("Pair count %s : %d ", pair, statusCount)

	}

	go bot.monitorTicker("BTC", "USD")

	for {
		time.Sleep(time.Second)
	}

}

func (bot *Bot) exchangeConnect() {
	log.Info("ExchangeConnect running")
	err := bot.api.Connect()
	log.Debug("Connect completed, checking error...")
	if err != nil {
		log.Fatal("Could not connect to CEXIO websocket service: ", err.Error())
	}

	log.Info("Completed api.connect")

	bot.apiOnline = true

	go bot.api.ResponseCollector()

	bot.MonitorPrice()

}

func (bot *Bot) statsCollector() {

	for _, pair := range bot.pairs {

		statusCount, err := bot.kr.GetCounter(bot.name, pair)

		if err != nil {
			anError := fmt.Errorf("GetCounter : %s_%s -> %s", bot.name, pair, err.Error())
			log.Fatal(anError)
		}

		log.Infof("Pair count %s : %d ", pair, statusCount)
		go bot.UpdatePriceLists(bot.name, pair)

	}

}

func (bot *Bot) MonitorPrice() {
	monTimer := time.NewTicker(time.Second)
	priceLock := &sync.Mutex{}
	emaMapLock := &sync.Mutex{}

	priceUpdateMap := make(map[string]cexioapi.ResponseTickerSubData)
	priceUpdateEmaMap := make(map[string]ewma.MovingAverage)

	go bot.api.TickerSub(bot.tickerSub)
	log.Info("Waiting for price change...")
	for {
		select {
		case lPriceUpdate := <-bot.tickerSub:
			{
				pair := fmt.Sprintf("%s%s", lPriceUpdate.Symbol1, lPriceUpdate.Symbol2)
				key := fmt.Sprintf("PRICE_%s_%s", bot.name, pair)

				// -------------
				// Save price
				// -------------
				priceLock.Lock()
				priceUpdateMap[key] = lPriceUpdate
				priceLock.Unlock()
				// -------------------------------
				// Check if ewma already exists
				// it not, create it
				// --------------------------------
				emaMapLock.Lock()
				_, ok := priceUpdateEmaMap[key]

				if !ok {
					priceUpdateEmaMap[key] = ewma.NewMovingAverage(float64(bot.sampleRate / 2)) //((60 / bot.sampleRate) / 2) //ewma.NewMovingAverage(float64(bot.sampleRate / 2))
					priceFloat, err := strconv.ParseFloat(lPriceUpdate.Price, 64)
					if err != nil {
						log.Fatal("converting string to float: ", err.Error())
					}
					//for i := 0; i < 10; i++ {
					priceUpdateEmaMap[key].Set(priceFloat)
					//}

				}
				emaMapLock.Unlock()

				log.Infof("Price update: %s : %s", key, lPriceUpdate.Price)
				pair = fmt.Sprintf("%s_RAW", pair)
				bot.kr.Update(bot.name, pair, lPriceUpdate.Price)
			}
		case <-bot.apiStop:
			{
				log.Info("ApiStop detected, exiting MonitorPrice")
				//monTimer.Stop()
				return
			}

		case <-monTimer.C:
			{

				for key := range priceUpdateMap {

					priceLock.Lock()
					xPriceUpdate := priceUpdateMap[key]
					priceLock.Unlock()

					pair := fmt.Sprintf("%s%s", xPriceUpdate.Symbol1, xPriceUpdate.Symbol2)
					key := fmt.Sprintf("PRICE_%s_%s", bot.name, pair)

					if pair == "" {
						log.Info("pair warm up...", key)
					} else {

						priceFloat, err := strconv.ParseFloat(xPriceUpdate.Price, 64)
						if err != nil {
							log.Fatal("converting string to float: ", err.Error())
						}

						emaMapLock.Lock()
						priceUpdateEmaMap[key].Add(priceFloat)
						avgPriceString := fmt.Sprintf("%f", priceUpdateEmaMap[key].Value())
						emaMapLock.Unlock()
						//log.Infof("Real Price update: %s : %s, %s", key, avgPriceString, xPriceUpdate.Price)
						bot.kr.Update(bot.name, pair, avgPriceString)
					}
				}
			}
		}
	}
}

//UpdatePriceLists updates prcie list db entry
func (bot *Bot) UpdatePriceLists(exchange, pair string) {
	log.Debugf("Starting price update routine for : %s_%s ", exchange, pair)

	// -----------------------------------------------------
	// Start price updater, it will not update price to db
	// until recovery is done
	// -----------------------------------------------------
	go bot.priceUpdater(exchange, pair)

	// -----------------------------------------------
	// Get price list to recover prices statistics
	// -----------------------------------------------

	key := fmt.Sprintf("%s_%s", exchange, pair)
	log.Info("Sarting DB recovery for pair: ", pair)
	list, err := bot.kr.GetRange(key, bot.historyCount)
	if err != nil {
		log.Fatal("Update Price list: error recovering --> ", err.Error())
	}

	historySize := len(list)
	log.Info("History size is: ", historySize)

	// --------------------------------------------------------
	// Do not allow Db updates by statistician & friends while
	// recovery is done
	// --------------------------------------------------------

	if false {

		bot.stats[pair].SetDbUpdates(false)

		if historySize > 1 {
			size := int(len(list))
			for i := size - 1; i >= 0; i-- {
				valueStr := list[i]
				value, err := strconv.ParseFloat(valueStr, 64)
				if err != nil {
					log.Error("Empty value in db: ", err.Error())

				} else {
					//log.Infof("Adding value %f for pair %s", value, pair)
					bot.stats[pair].Add(value)
				}

			}
		}
		log.Infof("DB recovery complete , reprocessed %d entries for pair %s : ", len(list), pair)
	}
	// -----------------------------------------------
	// re-enable Db updates for statistician & friends
	// -----------------------------------------------
	bot.stats[pair].SetDbUpdates(true)
	time.Sleep(time.Second)

	// ----------------------------------------
	// Send broadcast to enable writing to db
	// by running price collector (priceAdder)
	// -----------------------------------------
	bot.priceUpdaterCond.Broadcast()
}

func (bot *Bot) priceUpdater(exchange, pair string) {
	go bot.priceAdder(pair)
	counter := 0

	for _ = range bot.statsUpdateTimer.C {
		//valueStr, err := bot.kr.UpdateList(exchange, pair)

		valueInterface, err := bot.kr.GetPriceValue(exchange, pair)
		if err != nil {
			log.Fatal("error obtaining price value :", bot.pairs)
		}

		valueStr, err := bot.kr.PushToPriceList(valueInterface, exchange, pair)
		if err != nil {
			log.Fatal("error updating list: ", err.Error())
		}
		value, err := strconv.ParseFloat(valueStr, 64)

		//priceEma.Add(value)
		//log.Info("update value:", value)
		if value != 0 {
			bot.priceAdderChan[pair] <- value
		}
		//bot.stats[pair].Add(value)

		if counter%(60*5) == 0 {
			log.Infof("Saving value, exchange: %s , pair %s , value :%f", exchange, pair, value)
		}
		counter++
		daemon.SdNotify(false, "WATCHDOG=1")
	}
	log.Info("UpdatePriceLists exiting...")

}

func (bot *Bot) priceAdder(pair string) {

	bot.priceUpdaterCond.L.Lock()
	log.Info("priceAdder Waiting for pair :", pair)
	bot.priceUpdaterCond.Wait()
	bot.priceUpdaterCond.L.Unlock()
	log.Info("priceAdder waken up for pair :", pair)

	for {
		select {
		case value := <-bot.priceAdderChan[pair]:
			{

				log.Infof("update value for pair %s value: %f", pair, value)
				bot.stats[pair].Add(value)
			}

		case _ = <-bot.fullStop:
			{
				return
			}

		}
	}
}

func (bot *Bot) Stop() {
	bot.priceUpdateTimer.Stop()
	bot.statsUpdateTimer.Stop()
	err := bot.api.Close("MainStop")
	if err != nil {
		log.Fatal("error while stoping bot:", err.Error())
	}
	close(bot.fullStop)
	log.Info("Bot is down")
}

func (bot *Bot) runMonitors() {

}

func (bot *Bot) monitorTicker(cCode1, cCode2 string) {

	log.Info("Starting CEXIO monitor ticker")

	go bot.api.TickerSub(bot.tickerSub)

}
