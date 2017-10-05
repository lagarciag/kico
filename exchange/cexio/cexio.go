package cexio

import (
	cexioapi "github.com/lagarciag/cexioapi"
	log "github.com/sirupsen/logrus"

	"time"

	"sync"

	"fmt"

	"github.com/lagarciag/tayni/kredis"
)

const (
	startState      = "Start"
	restartState    = "restart"
	initializeState = "Initialize"
	warmUpState     = "WarmUp"
	watchState      = "Watch"
)

const (
	startEventName          = "start"
	restartEventName        = "restart"
	warmUpEventName         = "warmUp"
	warmUpCompleteEventName = "warmUpComplete"
)

/*
2017/08/29 06:05:58 websocket: close 1006 (abnormal closure): unexpected EOF
2017/08/29 07:26:34 websocket: close 1001 (going away): CloudFlare WebSocket proxy restarting
*/

const minuteTicks = 30
const warmUpCycles = (60 * 60 * 24 * 26) / 2

type CollectorConfig struct {
	CexioKey    string
	CexioSecret string
	Pairs       []string
}

type Bot struct {
	name           string
	key            string
	secret         string
	pairs          []string
	api            *cexioapi.API
	apiError       chan error
	apiLock        *sync.Mutex
	kr             *kredis.Kredis
	tickerSub      chan cexioapi.ResponseTickerSubData
	ticksPerMinute int
	btcUsdBase     float64
	stop           chan bool
	apiStop        chan bool
	apiOnline      bool
}

func NewBot(config CollectorConfig) (bot *Bot) {

	//--------------------------------
	//Move this to a secure location
	//--------------------------------
	bot = &Bot{}
	bot.name = "CEXIO"
	bot.key = config.CexioKey
	bot.pairs = config.Pairs
	bot.secret = config.CexioSecret
	bot.kr = kredis.NewKredis(1300000)
	bot.ticksPerMinute = minuteTicks
	bot.api, bot.apiError = cexioapi.NewPublicAPI()
	bot.tickerSub = make(chan cexioapi.ResponseTickerSubData)
	bot.apiStop = make(chan bool)

	return bot
}

func (bot *Bot) errorMonitor() {

	for {
		select {
		case err := <-bot.apiError:
			{
				log.Error("API Error detected: ", err.Error())
				bot.PublicRestart()
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

func (bot *Bot) PublicRestart() {
	log.Info("Restarting public api connection...")
	bot.api.Close("BOT")
	bot.apiOnline = false
	close(bot.apiStop)
	log.Info("Post close...")
	time.Sleep(time.Second * 2)
	log.Info("All conections closed, restarting...")
	bot.apiStop = make(chan bool)
	log.Info("Restarting exchangeConnect...")
	bot.kr.Start()
	go bot.exchangeConnect()
	log.Debug("PublicRestart Complete...")
}

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

func (bot *Bot) PublicStart() {
	log.Info("Starting Public CEXIO collector")
	bot.kr.Start()
	go bot.errorMonitor()
	go bot.exchangeConnect()

	<-bot.stop

	log.Info("PublicStart finished")

}

func (bot *Bot) UpdatePriceLists(exchange, pair string) {
	log.Debugf("Starting price update routine for : %s_%s ", exchange, pair)
	for bot.apiOnline {
		bot.kr.UpdateList(exchange, pair)
		time.Sleep(time.Second * 2)
	}
}

func (bot *Bot) Stop() {
	err := bot.api.Close("MainStop")
	if err != nil {
		log.Fatal("error while stoping bot:", err.Error())
	}

	log.Info("Bot is down")
}

func (bot *Bot) runMonitors() {

}

func (bot *Bot) monitorTicker(cCode1, cCode2 string) {

	log.Info("Starting CEXIO monitor ticker")

	go bot.api.TickerSub(bot.tickerSub)

}
