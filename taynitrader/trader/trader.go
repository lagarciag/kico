package trader

import (
	"github.com/looplab/fsm"
	log "github.com/sirupsen/logrus"
)

const (
	StartState    = "StartState"
	IdleState     = "IdleState"
	TradingState  = "TradingState"
	HoldState     = "HoldState"
	ShutdownState = "ShutdownState"

	// --------------
	// Buy States
	// --------------

	Minute120BuyState = "Minute120BuyState"
	Minute60BuyState  = "Minute60BuyState"
	Minute30BuyState  = "Minute30BuySate"

	DoBuyState = "DoBuyState"

	// -------------
	//   Sell Sates
	// -------------

	Minute120SellState = "Minute120SellState"
	Minute60SellState  = "Minute60SellState"
	Minute30SellState  = "Minute30SellState"

	DoSellState = "DoSellState"
)

const (
	StartEvent    = "startEvent"
	ShutdownEvent = "shutdownEvent"
	TradeEvent    = "TradeEvent"

	// --------------
	// Buy Events
	// --------------

	Minute120BuyEvent = "Minute120BuyEvent"
	Minute60BuyEvent  = "Minute60BuyEvent"
	Minute30BuyEvent  = "Minute30BuySate"

	NotMinute120BuyEvent = "NotMinute120BuyEvent"
	NotMinute60BuyEvent  = "NotMinute60BuyEvent"
	NotMinute30BuyEvent  = "NotMinute30BuySate"

	DoBuyEvent        = "DoBuyEvent"
	DoSellEvent       = "DoSellEvent"
	BuyCompleteEvent  = "BuyCompleteEvent"
	SellCompleteEvent = "SellCompleteEvent"

	// -------------
	//   Sell Sates
	// -------------

	Minute120SellEvent = "Minute120SellEvent"
	Minute60SellEvent  = "Minute60SellEvent"
	Minute30SellEvent  = "Minute30SellEvent"

	NotMinute120SellEvent = "NotMinute120SellEvent"
	NotMinute60SellEvent  = "NotMinute60SellEvent"
	NotMinute30SellEvent  = "NotMinute30SellEvent"
)

type TradeFsm struct {
	To  string
	FSM *fsm.FSM

	// ------------
	// Events
	// ------------
	startEvent    fsm.EventDesc
	shutdownEvent fsm.EventDesc
	tradeEvent    fsm.EventDesc

	minute120BuyEvent fsm.EventDesc
	minute60BuyEvent  fsm.EventDesc
	minute30BuyEvent  fsm.EventDesc

	notMinute120BuyEvent fsm.EventDesc
	notMinute60BuyEvent  fsm.EventDesc
	notMinute30BuyEvent  fsm.EventDesc

	minute120SellEvent fsm.EventDesc
	minute60SellEvent  fsm.EventDesc
	minute30SellEvent  fsm.EventDesc

	notMinute120SellEvent fsm.EventDesc
	notMinute60SellEvent  fsm.EventDesc
	notMinute30SellEvent  fsm.EventDesc

	doBuyEvent        fsm.EventDesc
	doSellEvent       fsm.EventDesc
	buyCompleteEvent  fsm.EventDesc
	sellCompleteEvent fsm.EventDesc

	// ------------------
	// Events List
	// ------------------
	eventsList fsm.Events

	// ------------------
	// Fsm Callbacks
	callbacks fsm.Callbacks

	StartEvent    chan bool
	ShutdownEvent chan bool
	TradeEvent    chan bool

	// --------------
	// Buy Events
	// --------------

	Minute120BuyEvent chan bool
	Minute60BuyEvent  chan bool
	Minute30BuyEvent  chan bool

	NotMinute120BuyEvent chan bool
	NotMinute60BuyEvent  chan bool
	NotMinute30BuyEvent  chan bool

	DoBuyEvent        chan bool
	DoSellEvent       chan bool
	BuyCompleteEvent  chan bool
	SellCompleteEvent chan bool
}

func NewTradeFsm() *TradeFsm {
	log.Info("Creating new trading fsm...")

	tFsm := &TradeFsm{}

	// ------------
	// Events
	// ------------
	startEvent := fsm.EventDesc{Name: StartEvent, Src: []string{StartState}, Dst: IdleState}
	tradeEvent := fsm.EventDesc{Name: TradeEvent, Src: []string{IdleState}, Dst: TradingState}

	// ----------------------
	// Buying related events
	// ----------------------

	minute120BuyEvent := fsm.EventDesc{Name: Minute120BuyEvent,
		Src: []string{TradingState},
		Dst: Minute120BuyState}

	minute60BuyEvent := fsm.EventDesc{Name: Minute60BuyEvent,
		Src: []string{Minute120BuyState},
		Dst: Minute60BuyState}

	minute30BuyEvent := fsm.EventDesc{Name: Minute30BuyEvent,
		Src: []string{Minute60BuyState},
		Dst: Minute30BuyState}

	notMinute120BuyEvent := fsm.EventDesc{Name: NotMinute120BuyEvent,
		Src: []string{Minute120BuyState, Minute60BuyState, Minute30BuyState},
		Dst: TradingState}

	notMinute60BuyEvent := fsm.EventDesc{Name: NotMinute60BuyEvent,
		Src: []string{Minute60BuyState, Minute30BuyState},
		Dst: Minute120BuyState}

	notMinute30BuyEvent := fsm.EventDesc{Name: NotMinute30BuyEvent,
		Src: []string{Minute30BuyState},
		Dst: Minute60BuyState}

	// ----------------------
	// Selling related events
	// ----------------------

	minute120SellEvent := fsm.EventDesc{Name: Minute120SellEvent,
		Src: []string{HoldState},
		Dst: TradingState}

	minute60SellEvent := fsm.EventDesc{Name: Minute60SellEvent,
		Src: []string{Minute60BuyState},
		Dst: Minute120BuyState}

	minute30SellEvent := fsm.EventDesc{Name: Minute30SellEvent,
		Src: []string{Minute30BuyState},
		Dst: Minute60BuyState}

	notMinute120SellEvent := fsm.EventDesc{Name: NotMinute120SellEvent,
		Src: []string{Minute120SellState, Minute60SellState, Minute30SellState},
		Dst: HoldState}

	notMinute60SellEvent := fsm.EventDesc{Name: NotMinute60SellEvent,
		Src: []string{Minute60SellState, Minute30SellEvent},
		Dst: Minute120SellState}

	notMinute30SellEvent := fsm.EventDesc{Name: NotMinute30SellEvent,
		Src: []string{Minute30SellState},
		Dst: Minute60SellState}

	doBuyEvent := fsm.EventDesc{Name: DoBuyEvent,
		Src: []string{Minute30BuyState},
		Dst: DoBuyState}

	doSellEvent := fsm.EventDesc{Name: DoSellEvent,
		Src: []string{Minute30SellState},
		Dst: DoBuyState}

	buyCompleteEvent := fsm.EventDesc{Name: BuyCompleteEvent,
		Src: []string{DoBuyState},
		Dst: HoldState}

	sellCompleteEvent := fsm.EventDesc{Name: SellCompleteEvent,
		Src: []string{DoSellState},
		Dst: TradingState}

	tFsm.startEvent = startEvent
	tFsm.tradeEvent = tradeEvent

	tFsm.minute120BuyEvent = minute120BuyEvent
	tFsm.minute60BuyEvent = minute60BuyEvent
	tFsm.minute120SellEvent = minute120SellEvent
	tFsm.notMinute120BuyEvent = notMinute120BuyEvent
	tFsm.notMinute120SellEvent = notMinute120SellEvent

	tFsm.notMinute120BuyEvent = notMinute120BuyEvent
	tFsm.minute60SellEvent = minute60SellEvent
	tFsm.notMinute60BuyEvent = notMinute60BuyEvent
	tFsm.minute30BuyEvent = minute30BuyEvent
	tFsm.notMinute30SellEvent = notMinute30SellEvent

	tFsm.notMinute60BuyEvent = notMinute60BuyEvent
	tFsm.minute30SellEvent = minute30SellEvent
	tFsm.notMinute30BuyEvent = notMinute30BuyEvent
	tFsm.minute60BuyEvent = minute60BuyEvent
	tFsm.notMinute60SellEvent = notMinute60SellEvent

	tFsm.doBuyEvent = doBuyEvent
	tFsm.doSellEvent = doSellEvent

	tFsm.buyCompleteEvent = buyCompleteEvent
	tFsm.sellCompleteEvent = sellCompleteEvent

	//minute60BuyEvent  fsm.EventDesc
	//minute30BuyEvent  fsm.EventDesc

	//notMinute120BuyEvent fsm.EventDesc
	//notMinute60BuyEvent  fsm.EventDesc
	//notMinute300BuyEvent fsm.EventDesc

	//minute120EventEvent fsm.EventDesc
	//minute60EventEvent  fsm.EventDesc
	//minute30EventEvent  fsm.EventDesc

	//notMinute120EventEvent fsm.EventDesc
	//notMinute60EventEvent  fsm.EventDesc
	//notMinute300EventEvent fsm.EventDesc

	tFsm.shutdownEvent = fsm.EventDesc{Name: ShutdownEvent,
		Src: []string{StartState,
			IdleState},
		Dst: ShutdownState}

	// ----------------------
	// Events List Registry
	// ----------------------
	tFsm.eventsList = fsm.Events{
		tFsm.startEvent,
		tFsm.startEvent,
		tFsm.tradeEvent,

		tFsm.minute120BuyEvent,
		tFsm.minute60BuyEvent,
		tFsm.minute120SellEvent,
		tFsm.notMinute120BuyEvent,
		tFsm.notMinute120SellEvent,

		tFsm.notMinute120BuyEvent,
		tFsm.minute60SellEvent,
		tFsm.notMinute60BuyEvent,
		tFsm.minute30BuyEvent,
		tFsm.notMinute30SellEvent,

		tFsm.notMinute60BuyEvent,
		tFsm.minute30SellEvent,
		tFsm.notMinute30BuyEvent,
		tFsm.minute60BuyEvent,
		tFsm.notMinute60SellEvent,

		tFsm.doBuyEvent,
		tFsm.doSellEvent,

		tFsm.buyCompleteEvent,
		tFsm.sellCompleteEvent,
	}

	// -------------------
	// Callbacks registry
	// -------------------
	tFsm.callbacks = fsm.Callbacks{
		StartState:    tFsm.CallBackInStartState,
		IdleState:     tFsm.CallBackInIdleState,
		ShutdownState: tFsm.CallBackInShutdownState,
		TradingState:  tFsm.CallBackInTradingState,

		Minute120BuyEvent: tFsm.CallBackInMinute120BuyState,
		Minute60BuyEvent:  tFsm.CallBackInMinute60BuyState,
		Minute30BuyEvent:  tFsm.CallBackInMinute30BuyState,

		Minute120SellEvent: tFsm.CallBackInMinute120SellState,
		Minute60SellEvent:  tFsm.CallBackInMinute60SellState,
		Minute30SellEvent:  tFsm.CallBackInMinute30SellState,

		NotMinute120BuyEvent: tFsm.CallBackInNotMinute120BuyState,
		NotMinute60BuyEvent:  tFsm.CallBackInNotMinute60BuyState,
		NotMinute30BuyEvent:  tFsm.CallBackInNotMinute30BuyState,

		NotMinute120SellEvent: tFsm.CallBackInNotMinute120SellState,
		NotMinute60SellEvent:  tFsm.CallBackInNotMinute60SellState,
		NotMinute30SellEvent:  tFsm.CallBackInNotMinute30SellState}

	// ------------------
	// Event Channels
	// ------------------

	tFsm.StartEvent = make(chan bool)
	tFsm.StartEvent    = make(chan bool)
		ShutdownEvent = make(chan bool)
		TradeEvent    = make(chan bool)

		// --------------
		// Buy Events
		// --------------

		Minute120BuyEvent = make(chan bool)
		Minute60BuyEvent  = make(chan bool)
		Minute30BuyEvent  = "Minute30BuySate"

		NotMinute120BuyEvent = "NotMinute120BuyEvent"
		NotMinute60BuyEvent  = "NotMinute60BuyEvent"
		NotMinute30BuyEvent  = "NotMinute30BuySate"

		DoBuyEvent        = "DoBuyEvent"
		DoSellEvent       = "DoSellEvent"
		BuyCompleteEvent  = "BuyCompleteEvent"
		SellCompleteEvent = "SellCompleteEvent"
	*/

	// -------------
	// FSM creation
	// -------------
	tFsm.FSM = fsm.NewFSM(StartState,
		tFsm.eventsList,
		tFsm.callbacks)

	return tFsm

}

/*

func (fsm *TradeFsm) FsmController {

	select :




}
*/
