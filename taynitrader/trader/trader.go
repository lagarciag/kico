package trader

import (
	"github.com/looplab/fsm"
	log "github.com/sirupsen/logrus"
)

const (
	StartState    = "StartState"
	IdleState     = "IdleState"
	WatchState    = "WatchState"
	BuyState      = "BuyState"
	SellState     = "SellState"
	ShutdownState = "ShutdownState"
)

const (
	StartEvent         = "startEvent"
	TradingActiveEvent = "tradingActiveEvent"
	BuyEvent           = "buyEvent"
	SellEvent          = "sellEvent"
	ShutdownEvent      = "shutdownEvent"
)

type TradeFsm struct {
	To  string
	FSM *fsm.FSM

	// ------------
	// Events
	// ------------
	startEvent         fsm.EventDesc
	tradingActiveEvent fsm.EventDesc
	buyEvent           fsm.EventDesc
	sellEvent          fsm.EventDesc
	shutdownEvent      fsm.EventDesc

	// ------------------
	// Events List
	// ------------------
	eventsList fsm.Events

	// ------------------
	// Fsm Callbacks
	callbacks fsm.Callbacks
}

func NewTradeFsm() *TradeFsm {

	tFsm := &TradeFsm{}

	// ------------
	// Events
	// ------------
	tFsm.startEvent = fsm.EventDesc{Name: StartEvent, Src: []string{StartState}, Dst: IdleState}

	tFsm.tradingActiveEvent = fsm.EventDesc{Name: TradingActiveEvent, Src: []string{IdleState}, Dst: WatchState}

	tFsm.buyEvent = fsm.EventDesc{Name: BuyEvent, Src: []string{WatchState}, Dst: BuyState}

	tFsm.sellEvent = fsm.EventDesc{Name: SellEvent, Src: []string{WatchState}, Dst: SellState}

	tFsm.shutdownEvent = fsm.EventDesc{Name: ShutdownEvent,
		Src: []string{StartState,
			WatchState,
			IdleState},
		Dst: ShutdownState}
	// ----------------------
	// Events List Registry
	// ----------------------
	tFsm.eventsList = fsm.Events{tFsm.startEvent, tFsm.tradingActiveEvent, tFsm.shutdownEvent}

	// -------------------
	// Callbacks registry
	// -------------------
	tFsm.callbacks = fsm.Callbacks{
		StartState:    tFsm.CallBackInStartState,
		IdleState:     tFsm.CallBackInIdleState,
		WatchState:    tFsm.CallBackInWatchState,
		ShutdownState: tFsm.CallBackInShutdownState}

	// -------------
	// FSM creation
	// -------------
	tFsm.FSM = fsm.NewFSM(StartState,
		tFsm.eventsList,
		tFsm.callbacks)

	return tFsm

}

func (tf *TradeFsm) CallBackInStartState(e *fsm.Event) {
	log.Info("In start state...")
}

func (tf *TradeFsm) CallBackInIdleState(e *fsm.Event) {
	log.Info("In idle state...")
}

func (tf *TradeFsm) CallBackInWatchState(e *fsm.Event) {
	log.Info("In watch state...")
}

func (tf *TradeFsm) CallBackInShutdownState(e *fsm.Event) {
	log.Info("In shutdown state...")
}

func Start() {
	log.Info("Tayni Trader starting...")
}
