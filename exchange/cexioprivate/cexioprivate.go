package cexioprivate

import (
	cexioapi "github.com/lagarciag/cexioapi"
	"github.com/lagarciag/tayni/exchange/cexio"

	"sync"

	"github.com/lagarciag/tayni/kredis"
)

type Bot struct {
	cexio.Bot
	apiLock      *sync.Mutex
	shutdownCond *sync.Cond
}

func NewBot(config cexio.CollectorConfig, kr *kredis.Kredis, publicConnect bool) (bot *Bot) {

	//--------------------------------
	//Move this to a secure location
	//--------------------------------
	bot = &Bot{}
	bot.Orders = make(chan cexioapi.ResponseOrder, 100)
	bot.HistoryCount = config.HistoryCount
	bot.Name = "CEXIO"
	bot.Key = config.CexioKey
	bot.Pairs = config.Pairs
	bot.SampleRate = config.SampleRate
	bot.Secret = config.CexioSecret
	bot.Kr = kr
	bot.Kr.Start()
	bot.Api, bot.ApiError = cexioapi.NewAPI(bot.Key, bot.Secret, publicConnect)

	bot.shutdownCond = sync.NewCond(&sync.Mutex{})
	bot.ApiStop = make(chan bool)
	bot.FullStop = make(chan bool)

	return bot
}
