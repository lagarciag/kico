package kicobot

import (
	cexioapi "github.com/lagarciag/cexioapi"
	log "github.com/sirupsen/logrus"

	"fmt"
	"time"

	"sync"

	"github.com/lagarciag/kico/statistician"

	"github.com/looplab/fsm"
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

type BotConfig struct {
	CexioKey    string
	CexioSecret string
}

type Bot struct {
	key            string
	secret         string
	api            *cexioapi.API
	apiLock        *sync.Mutex
	ticksPerMinute int
	btcUsdBase     float64
	statistician   *statistician.Statistician

	fsm                 *fsm.FSM
	startEvent          fsm.EventDesc
	restartEvent        fsm.EventDesc
	warmUpEvent         fsm.EventDesc
	warmUpCompleteEvent fsm.EventDesc

	eventsList    fsm.Events
	callbacksList fsm.Callbacks

	event chan string
}

func NewBot(config BotConfig) (bot *Bot) {

	//--------------------------------
	//Move this to a secure location
	//--------------------------------
	bot = &Bot{}
	bot.key = config.CexioKey
	bot.secret = config.CexioSecret

	bot.apiLock = &sync.Mutex{}
	bot.statistician = statistician.NewStatistician(true)
	bot.ticksPerMinute = minuteTicks
	bot.api = cexioapi.NewAPI(bot.key, bot.secret)
	bot.event = make(chan string, 1)

	// -----------------------
	// Initialize FSM machine
	// -----------------------

	bot.startEvent = fsm.EventDesc{Name: startEventName, Src: []string{startState, restartState}, Dst: initializeState}

	bot.restartEvent = fsm.EventDesc{Name: restartEventName,
		Src: []string{startState, warmUpState, initializeState, watchState},
		Dst: initializeState}
	bot.warmUpEvent = fsm.EventDesc{Name: warmUpEventName, Src: []string{initializeState}, Dst: warmUpState}

	bot.warmUpCompleteEvent = fsm.EventDesc{Name: warmUpCompleteEventName, Src: []string{warmUpState}, Dst: watchState}

	bot.eventsList = fsm.Events{
		bot.startEvent,
		bot.warmUpEvent,
		bot.warmUpCompleteEvent}

	bot.callbacksList = fsm.Callbacks{
		initializeState: bot.callBackInInitializeState,
		warmUpState:     bot.callBackInWarmUpState,
		watchState:      bot.callBackInWatchState,
	}

	bot.fsm = fsm.NewFSM(
		startState,
		bot.eventsList,
		bot.callbacksList,
	)

	return bot
}

func (bot *Bot) Start() {
	go bot.runEventManager()
	log.Info("Machine State starting:", bot.fsm.Current())

	log.Debug("Event: ", startEventName)
	bot.fsm.Event(startEventName)
}

func (bot *Bot) Stop() {
	err := bot.api.Close("MainStop")
	if err != nil {
		log.Fatal("error while stoping bot:", err.Error())
	}

	log.Info("Bot is down")
}

/*
func (bot *Bot) monitorBalance() {
	for {
		balance, err := bot.api.GetBalance()

		if err != nil {
			log.Error("Error reading balance:", err.Error())
		}

		newUSD, err := strconv.ParseFloat(balance.Data.Balance.USD, 64)

		if err != nil {
			log.Error("Could not convert string to float64, USD balance")
		}

		newBTC, err := strconv.ParseFloat(balance.Data.Balance.BTC, 64)

		if err != nil {
			log.Error("Could not convert string to float64, BTC balance")
		}

		if bot.balance.USD() != newUSD {
			bot.balance.SetUSD(newUSD)
			fmt.Println("USD balance", balance.Data.Balance.USD)

		}
		if bot.balance.BTC() != newBTC {
			bot.balance.SetBTC(newBTC, bot.average.Get5MinAverage())
			fmt.Println("BTC balance", balance.Data.Balance.BTC, bot.balance.btcUsdBase)
		}

		//av5min := bot.average.Get5MinAverage()

		time.Sleep(time.Second * 10)
	}

}
*/

func (bot *Bot) runMonitors() {
	go bot.monitorTicker("BTC", "USD")
	//bot.monitorOhlcvNew("BTC", "USD")
	//go bot.monitorBalance()
}

func (bot *Bot) monitorOhlcvNew(cCode1, cCode2 string) {
	log.Info("requestiong ohlcv...")
	rest, err := bot.api.InitOhlcvNew(cCode1, cCode2)
	log.Info("requestiong ohlcv done!")
	if err != nil {
		log.Error("OhlcvNew Error:", err.Error())
	}

	fmt.Println(rest)

}

func (bot *Bot) monitorTicker(cCode1, cCode2 string) {
	count := 0

	for {
		startTime := time.Now()
		ticker, err := bot.api.Ticker(cCode1, cCode2)
		if err != nil {
			log.Error("Unhandled error:", err.Error())
		}

		if ticker == nil {
			log.Error("Ticker is nil, continue")
			continue
		}

		elapsed := time.Since(startTime)

		if elapsed.Seconds() > 1 {
			log.Warnf("Elapsed time > 1 second : %s", elapsed)
		}

		bot.statistician.Add(ticker.Data.Ask)
		log.Info("Adding value: ", ticker.Data.Ask)
		count++
		time.Sleep(time.Second * (60 / minuteTicks))

	}

}

/*
func (bot *Bot) warmUpTickerGet(cCode1, cCode2 string) {
	count := 0
	log.Info("Warming up...")
	fmt.Println("\n")
	warmUpCount := 5
	priceSum := float64(0)
	for count < warmUpCount {

		ticker, err := bot.api.Ticker(cCode1, cCode2)
		if err != nil {
			log.Error("Unhandled error:", err.Error())
		}

		if ticker == nil {
			log.Error("Ticker is nil, continue")
			continue
		}

		//bot.statistician.Add(ticker.Data.Ask)

		priceSum = priceSum + ticker.Data.Ask

		count++
		time.Sleep(time.Second * (60 / minuteTicks))

		fmt.Print(".")

	}
	fmt.Println("\n")
	log.Debug("WarmUp tick collection complete")

	bot.api.Close("WarmUp")

	//average, err := bot.statistician.Sma(statistician.Minute)

	//if err != nil {
	//	log.Fatal("Fatal error calculating simple moving average: ", err.Error())
	//}

	average := priceSum / float64(warmUpCount)

	log.Info("Warm up average: ", average)

	timeNow := time.Now()

	log.Info("Adding history data...")

	for i := 0; i < warmUpCycles; i++ {
		bot.statistician.Add(average)

	}
	elapsedTime := time.Since(timeNow)
	log.Infof("Finished warmup process in: %s seconds", elapsedTime)

}
*/
