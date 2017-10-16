package cexio

import (
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

const minuteTicks = 30

type CollectorConfig struct {
	CexioKey    string
	CexioSecret string
	Pairs       []string
}

type Bot struct {
	name   string
	key    string
	secret string
	pairs  []string
	api    *cexioapi.API

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
	shutdownCond     *sync.Cond
	apiError         chan error
	stop             chan bool
	apiStop          chan bool
	apiOnline        bool
}

func NewBot(config CollectorConfig, kr *kredis.Kredis) (bot *Bot) {

	//--------------------------------
	//Move this to a secure location
	//--------------------------------
	bot = &Bot{}

	bot.name = "CEXIO"
	bot.key = config.CexioKey
	bot.pairs = config.Pairs
	bot.secret = config.CexioSecret
	bot.kr = kr
	bot.ticksPerMinute = minuteTicks
	bot.api, bot.apiError = cexioapi.NewPublicAPI()

	bot.tickerSub = make(chan cexioapi.ResponseTickerSubData)

	bot.shutdownCond = sync.NewCond(&sync.Mutex{})
	bot.apiStop = make(chan bool)
	bot.stop = make(chan bool)
	// -----------------------
	// Start Error monitoring
	// -----------------------
	go bot.errorMonitor()

	bot.stats = make(map[string]*statistician.Statistician)

	cID := 0
	for _, pairName := range bot.pairs {
		bot.stats[pairName] = statistician.NewStatistician(bot.name, bot.pairs[cID], bot.kr, false)
		cID++
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

		case <-bot.stop:
			{
				log.Info("Exiting errorMonitor()")
				return
			}
		}
	}
}

//pStart are commong Start functionality
func (bot *Bot) pStart() {
	log.Info("Starting Public CEXIO collector")
	bot.kr.Start()
	bot.priceUpdateTimer = time.NewTicker(time.Second * 2)
	go bot.exchangeConnect()

}

func (bot *Bot) PublicStart() {
	go bot.pStart()
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
	go bot.pStart()

}

//Deprecated
func (bot *Bot) Start() {
	log.Info("Starting CEXIO collector")

	bot.kr.Start()
	bot.api.Connect()

	bot.apiOnline = true

	go bot.api.ResponseCollector()

	for _, pair := range bot.pairs {

		//ccList := pairs.PairsHash[pair]

		//code1 := ccList[0]
		//code2 := ccList[1]

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

	for _, pair := range bot.pairs {

		//ccList := pairs.PairsHash[pair]

		//code1 := ccList[0]
		//code2 := ccList[1]

		statusCount, err := bot.kr.GetCounter(bot.name, pair)

		if err != nil {
			anError := fmt.Errorf("GetCounter : %s_%s -> %s", bot.name, pair, err.Error())
			log.Fatal(anError)
		}

		log.Infof("Pair count %s : %d ", pair, statusCount)
		go bot.UpdatePriceLists(bot.name, pair)

	}

	bot.MonitorPrice()

}

func (bot *Bot) MonitorPrice() {

	go bot.api.TickerSub(bot.tickerSub)
	log.Info("Waiting for price change...")
	for {
		select {
		case priceUpdate := <-bot.tickerSub:
			{
				pair := fmt.Sprintf("%s%s", priceUpdate.Symbol1, priceUpdate.Symbol2)
				key := fmt.Sprintf("PRICE_%s_%s", bot.name, pair)
				log.Infof("Price update: %s : %s", key, priceUpdate.Price)
				bot.kr.Update(bot.name, pair, priceUpdate.Price)
			}
		case <-bot.apiStop:
			{
				log.Info("ApiStop detected, exiting MonitorPrice")
				return
			}
		}
	}
}

func (bot *Bot) UpdatePriceLists(exchange, pair string) {
	log.Debugf("Starting price update routine for : %s_%s ", exchange, pair)
	counter := 0
	for _ = range bot.priceUpdateTimer.C {
		valueStr, err := bot.kr.UpdateList(exchange, pair)

		if err != nil {
			log.Fatal("error updating list: ", err.Error())
		}
		value, err := strconv.ParseFloat(valueStr, 64)
		bot.stats[pair].Add(value)

		if counter%(60*5) == 0 {
			log.Infof("Saving value, exchange: %s , pair %s , value :%f", exchange, pair, value)
		}
		counter++
		daemon.SdNotify(false, "WATCHDOG=1")
	}
	log.Info("UpdatePriceLists exiting...")
}

func (bot *Bot) Stop() {
	bot.priceUpdateTimer.Stop()
	err := bot.api.Close("MainStop")
	if err != nil {
		log.Fatal("error while stoping bot:", err.Error())
	}
	close(bot.stop)
	log.Info("Bot is down")
}

func (bot *Bot) runMonitors() {

}

func (bot *Bot) monitorTicker(cCode1, cCode2 string) {

	log.Info("Starting CEXIO monitor ticker")

	go bot.api.TickerSub(bot.tickerSub)

}
