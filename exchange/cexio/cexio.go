package cexio

import (
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"

	cexioapi "github.com/lagarciag/cexioapi"
	"github.com/lagarciag/tayni/kredis"
)

type CollectorConfig struct {
	CexioKey     string
	CexioSecret  string
	Pairs        []string
	SampleRate   int
	HistoryCount int
}

type Automata interface {
	Start()
	Restart()
	Stop()
}

type Bot struct {
	Name         string
	Key          string
	Secret       string
	Pairs        []string
	SampleRate   int
	HistoryCount int
	Api          *cexioapi.API
	Kr           *kredis.Kredis

	Orders chan cexioapi.ResponseOrder

	ApiOnline bool
	FullStop  chan bool
	ApiStop   chan bool
	ApiError  chan error
}

// ---------------------
// Public Methods
// ---------------------

//Start starts the bot
func (bot *Bot) Start() {
	log.Info("Running start...")
	go bot.errorMonitor()
	go bot.exchangeConnect()
}

//Restart restarts the bot
func (bot *Bot) Restart() {
	log.Info("Restarting Api connection...")
	if err := bot.Api.Close("BOT"); err != nil {
		log.Error(err.Error())
	}

/*
	bot.apiOnline = false

	//--------------------------
	// Stop price update timer
	//--------------------------
	for _, pair := range bot.pairs {
		if bot.priceUpdateTimer[pair] != nil {
			bot.priceUpdateTimer[pair].Stop()
		}
	}
*/
	bot.ApiOnline = false

	close(bot.ApiStop)
	time.Sleep(time.Second * 2)
	log.Info("All conections closed, restarting...")

	bot.ApiStop = make(chan bool)
	log.Info("Restarting exchangeConnect...")

	// ------------------------
	// Start common routines
	// ------------------------
	go bot.exchangeConnect()

}

//Stop stops the bot
func (bot *Bot) Stop() {

	err := bot.Api.Close("MainStop")
	if err != nil {
		log.Fatal("error while stoping bot:", err.Error())
	}
	close(bot.FullStop)
	log.Info("Bot is down")
}

//GetBalance returns stored balances in exchange
func (bot *Bot) GetBalance() (tickerBalanceMap map[string]float64, err error) {

	balance, err := bot.Api.GetBalance()
	if err != nil {
		log.Error("Error obtaining balance: ", err.Error())
		return tickerBalanceMap, err
	}

	stringMap := balance.Data.Balance
	tickerBalanceMap = make(map[string]float64)
	for key, stringFloat := range stringMap {
		aFloat, err := strconv.ParseFloat(stringFloat, 64)
		if err != nil {
			log.Error("Error parsing float string: ", err.Error())
			return tickerBalanceMap, err
		}
		tickerBalanceMap[key] = aFloat
	}
	return tickerBalanceMap, err
}

func (bot *Bot) GetTickerPrice(cc1 string, cc2 string) (bid float64, ask float64, err error) {

	priceResponce, err := bot.Api.Ticker(cc1, cc2)

	if err != nil {
		log.Error("nil response: ", cc1, cc2)
		return bid, ask, err
	}
	bid = priceResponce.Data.Bid
	ask = priceResponce.Data.Ask

	return bid, ask, err
}

func (bot *Bot) PlaceOrder(cc1 string, cc2 string, amount, price float64, aType string) (complete bool,
	pending float64,
	amountPlaced float64,
	transactionID string,
	orderID string,
	err error) {

	resp, err := bot.Api.PlaceOrder(cc1, cc2, amount, price, aType)
	if err != nil {
		return complete, pending, amountPlaced, transactionID, orderID, err
	}

	log.Debug("PlaceOrder response: \n%s", spew.Sdump(resp))

	complete = resp.Data.Complete
	pendingString := resp.Data.Pending
	amountString := resp.Data.Amount
	transactionID = resp.Oid
	orderID = resp.Data.ID

	pending, err = strconv.ParseFloat(pendingString, 64)
	if err != nil {
		log.Error("Error parsing float string: ", err.Error())
	}

	amount, err = strconv.ParseFloat(amountString, 64)
	if err != nil {
		log.Error("Error parsing float string: ", err.Error())
	}

	return complete, pending, amountPlaced, transactionID, orderID, err
}

func (bot *Bot) GetOpenOrdersList(cc1 string, cc2 string) (ordersList []cexioapi.ResponseOpenOrdersData,
	err error) {

	ordersList, err = bot.Api.GetOpenOrdersList(cc1, cc2)
	if err != nil {
		log.Error("GetOpenOrdersList: ", err.Error())
		return ordersList, err
	}

	return ordersList, err
}

func (bot *Bot) WaitForOrder(orderID string) (completeCancel bool, err error) {
	completeCancel, err = bot.Api.WaitOrderComplete(orderID)
	return completeCancel, err
}
