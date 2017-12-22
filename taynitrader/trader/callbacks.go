package trader

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lagarciag/movingstats"
	"github.com/looplab/fsm"
	log "github.com/sirupsen/logrus"
)

func (tf *TradeFsm) CallBackInGenericState(e *fsm.Event) {
	log.Infof("In state %s --> %s:", tf.FSM.Current(), tf.pairID)

	key := fmt.Sprintf("%s_TRADE_FSM_STATE", tf.pairID)
	tf.kr.Set(key, tf.FSM.Current())

	switch {

	case tf.FSM.Current() == Minute30BuyState:
		{

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

	case tf.FSM.Current() == Minute30SellState:
		{
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

	default:
		log.Infof("No action state")

	}

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
		last := indicators.LastValue
		atrp := indicators.ATRP
		pDmi := indicators.PDI
		mDmi := indicators.MDI

		var twitMessage string

		switch {

		case tf.pairID == "ETHBTC":
			{

				twitMessage = `
		TayniBot (beta tests) says: ETH is downperforming BTC.
		Watch for BTC buy signal
		ema : %f
		sma : %f
		last: %f
		ATRP: %f
		PDMI: %f
		MDMI: %f
		time: %s
		`
			}

		case tf.pairID == "BCHBTC":
			{

				twitMessage = `
		TayniBot (beta tests) says: BCH is downperforming BTC.
		Watch for BTC buy signal
		ema : %f
		sma : %f
		last: %f
		ATRP: %f
		PDMI: %f
		MDMI: %f
		time: %s
		`
			}
		case tf.pairID == "DASHBTC":
			{

				twitMessage = `
		TayniBot (beta tests) says: DASH is downperforming BTC.
		Watch for BTC buy signal
		ema : %f
		sma : %f
		last: %f
		ATRP: %f
		PDMI: %f
		MDMI: %f
		time: %s
		`
			}
		case tf.pairID == "ZECBTC":
			{

				twitMessage = `
		TayniBot (beta tests) says: ZEC is downperforming BTC.
		Watch for BTC buy signal
		ema : %f
		sma : %f
		last: %f
		ATRP: %f
		PDMI: %f
		MDMI: %f
		time: %s
		`
			}

		case tf.pairID == "XRPBTC":
			{

				twitMessage = `
		TayniBot (beta tests) says: XRP is downperforming BTC.
		Watch for BTC buy signal
		ema : %f
		sma : %f
		last: %f
		ATRP: %f
		PDMI: %f
		MDMI: %f
		time: %s
		`
			}

		default:
			{
				twitMessage = `
		TayniBot (beta tests) says: BUY %s
		ema : %f
		sma : %f
		last: %f
		ATRP: %f
		PDMI: %f
		MDMI: %f
		time: %s
		`
			}

		}

		twit := fmt.Sprintf(twitMessage, tf.pairID, ema, sma, last, atrp, pDmi, mDmi, theTime)

		if err := tf.tc.Twit(twit); err != nil {
			log.Error(err.Error())
			log.Info(twit)
		}

		sellKey := fmt.Sprintf("CEXIO_%s_SELL", tf.pairID)
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
	last := indicators.LastValue
	atrp := indicators.ATRP
	pDmi := indicators.PDI
	mDmi := indicators.MDI

	var twitMessage string

	switch {

	case tf.pairID == "ETHBTC":
		{

			twitMessage = `
		TayniBot (beta tests) says: ETH is outperforming BTC.
		Watch for ETH buy signal
		ema : %f
		sma : %f
		last: %f
		ATRP: %f
		PDMI: %f
		MDMI: %f
		time: %s
		`
		}

	case tf.pairID == "BCHBTC":
		{

			twitMessage = `
		TayniBot (beta tests) says: BCH is outperforming BTC.
		Watch for BCH buy signal
		ema : %f
		sma : %f
		last: %f
		ATRP: %f
		PDMI: %f
		MDMI: %f
		time: %s
		`
		}
	case tf.pairID == "DASHBTC":
		{

			twitMessage = `
		TayniBot (beta tests) says: DASH is outperforming BTC.
		Watch for DASH buy signal
		ema : %f
		sma : %f
		last: %f
		ATRP: %f
		PDMI: %f
		MDMI: %f
		time: %s
		`
		}
	case tf.pairID == "ZECBTC":
		{

			twitMessage = `
		TayniBot (beta tests) says: ZEC is outperforming BTC.
		Watch for ZEC buy signal
		ema : %f
		sma : %f
		last: %f
		ATRP: %f
		PDMI: %f
		MDMI: %f
		time: %s
		`
		}

	case tf.pairID == "XRPBTC":
		{

			twitMessage = `
		TayniBot (beta tests) says: XRP is outperforming BTC.
		Watch for XRP buy signal
		ema : %f
		sma : %f
		last: %f
		ATRP: %f
		PDMI: %f
		MDMI: %f
		time: %s
		`
		}

	default:
		{
			twitMessage = `
		TayniBot (beta tests) says: BUY %s
		ema : %f
		sma : %f
		last: %f
		ATRP: %f
		PDMI: %f
		MDMI: %f
		time: %s
		`
		}

	}

	twit := fmt.Sprintf(twitMessage, tf.pairID, ema, sma, last, atrp, pDmi, mDmi, theTime)

	if err := tf.tc.Twit(twit); err != nil {
		log.Error(err.Error())
		log.Info(twit)
	}

	buyKey := fmt.Sprintf("CEXIO_%s_BUY", tf.pairID)
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

func (tf *TradeFsm) CallBackInBuyCompleteState(e *fsm.Event) {
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

	key := fmt.Sprintf("CEXIO_%s_MS_120_INDICATORS", tf.pairID)
	indicatorsJson, err := tf.kr.GetRawString(key, index)

	if err != nil {
		log.Fatal("Fatal error getting indicators: ", err.Error())
	}

	if err = json.Unmarshal([]byte(indicatorsJson), &indicators); err != nil {
		log.Error("unmarshaling indicators json")
	}

	return indicators

}
