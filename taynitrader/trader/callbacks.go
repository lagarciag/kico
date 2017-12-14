package trader

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lagarciag/movingstats"
	"github.com/looplab/fsm"
	log "github.com/sirupsen/logrus"
)

func (tf *TradeFsm) CallBackInStartState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
}

func (tf *TradeFsm) CallBackInIdleState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
}

func (tf *TradeFsm) CallBackInHoldState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)

}

func (tf *TradeFsm) CallBackInDoSellState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())
	done := func() {
		if err := tf.FSM.Event(SellCompleteEvent); err != nil {
			log.Warn(err.Error())
		}
	}

	if tf.pairID != "TEST" {

		t := time.Now().UTC()
		theTime := fmt.Sprint(t.String())

		indicators := tf.indicatorsGetter(0)

		ema := indicators.Ema
		sma := indicators.Sma

		twitMessage := `
		TayniBot (beta tests) says: SELL %s
		ema : %f
		sma : %f
		time: %s
		`

		twit := fmt.Sprintf(twitMessage, tf.pairID, ema, sma, theTime)

		if err := tf.tc.Twit(twit); err != nil {
			log.Error(err.Error())
			log.Info(twit)
		}

		sellKey := fmt.Sprintf("%s_SELL", tf.pairID)
		if err := tf.kr.Publish(sellKey, "true"); err != nil {
			log.Errorf("Publishing to: %s -> %s ", sellKey, "true")
		}

	}

	message := `
	----------------------------------------------------
	SELL COMPLETE for PAIR: %s
	----------------------------------------------------
	`
	log.Infof(message, tf.pairID)
	time.Sleep(time.Millisecond * 100)
	go done()
}

func (tf *TradeFsm) CallBackInDoBuyState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())

	done := func() {
		if err := tf.FSM.Event(BuyCompleteEvent); err != nil {
			log.Warn(err.Error())
		}
	}

	//if tf.pairID != "TEST" {

	t := time.Now().UTC()

	theTime := fmt.Sprint(t.String())

	indicators := tf.indicatorsGetter(0)

	ema := indicators.Ema
	sma := indicators.Sma

	twitMessage := `
		TayniBot (beta tests) says: BUY %s
		ema : %f
		sma : %f
		time: %s
		`

	twit := fmt.Sprintf(twitMessage, tf.pairID, ema, sma, theTime)

	if err := tf.tc.Twit(twit); err != nil {
		log.Error(err.Error())
		log.Info(twit)
	}

	buyKey := fmt.Sprintf("%s_BUY", tf.pairID)
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
	log.Infof(message, tf.pairID)
	time.Sleep(time.Millisecond * 100)
	go done()

}

func (tf *TradeFsm) CallBackInTradingState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInShutdownState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInMinute1BuyState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current(), tf.pairID)

	done := func() {
		if err := tf.FSM.Event(TestDoBuyEvent); err != nil {
			log.Warn(err.Error())
		}
	}
	log.Info("Test executing buy for ", tf.pairID)

	time.Sleep(time.Millisecond * 100)
	go done()
	log.Infof("CallBack done: %s, %s", tf.FSM.Current(), tf.pairID)

}

func (tf *TradeFsm) CallBackInMinute120BuyState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInMinute60BuyState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInMinute30BuyState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())

	done := func() {
		if err := tf.FSM.Event(DoBuyEvent); err != nil {
			log.Warn(err.Error())
		}
	}
	log.Info("Test executing buy for ", tf.pairID)
	time.Sleep(time.Millisecond * 100)
	go done()
	log.Infof("CallBack done: %s, %s", tf.FSM.Current(), tf.pairID)

}

func (tf *TradeFsm) CallBackInMinute1SellState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())

	done := func() {
		if err := tf.FSM.Event(TestDoSellEvent); err != nil {
			log.Warn(err.Error())
		}
	}
	log.Info("Executing buy for ", tf.pairID)
	time.Sleep(time.Millisecond * 100)
	go done()
	log.Infof("CallBack done: %s, %s", tf.FSM.Current(), tf.pairID)

}

func (tf *TradeFsm) CallBackInMinute120SellState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInMinute60SellState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInMinute30SellState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())
	done := func() {
		if err := tf.FSM.Event(DoSellEvent); err != nil {
			log.Warn(err.Error())
		}

	}
	log.Info("Executing buy for ", tf.pairID)
	time.Sleep(time.Millisecond * 100)
	go done()
	log.Infof("CallBack done: %s, %s", tf.FSM.Current(), tf.pairID)

}

// -----------
// Not Events
// -----------

func (tf *TradeFsm) CallBackInNotMinute1BuyState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInNotMinute120BuyState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInNotMinute60BuyState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInNotMinute30BuyState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInNotMinute1SellState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInNotMinute120SellState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInNotMinute60SellState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInNotMinute30SellState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInTestHoldState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())
	time.Sleep(time.Millisecond * 100)

}

func (tf *TradeFsm) CallBackInTestDoSellState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())

	done := func() {
		if err := tf.FSM.Event(SellCompleteEvent); err != nil {
			log.Warn(err.Error())
		}
	}
	time.Sleep(time.Millisecond * 100)
	go done()

}

func (tf *TradeFsm) CallBackInTestDoBuyState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current(), tf.pairID)

	done := func() {
		if err := tf.FSM.Event(TestBuyCompleteEvent); err != nil {
			log.Warn(err.Error())
		}
	}
	time.Sleep(time.Millisecond * 100)
	go done()

}

func (tf *TradeFsm) CallBackInTestTradingState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())
}

func (tf *TradeFsm) CallBackInDoTestDoSellState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())
}
func (tf *TradeFsm) CallBackInBuyCompleteState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())

}
func (tf *TradeFsm) CallBackInTestBuyCompleteState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())
}
func (tf *TradeFsm) CallBackInSellCompleteState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())
}
func (tf *TradeFsm) CallBackInTestSellCompleteState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)
	//log.Info("In state :", tf.FSM.Current())
}

//TODO: This does not go here
func (tf *TradeFsm) indicatorsGetter(index int) (indicators movingstats.Indicators) {

	key := fmt.Sprintf("CEXIO_%s_MS_30_INDICATORS", tf.pairID)
	indicatorsJson, err := tf.kr.GetRawString(key, index)

	if err != nil {
		log.Fatal("Fatal error getting indicators: ", err.Error())
	}

	if err = json.Unmarshal([]byte(indicatorsJson), &indicators); err != nil {
		log.Error("unmarshaling indicators json")
	}

	return indicators

}
