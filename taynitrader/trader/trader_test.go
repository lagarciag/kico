package trader_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/lagarciag/tayni/taynitrader/trader"
	log "github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	fmt.Println("SEED:", seed)
	// -----------------------------
	// Setup log format
	// -----------------------------
	formatter := &log.TextFormatter{}
	formatter.FullTimestamp = true
	formatter.ForceColors = true
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(formatter)

	os.Exit(m.Run())
}

func TestTraderController(t *testing.T) {

	tFsm := trader.NewTradeFsm()

	go tFsm.FsmController()

	time.Sleep(time.Second)

	//tFsm.FSM.Event(trader.StartEvent)

	tFsm.ChanStartEvent <- true

	time.Sleep(time.Second)

	if tFsm.FSM.Current() != trader.IdleState {
		t.Error("Bad current state: ", tFsm.FSM.Current())
	} else {
		t.Log("State : ", tFsm.FSM.Current())
	}

	//tFsm.FSM.Event(trader.TradeEvent)
	tFsm.ChanTradeEvent <- true

	time.Sleep(time.Second)

	if tFsm.FSM.Current() != trader.TradingState {
		t.Error("Bad current state: ", tFsm.FSM.Current())
	} else {
		t.Log("State : ", tFsm.FSM.Current())
	}

	//err := tFsm.FSM.Event(trader.Minute120BuyEvent)
	tFsm.ChanMinute120BuyEvent <- true
	time.Sleep(time.Second)

	checkState(t, tFsm, trader.Minute120BuyState)

	//err = tFsm.FSM.Event(trader.NotMinute120BuyEvent)
	tFsm.ChanNotMinute120BuyEvent <- true
	time.Sleep(time.Second)
	//err = tFsm.FSM.Event(trader.Minute60BuyEvent)
	//errorExpected(t, err)
	tFsm.ChanMinute60BuyEvent <- true
	time.Sleep(time.Second)
	//err = tFsm.FSM.Event(trader.Minute30BuyEvent)
	//errorExpected(t, err)
	tFsm.ChanMinute30BuyEvent <- true
	time.Sleep(time.Second)
	checkState(t, tFsm, trader.TradingState)

	/*
		fromTradingTo30MinBuy(t, tFsm)

		err = tFsm.FSM.Event(trader.NotMinute120BuyEvent)
		errorNotExpected(t, err)
		checkState(t, tFsm, trader.TradingState)

		fromTradingTo30MinBuy(t, tFsm)

		err = tFsm.FSM.Event(trader.DoBuyEvent)
		errorNotExpected(t, err)
		checkState(t, tFsm, trader.DoBuyState)
	*/
}

func TestTraderBasic(t *testing.T) {

	tFsm := trader.NewTradeFsm()

	tFsm.FSM.Event(trader.StartEvent)

	if tFsm.FSM.Current() != trader.IdleState {
		t.Error("Bad current state: ", tFsm.FSM.Current())
	} else {
		t.Log("State : ", tFsm.FSM.Current())
	}

	tFsm.FSM.Event(trader.TradeEvent)

	if tFsm.FSM.Current() != trader.TradingState {
		t.Error("Bad current state: ", tFsm.FSM.Current())
	} else {
		t.Log("State : ", tFsm.FSM.Current())
	}

	err := tFsm.FSM.Event(trader.Minute120BuyEvent)
	errorNotExpected(t, err)
	checkState(t, tFsm, trader.Minute120BuyState)

	err = tFsm.FSM.Event(trader.NotMinute120BuyEvent)
	errorNotExpected(t, err)

	err = tFsm.FSM.Event(trader.Minute60BuyEvent)
	errorExpected(t, err)

	err = tFsm.FSM.Event(trader.Minute30BuyEvent)
	errorExpected(t, err)
	checkState(t, tFsm, trader.TradingState)

	fromTradingTo30MinBuy(t, tFsm)

	err = tFsm.FSM.Event(trader.NotMinute120BuyEvent)
	errorNotExpected(t, err)
	checkState(t, tFsm, trader.TradingState)

	fromTradingTo30MinBuy(t, tFsm)

	err = tFsm.FSM.Event(trader.DoBuyEvent)
	errorNotExpected(t, err)
	checkState(t, tFsm, trader.DoBuyState)

}

func fromTradingTo30MinBuy(t *testing.T, tFsm *trader.TradeFsm) {

	err := tFsm.FSM.Event(trader.Minute120BuyEvent)
	errorNotExpected(t, err)
	checkState(t, tFsm, trader.Minute120BuyState)

	err = tFsm.FSM.Event(trader.Minute60BuyEvent)
	errorNotExpected(t, err)
	checkState(t, tFsm, trader.Minute60BuyState)

	err = tFsm.FSM.Event(trader.Minute30BuyEvent)
	errorNotExpected(t, err)
	checkState(t, tFsm, trader.Minute30BuyState)

}

func checkState(t *testing.T, tFsm *trader.TradeFsm, state string) {
	t.Log(state, tFsm.FSM.Current())
	if tFsm.FSM.Current() != state {
		t.Errorf("Bad current state: %s, expected %s", tFsm.FSM.Current(), state)
	} else {
		t.Log("State : ", tFsm.FSM.Current())
	}
}

func errorNotExpected(t *testing.T, err error) {
	if err != nil {
		t.Error(err.Error())
	}
}

func errorExpected(t *testing.T, err error) {
	if err == nil {
		t.Error("not permited transition")
	}

}
