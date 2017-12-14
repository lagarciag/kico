package buysell

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lagarciag/movingstats"
	"github.com/looplab/fsm"
	log "github.com/sirupsen/logrus"
)

func (tf *CryptoSelectorFsm) CallBackInStartState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
}

func (tf *CryptoSelectorFsm) CallBackInIdleState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
}

func (tf *CryptoSelectorFsm) CallBackInState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
}

func (tf *CryptoSelectorFsm) CallBackInHoldState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")

}

func (tf *CryptoSelectorFsm) CallBackInDoSellState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
	//log.Info("In state :", tf.FSM.Current())
	done := func() {
		if err := tf.FSM.Event(SellCompleteEvent); err != nil {
			log.Warn(err.Error())
		}
	}

	if "void" != "TEST" {

		t := time.Now().UTC()
		theTime := fmt.Sprint(t.String())

		indicators := tf.indicatorsGetter(0)

		ema := indicators.Ema
		sma := indicators.Sma
		last := indicators.LastValue
		atrp := indicators.ATRP

		twitMessage := `
		TayniBot (beta tests) says: SELL %s
		ema : %f
		sma : %f
		last: %f
		ATRP: %f
		time: %s
		`

		twit := fmt.Sprintf(twitMessage, "void", ema, sma, last, atrp, theTime)

		if err := tf.tc.Twit(twit); err != nil {
			log.Error(err.Error())
			log.Info(twit)
		}

		sellKey := fmt.Sprintf("%s_SELL", "void")
		if err := tf.kr.Publish(sellKey, "true"); err != nil {
			log.Errorf("Publishing to: %s -> %s ", sellKey, "true")
		}

	}

	message := `
	----------------------------------------------------
	SELL COMPLETE for PAIR: %s
	----------------------------------------------------
	`
	log.Infof(message, "void")
	time.Sleep(time.Millisecond * 100)
	go done()
}

func (tf *CryptoSelectorFsm) CallBackInDoBuyState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
	//log.Info("In state :", tf.FSM.Current())

	done := func() {
		if err := tf.FSM.Event(BuyCompleteEvent); err != nil {
			log.Warn(err.Error())
		}
	}

	//if "void" != "TEST" {

	t := time.Now().UTC()

	theTime := fmt.Sprint(t.String())

	indicators := tf.indicatorsGetter(0)

	ema := indicators.Ema
	sma := indicators.Sma
	last := indicators.LastValue
	atrp := indicators.ATRP

	twitMessage := `
		TayniBot (beta tests) says: BUY %s
		ema : %f
		sma : %f
	    last: %f
	    atrp: %f
		time: %s

		`
	twit := fmt.Sprintf(twitMessage, "void", ema, sma, last, atrp, theTime)

	if err := tf.tc.Twit(twit); err != nil {
		log.Error(err.Error())
		log.Info(twit)
	}

	buyKey := fmt.Sprintf("%s_BUY", "void")
	if err := tf.kr.Publish(buyKey, "true"); err != nil {
		log.Errorf("Publishing to: %s -> %s ", buyKey, "true")
	}

	//} else {

	//}

	message := `
	----------------------------------------------------
	BUY COMPLETE for PAIR: %s
	----------------------------------------------------
	`
	log.Infof(message, "void")
	time.Sleep(time.Millisecond * 100)
	go done()

}

func (tf *CryptoSelectorFsm) CallBackInTradingState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *CryptoSelectorFsm) CallBackInShutdownState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *CryptoSelectorFsm) CallBackInMinute120BuyState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *CryptoSelectorFsm) CallBackInMinute60BuyState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *CryptoSelectorFsm) CallBackInMinute30BuyState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
	//log.Info("In state :", tf.FSM.Current())

	done := func() {
		if err := tf.FSM.Event(DoBuyEvent); err != nil {
			log.Warn(err.Error())
		}
	}
	log.Info("Test executing buy for ", "void")
	time.Sleep(time.Millisecond * 100)
	go done()
	log.Infof("CallBack done: %s, %s", tf.FSM.Current(), "void")

}

func (tf *CryptoSelectorFsm) CallBackInMinute120SellState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *CryptoSelectorFsm) CallBackInMinute60SellState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *CryptoSelectorFsm) CallBackInMinute30SellState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
	//log.Info("In state :", tf.FSM.Current())
	done := func() {
		if err := tf.FSM.Event(DoSellEvent); err != nil {
			log.Warn(err.Error())
		}

	}
	log.Info("Executing buy for ", "void")
	time.Sleep(time.Millisecond * 100)
	go done()
	log.Infof("CallBack done: %s, %s", tf.FSM.Current(), "void")

}

// -----------
// Not Events
// -----------

func (tf *CryptoSelectorFsm) CallBackInNotMinute1BuyState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *CryptoSelectorFsm) CallBackInNotMinute120BuyState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *CryptoSelectorFsm) CallBackInNotMinute60BuyState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *CryptoSelectorFsm) CallBackInNotMinute30BuyState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *CryptoSelectorFsm) CallBackInNotMinute1SellState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *CryptoSelectorFsm) CallBackInNotMinute120SellState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *CryptoSelectorFsm) CallBackInNotMinute60SellState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *CryptoSelectorFsm) CallBackInNotMinute30SellState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *CryptoSelectorFsm) CallBackInTestHoldState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
	//log.Info("In state :", tf.FSM.Current())
	time.Sleep(time.Millisecond * 100)

}

func (tf *CryptoSelectorFsm) CallBackInTestDoSellState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
	//log.Info("In state :", tf.FSM.Current())

	done := func() {
		if err := tf.FSM.Event(SellCompleteEvent); err != nil {
			log.Warn(err.Error())
		}
	}
	time.Sleep(time.Millisecond * 100)
	go done()

}

func (tf *CryptoSelectorFsm) CallBackInBuyCompleteState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
	//log.Info("In state :", tf.FSM.Current())

}

func (tf *CryptoSelectorFsm) CallBackInSellCompleteState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), "void")
	//log.Info("In state :", tf.FSM.Current())
}

//TODO: This does not go here
func (tf *CryptoSelectorFsm) indicatorsGetter(index int) (indicators movingstats.Indicators) {

	key := fmt.Sprintf("CEXIO_%s_MS_30_INDICATORS", "void")
	indicatorsJson, err := tf.kr.GetRawString(key, index)

	if err != nil {
		log.Fatal("Fatal error getting indicators: ", err.Error())
	}

	if err = json.Unmarshal([]byte(indicatorsJson), &indicators); err != nil {
		log.Error("unmarshaling indicators json")
	}

	return indicators

}
