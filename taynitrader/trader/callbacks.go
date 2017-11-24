package trader

import (
	"github.com/looplab/fsm"
	log "github.com/sirupsen/logrus"
)

func (tf *TradeFsm) CallBackInStartState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInIdleState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInHoldState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInDoSellState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInDoBuyState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInTradingState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInShutdownState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInMinute120BuyState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInMinute60BuyState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInMinute30BuyState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInMinute120SellState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInMinute60SellState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInMinute30SellState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

// -----------
// Not Events
// -----------

func (tf *TradeFsm) CallBackInNotMinute120BuyState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInNotMinute60BuyState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInNotMinute30BuyState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInNotMinute120SellState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInNotMinute60SellState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInNotMinute30SellState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}
