package trader

import (
	"time"

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

	done := func() {
		if err := tf.FSM.Event(BuyCompleteEvent); err != nil {
			log.Warn(err.Error())
		}
	}
	time.Sleep(time.Second)
	go done()

}

func (tf *TradeFsm) CallBackInTradingState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInShutdownState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInMinute1BuyState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current(), tf.pairID)

	done := func() {
		if err := tf.FSM.Event(TestDoBuyEvent); err != nil {
			log.Warn(err.Error())
		}
	}
	log.Info("Test executing buy for ", tf.pairID)
	time.Sleep(time.Second)
	go done()
	log.Infof("CallBack done: %s, %s", tf.FSM.Current(), tf.pairID)

}

func (tf *TradeFsm) CallBackInMinute120BuyState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInMinute60BuyState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInMinute30BuyState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())

	done := func() {
		if err := tf.FSM.Event(DoBuyEvent); err != nil {
			log.Warn(err.Error())
		}
	}
	log.Info("Test executing buy for ", tf.pairID)
	time.Sleep(time.Second)
	go done()
	log.Infof("CallBack done: %s, %s", tf.FSM.Current(), tf.pairID)

}

func (tf *TradeFsm) CallBackInMinute1SellState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())

	done := func() {
		if err := tf.FSM.Event(TestDoSellEvent); err != nil {
			log.Warn(err.Error())
		}
	}
	log.Info("Executing buy for ", tf.pairID)
	time.Sleep(time.Second)
	go done()
	log.Infof("CallBack done: %s, %s", tf.FSM.Current(), tf.pairID)

}

func (tf *TradeFsm) CallBackInMinute120SellState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInMinute60SellState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInMinute30SellState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
	done := func() {
		if err := tf.FSM.Event(DoSellEvent); err != nil {
			log.Warn(err.Error())
		}

	}
	log.Info("Executing buy for ", tf.pairID)
	time.Sleep(time.Second)
	go done()
	log.Infof("CallBack done: %s, %s", tf.FSM.Current(), tf.pairID)

}

// -----------
// Not Events
// -----------

func (tf *TradeFsm) CallBackInNotMinute1BuyState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInNotMinute120BuyState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInNotMinute60BuyState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInNotMinute30BuyState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInNotMinute1SellState(e *fsm.Event) {
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

func (tf *TradeFsm) CallBackInTestHoldState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
	time.Sleep(time.Second)

}

func (tf *TradeFsm) CallBackInTestDoSellState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())

	done := func() {
		if err := tf.FSM.Event(SellCompleteEvent); err != nil {
			log.Warn(err.Error())
		}
	}
	time.Sleep(time.Second)
	go done()

}

func (tf *TradeFsm) CallBackInTestDoBuyState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current(), tf.pairID)

	done := func() {
		if err := tf.FSM.Event(TestBuyCompleteEvent); err != nil {
			log.Warn(err.Error())
		}
	}
	time.Sleep(time.Second)
	go done()

}

func (tf *TradeFsm) CallBackInTestTradingState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInDoTestDoSellState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}
func (tf *TradeFsm) CallBackInBuyCompleteState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())

}
func (tf *TradeFsm) CallBackInTestBuyCompleteState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}
func (tf *TradeFsm) CallBackInSellCompleteState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}
func (tf *TradeFsm) CallBackInTestSellCompleteState(e *fsm.Event) {
	log.Info("In state :", tf.FSM.Current())
}
