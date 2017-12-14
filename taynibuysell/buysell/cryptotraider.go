package buysell

import (
	"fmt"
	"strings"

	"github.com/lagarciag/tayni/kredis"
	"github.com/lagarciag/tayni/twitter"
	"github.com/looplab/fsm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	StartState    = "StartState"
	IdleState     = "IdleState"
	TradingState  = "TradingState"
	HoldState     = "HoldState"
	ShutdownState = "ShutdownState"
	DoBuyState    = "DoBuyState"
	DoSellState   = "DoSellState"

	// --------------
	// Buy States
	// --------------
	BuyBTCState = "BuyBTCState"
	BuyBCHState = "BuyBCHState"
	BuyZECState = "BuyZECState"
	BuyBGDState = "BuyBGDState"
	BuyETHState = "BuyETHState"

	// -------------
	//   Sell Sates
	// -------------

	SellBTCState = "SellBTCState"
	SellBCHState = "SellBCHState"
	SellZECState = "SellZECState"
	SellBGDState = "SellBGDState"
	SellETHState = "SellETHState"

/*
	BuyBTCState
	BuyBCHState
	BuyZECState
	BuyBGDState
	BuyETHState

	SellBTCState
	SellBCHState
	SellZECState
	SellBGDState
	SellETHState
*/
)

const (
	StartEvent    = "startEvent"
	StopEvent     = "stopEvent"
	ShutdownEvent = "shutdownEvent"
	TradeEvent    = "TradeEvent"
	Test1Event    = "Test2Event"

	DoBuyEvent        = "DoBuyEvent"
	DoSellEvent       = "DoSellEvent"
	BuyCompleteEvent  = "BuyCompleteEvent"
	SellCompleteEvent = "SellCompleteEvent"

	// --------------
	// Buy Events
	// --------------

	BuyBTCEvent = "BuyBTCEvent"
	BuyBCHEvent = "BuyBCHEvent"
	BuyZECEvent = "BuyZECEvent"
	BuyBGDEvent = "BuyBGDEvent"
	BuyETHEvent = "BuyETHEvent"

	SellBTCEvent = "SellBtcEvent"
	SellBCHEvent = "SellBCHEvent"
	SellZECEvent = "SellZECEvent"
	SellBGDEvent = "SellBGDEvent"
	SellETHEvent = "SellETHEvent"
)

type Message struct {
	Event string
	Signal bool
}


type CryptoSelectorFsm struct {
	tc *twitter.TwitterClient
	kr *kredis.Kredis
	To string

	eventsStringList    []string
	statesStringList    []string
	redisMessagesBuyMap map[string]string
	redisMessagesSellMap map[string]string

	FSM *fsm.FSM

	// ------------
	// Events
	// ------------
	startEvent    fsm.EventDesc
	stopEvent     fsm.EventDesc
	shutdownEvent fsm.EventDesc
	tradeEvent    fsm.EventDesc

	buyBTCEvent fsm.EventDesc
	buyBCHEvent fsm.EventDesc
	buyZECEvent fsm.EventDesc
	buyBGDEvent fsm.EventDesc
	buyETHEvent fsm.EventDesc

	sellBTCEvent fsm.EventDesc
	sellBCHEvent fsm.EventDesc
	sellZECEvent fsm.EventDesc
	sellBGDEvent fsm.EventDesc
	sellETHEvent fsm.EventDesc

	// ------------------
	// Events List
	// ------------------
	eventsList fsm.Events

	// ------------------
	// Fsm Callbacks
	callbacks fsm.Callbacks

	ChanStartEvent     chan bool
	ChanStopEvent      chan bool
	ChanShutdownEvent  chan bool
	ChanTradeEvent     chan bool
	testChanTradeEvent chan bool

	// --------------
	// Buy Events
	// --------------

	ChanDoBuyEvent        chan bool
	ChanDoSellEvent       chan bool
	ChanBuyCompleteEvent  chan bool
	ChanSellCompleteEvent chan bool

	ChanBuyBTCEvent chan bool
	ChanBuyBCHEvent chan bool
	ChanBuyZECEvent chan bool
	ChanBuyBGDEvent chan bool
	ChanBuyETHEvent chan bool

	ChanSellBTCEvent      chan bool
	ChanSellBCHEvent      chan bool
	ChanSellZECEvent      chan bool
	ChanSellBGDEvent      chan bool
	ChanSellETHEvent      chan bool
	ChanMessageMap        map[string]chan Message
	ChanMapForRedisEvents map[string]chan Message
}

func NewCryptoTradeFsm(pairList []string) *CryptoSelectorFsm {
	log.Info("Creating new crytop trading fsm for pairs: ", pairList)

	tFsm := &CryptoSelectorFsm{}

	tFsm.eventsStringList = []string{StartEvent,
		StopEvent,
		ShutdownEvent,
		TradeEvent}
	tFsm.statesStringList = []string{StartState,
		IdleState,
		TradingState,
		HoldState,
		ShutdownState,
	}

	tFsm.redisMessagesBuyMap = make(map[string]string)
	tFsm.redisMessagesSellMap = make(map[string]string)

	//CEXIO_BTC_BUY"

	for _, aPair := range pairList {

		pairX := strings.Split(aPair, "BTC")
		pair := pairX[0]

		buyEventName := fmt.Sprintf("Buy%sEvent", pair)
		sellEventName := fmt.Sprintf("Sell%sEvent", pair)

		tFsm.eventsStringList = append(tFsm.eventsStringList, buyEventName)
		tFsm.eventsStringList = append(tFsm.eventsStringList, sellEventName)

		buyStateName := fmt.Sprintf("Buy%sState", pair)
		sellStateName := fmt.Sprintf("Sell%sState", pair)

		tFsm.statesStringList = append(tFsm.statesStringList, buyStateName)
		tFsm.statesStringList = append(tFsm.statesStringList, sellStateName)

		tFsm.redisMessagesBuyMap[pair] = fmt.Sprintf("CEXIO_%s_BUY", pair)
		tFsm.redisMessagesSellMap[pair] = fmt.Sprintf("CEXIO_%s_SELL", pair)

	}

	log.Info("EVENTS: ", tFsm.eventsStringList)
	log.Info("STATES: ", tFsm.statesStringList)

	tFsm.kr = kredis.NewKredis(1000000)
	tFsm.kr.Start()
	config := twitter.Config{}

	vTwitterConfig := viper.Get("twitter").(map[string]interface{})
	config.Twit = vTwitterConfig["twit"].(bool)
	config.ConsumerKey = vTwitterConfig["consumer_key"].(string)
	config.ConsumerSecret = vTwitterConfig["consumer_secret"].(string)
	config.AccessToken = vTwitterConfig["access_token"].(string)
	config.AccessTokenSecret = vTwitterConfig["access_token_secret"].(string)

	if config.ConsumerKey == "" {
		log.Fatal("bad consumerkey")
	}

	tFsm.tc = twitter.NewTwitterClient(config)

	// ------------
	// Events
	// ------------
	startEvent := fsm.EventDesc{Name: StartEvent, Src: []string{StartState}, Dst: IdleState}

	stopEvent := fsm.EventDesc{Name: StopEvent,
		Src: []string{
			TradingState,
			BuyBTCState,
			BuyBCHState,
			BuyZECState,
			BuyBGDState,
			BuyETHState,
			HoldState,
		}, Dst: IdleState}

	tradeEvent := fsm.EventDesc{Name: TradeEvent, Src: []string{IdleState}, Dst: TradingState}

	// ----------------------
	// Buying related events
	// ----------------------

	buyBTCEvent := fsm.EventDesc{Name: BuyBTCEvent,
		Src: []string{TradingState},
		Dst: BuyBTCState}

	buyBCHEvent := fsm.EventDesc{Name: BuyBCHEvent,
		Src: []string{TradingState},
		Dst: BuyBCHState}

	buyZECEvent := fsm.EventDesc{Name: BuyZECEvent,
		Src: []string{TradingState},
		Dst: BuyZECState}

	buyBGDEvent := fsm.EventDesc{Name: BuyBGDEvent,
		Src: []string{TradingState},
		Dst: BuyBGDState}

	buyETHEvent := fsm.EventDesc{Name: BuyETHEvent,
		Src: []string{TradingState},
		Dst: BuyETHState}

	// ----------------------
	// Selling related events
	// ----------------------

	sellBTCEvent := fsm.EventDesc{Name: SellBTCEvent,
		Src: []string{TradingState},
		Dst: SellBTCState}

	sellBCHEvent := fsm.EventDesc{Name: SellBCHEvent,
		Src: []string{TradingState},
		Dst: SellBCHState}

	sellZECEvent := fsm.EventDesc{Name: SellZECEvent,
		Src: []string{TradingState},
		Dst: SellZECState}

	sellBGDEvent := fsm.EventDesc{Name: SellBGDEvent,
		Src: []string{TradingState},
		Dst: SellBGDState}

	sellETHEvent := fsm.EventDesc{Name: SellETHEvent,
		Src: []string{TradingState},
		Dst: SellETHState}

	tFsm.startEvent = startEvent
	tFsm.stopEvent = stopEvent
	tFsm.tradeEvent = tradeEvent
	tFsm.buyBTCEvent = buyBTCEvent
	tFsm.buyBCHEvent = buyBCHEvent
	tFsm.buyZECEvent = buyZECEvent
	tFsm.buyBGDEvent = buyBGDEvent
	tFsm.buyETHEvent = buyETHEvent

	tFsm.sellBTCEvent = sellBTCEvent
	tFsm.sellBCHEvent = sellBCHEvent
	tFsm.sellZECEvent = sellZECEvent
	tFsm.sellBGDEvent = sellBGDEvent
	tFsm.sellETHEvent = sellETHEvent

	tFsm.shutdownEvent = fsm.EventDesc{Name: ShutdownEvent,
		Src: []string{StartState,
			IdleState},
		Dst: ShutdownState}

	// ----------------------
	// Events List Registry
	// ----------------------
	tFsm.eventsList = fsm.Events{
		tFsm.startEvent,
		tFsm.stopEvent,
		tFsm.shutdownEvent,
		tFsm.tradeEvent,
		tFsm.buyBTCEvent,
		tFsm.buyBTCEvent,
		tFsm.buyBCHEvent,
		tFsm.buyZECEvent,
		tFsm.buyBGDEvent,
		tFsm.buyETHEvent,

		tFsm.sellBTCEvent,
		tFsm.sellBCHEvent,
		tFsm.sellZECEvent,
		tFsm.sellBGDEvent,
		tFsm.sellETHEvent,
	}

	// -------------------
	// Callbacks registry
	// -------------------
	tFsm.callbacks = fsm.Callbacks{
		StartState:    tFsm.CallBackInStartState,
		IdleState:     tFsm.CallBackInIdleState,
		ShutdownState: tFsm.CallBackInShutdownState,
		TradingState:  tFsm.CallBackInTradingState,

		DoBuyEvent:        tFsm.CallBackInDoBuyState,
		DoSellEvent:       tFsm.CallBackInDoSellState,
		BuyCompleteEvent:  tFsm.CallBackInBuyCompleteState,
		SellCompleteEvent: tFsm.CallBackInSellCompleteState,

		BuyBTCEvent: tFsm.CallBackInState,
		BuyBCHEvent: tFsm.CallBackInState,
		BuyZECEvent: tFsm.CallBackInState,
		BuyBGDEvent: tFsm.CallBackInState,
		BuyETHEvent: tFsm.CallBackInState,

		SellBTCEvent: tFsm.CallBackInState,
		SellBCHEvent: tFsm.CallBackInState,
		SellZECEvent: tFsm.CallBackInState,
		SellBGDEvent: tFsm.CallBackInState,
		SellETHEvent: tFsm.CallBackInState,
	}

	// ------------------
	// Event Channels
	// ------------------

	/*
	tFsm.ChanStartEvent = make(chan bool)
	tFsm.ChanStopEvent = make(chan bool)
	tFsm.ChanShutdownEvent = make(chan bool)
	tFsm.ChanTradeEvent = make(chan bool)
	tFsm.testChanTradeEvent = make(chan bool)

	// --------------
	// Buy Events
	// --------------

	tFsm.ChanDoBuyEvent = make(chan bool, 1)
	tFsm.ChanDoSellEvent = make(chan bool, 1)
	tFsm.ChanBuyCompleteEvent = make(chan bool, 1)
	tFsm.ChanSellCompleteEvent = make(chan bool, 1)

	tFsm.ChanBuyBTCEvent = make(chan bool, 1)
	tFsm.ChanBuyBCHEvent = make(chan bool, 1)
	tFsm.ChanBuyZECEvent = make(chan bool, 1)
	tFsm.ChanBuyBGDEvent = make(chan bool, 1)
	tFsm.ChanBuyETHEvent = make(chan bool, 1)

	tFsm.ChanSellBTCEvent = make(chan bool, 1)
	tFsm.ChanSellBCHEvent = make(chan bool, 1)
	tFsm.ChanSellZECEvent = make(chan bool, 1)
	tFsm.ChanSellBGDEvent = make(chan bool, 1)
	tFsm.ChanSellETHEvent = make(chan bool, 1)
    */
	tFsm.ChanMessageMap = make(map[string]chan Message)

	for _, event  := range tFsm.eventsStringList {
		tFsm.ChanMessageMap[event] = make(chan Message)
	}


	// -------------
	// FSM creation
	// -------------
	tFsm.FSM = fsm.NewFSM(StartState,
		tFsm.eventsList,
		tFsm.callbacks)

	tFsm.ChanMapForRedisEvents = make(map[string]chan Message)




	tFsm.ChanMapForRedisEvents[fmt.Sprintf("CEXIO_BTC_BUY")] = tFsm.ChanBuyBTCEvent
	tFsm.ChanMapForRedisEvents[fmt.Sprintf("CEXIO_BCH_BUY")] = tFsm.ChanBuyBTCEvent
	tFsm.ChanMapForRedisEvents[fmt.Sprintf("CEXIO_ZEC_BUY")] = tFsm.ChanBuyBTCEvent
	tFsm.ChanMapForRedisEvents[fmt.Sprintf("CEXIO_BGD_BUY")] = tFsm.ChanBuyBTCEvent
	tFsm.ChanMapForRedisEvents[fmt.Sprintf("CEXIO_ETH_BUY")] = tFsm.ChanBuyBTCEvent
	tFsm.ChanMapForRedisEvents[fmt.Sprintf("CEXIO_DASH_BUY")] = tFsm.ChanBuyBTCEvent

	tFsm.ChanMapForRedisEvents[fmt.Sprintf("CEXIO_BTC_SELL")] = tFsm.ChanSellBTCEvent
	tFsm.ChanMapForRedisEvents[fmt.Sprintf("CEXIO_BCH_SELL")] = tFsm.ChanSellBTCEvent
	tFsm.ChanMapForRedisEvents[fmt.Sprintf("CEXIO_ZEC_SELL")] = tFsm.ChanSellBTCEvent
	tFsm.ChanMapForRedisEvents[fmt.Sprintf("CEXIO_BGD_SELL")] = tFsm.ChanSellBTCEvent
	tFsm.ChanMapForRedisEvents[fmt.Sprintf("CEXIO_ETH_SELL")] = tFsm.ChanSellBTCEvent
	tFsm.ChanMapForRedisEvents[fmt.Sprintf("CEXIO_DASH_SELL")] = tFsm.ChanSellBTCEvent

	tFsm.ChanMapForRedisEvents["TRADE"] = tFsm.ChanTradeEvent
	tFsm.ChanMapForRedisEvents["START"] = tFsm.ChanStartEvent
	tFsm.ChanMapForRedisEvents["STOP"] = tFsm.ChanStartEvent



	return tFsm

}

func (tFsm *CryptoSelectorFsm) SignalChannelsMap() map[string]chan bool {
	return tFsm.ChanMapForRedisEvents
}

func (tFsm *CryptoSelectorFsm) FsmController() {

	logMap := make(map[string]bool)

	for _ , event := range tFsm.eventsStringList {
		logMap[event] = false
	}

		log.Info("Starting tFsm controlloer for : ", tFsm.pairID)

		for {
			select {

			case _ = <-tFsm.ChanStartEvent:
				{
					log.Infof("tFsm %s Controller Event: %s", tFsm.pairID, StartEvent)
					if err := tFsm.FSM.Event(StartEvent); err != nil {
						//log.Warn(err.Error())
					}

				}
			case _ = <-tFsm.ChanShutdownEvent:
				{
					log.Infof("tFsm %s Controller Event: %s", tFsm.pairID, ShutdownEvent)
					if err := tFsm.FSM.Event(ShutdownState); err != nil {
						//log.Warn(err.Error())
					}
				}
			case _ = <-tFsm.ChanTradeEvent:
				{
					log.Infof("tFsm %s Controller Event: %s", tFsm.pairID, TradeEvent)
					if err := tFsm.FSM.Event(TradeEvent); err != nil {
						//log.Warn(err.Error())
					}
				}

			case ev := <-tFsm.ChanMinute120BuyEvent:
				{

					doLog := false
					if logMap[Minute120BuyEvent] != ev {
						doLog = true
					}
					logMap[Minute120BuyEvent] = ev

					if ev {
						if doLog {
							log.Infof("tFsm %s Controller Event: %s , %v", tFsm.pairID, Minute120BuyEvent, ev)
						}
						if err := tFsm.FSM.Event(Minute120BuyEvent); err != nil {
							//log.Warn(err.Error())
						}
					} else {
						if doLog {
							log.Infof("tFsm %s Controller Event: %s , %v", tFsm.pairID, NotMinute120BuyEvent, ev)
						}
						if err := tFsm.FSM.Event(NotMinute120BuyEvent); err != nil {
							//log.Warn(err.Error())
						}
					}

				}
			case ev := <-tFsm.ChanMinute60BuyEvent:
				{

					doLog := false
					if logMap[Minute60BuyEvent] != ev {
						doLog = true
					}
					logMap[Minute60BuyEvent] = ev

					if ev {
						if doLog {
							log.Infof("tFsm %s Controller Event: %s, %v", tFsm.pairID, Minute60BuyEvent, ev)
						}

						if err := tFsm.FSM.Event(Minute60BuyEvent); err != nil {
							//log.Warn(err.Error())
						}
					} else {
						if doLog {
							log.Infof("tFsm %s Controller Event: %s, %v", tFsm.pairID, NotMinute60BuyEvent, ev)
						}

						if err := tFsm.FSM.Event(NotMinute60BuyEvent); err != nil {
							//log.Warn(err.Error())
						}
					}

				}
			case ev := <-tFsm.ChanMinute30BuyEvent:
				{
					doLog := false
					if logMap[Minute30BuyEvent] != ev {
						doLog = true
					}
					logMap[Minute30BuyEvent] = ev

					if ev {
						if doLog {
							log.Infof("tFsm %s Controller Event: %s, %v", tFsm.pairID, Minute30BuyEvent, ev)
						}

						if err := tFsm.FSM.Event(Minute30BuyEvent); err != nil {
							//log.Warn(err.Error())
						}
					} else {
						if doLog {
							log.Infof("tFsm %s Controller Event: %s, %v", tFsm.pairID, NotMinute30BuyEvent, ev)
						}

						if err := tFsm.FSM.Event(NotMinute30BuyEvent); err != nil {
							//log.Warn(err.Error())
						}
					}
				}

			case ev := <-tFsm.ChanMinute120SellEvent:
				{

					doLog := false
					if logMap[Minute120SellEvent] != ev {
						doLog = true
					}
					logMap[Minute120SellEvent] = ev

					if ev {
						if doLog {
							log.Infof("tFsm %s Controller Event: %s, %v", tFsm.pairID, Minute120SellEvent, ev)
						}

						if err := tFsm.FSM.Event(Minute120SellEvent); err != nil {
							//log.Warn(err.Error())
						}
					} else {
						if doLog {
							log.Infof("tFsm %s Controller Event: %s, %v", tFsm.pairID, NotMinute120SellEvent, ev)
						}
						if err := tFsm.FSM.Event(NotMinute120SellEvent); err != nil {
							//log.Warn(err.Error())
						}
					}

				}
			case ev := <-tFsm.ChanMinute60SellEvent:
				{

					doLog := false
					if logMap[Minute60SellEvent] != ev {
						doLog = true
					}
					logMap[Minute60SellEvent] = ev

					if ev {
						if doLog {
							log.Infof("tFsm %s Controller Event: %s, %v", tFsm.pairID, Minute60SellEvent, ev)
						}

						if err := tFsm.FSM.Event(Minute60SellEvent); err != nil {
							//log.Warn(err.Error())
						}

					} else {
						if doLog {
							log.Infof("tFsm %s Controller Event: %s, %v", tFsm.pairID, NotMinute60SellEvent, ev)
						}

						if err := tFsm.FSM.Event(NotMinute60SellEvent); err != nil {
							//log.Warn(err.Error())
						}

					}
				}
			case ev := <-tFsm.ChanMinute30SellEvent:
				{

					doLog := false
					if logMap[Minute30SellEvent] != ev {
						doLog = true
					}
					logMap[Minute30SellEvent] = ev

					if ev {
						if doLog {
							log.Infof("tFsm %s Controller Event: %s, %v", tFsm.pairID, Minute30SellEvent, ev)
						}

						if err := tFsm.FSM.Event(Minute30SellEvent); err != nil {
							//log.Warn(err.Error())
						}
					} else {
						if doLog {
							log.Infof("tFsm %s Controller Event: %s, %v", tFsm.pairID, NotMinute30SellEvent, ev)
						}
						if err := tFsm.FSM.Event(NotMinute30SellEvent); err != nil {
							//log.Warn(err.Error())
						}
					}
				}

			case _ = <-tFsm.ChanDoBuyEvent:
				{
					log.Infof("tFsm %s Controller Event: %s, %v", tFsm.pairID, DoBuyEvent)
					if err := tFsm.FSM.Event(DoBuyEvent); err != nil {
						//log.Warn(err.Error())
					}
				}
			case _ = <-tFsm.ChanDoSellEvent:
				{
					log.Infof("tFsm %s Controller Event: %s", tFsm.pairID, DoSellEvent)
					if err := tFsm.FSM.Event(DoSellEvent); err != nil {
						//log.Warn(err.Error())
					}
				}
			case _ = <-tFsm.ChanBuyCompleteEvent:
				{
					log.Infof("tFsm %s Controller Event: %s", tFsm.pairID, BuyCompleteEvent)
					if err := tFsm.FSM.Event(BuyCompleteEvent); err != nil {
						//log.Warn(err.Error())
					}
				}
			case _ = <-tFsm.ChanSellCompleteEvent:
				{
					log.Infof("tFsm %s Controller Event: %s", tFsm.pairID, SellCompleteEvent)
					if err := tFsm.FSM.Event(SellCompleteEvent); err != nil {
						//log.Warn(err.Error())
					}
				}

			}
			//time.Sleep(time.Second)
		}
	*/
}
