package trader_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/lagarciag/tayni/taynitrader/trader"
)

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	fmt.Println("SEED:", seed)
	os.Exit(m.Run())
}

func TestTraderBasic(t *testing.T) {

	tFsm := trader.NewTradeFsm()

	tFsm.FSM.Event(trader.StartEvent)

	if tFsm.FSM.Current() != trader.IdleState {
		t.Error("Bad current state: ", tFsm.FSM.Current())
	} else {
		t.Log("State : ", tFsm.FSM.Current())
	}

	tFsm.FSM.Event(trader.TradingActiveEvent)

	if tFsm.FSM.Current() != trader.WatchState {
		t.Error("Bad current state: ", tFsm.FSM.Current())
	} else {
		t.Log("State : ", tFsm.FSM.Current())
	}

	tFsm.FSM.Event(trader.StartEvent)

	if tFsm.FSM.Current() != trader.WatchState {
		t.Error("Bad current state: ", tFsm.FSM.Current())
	} else {
		t.Log("State : ", tFsm.FSM.Current())
	}

	tFsm.FSM.Event(trader.TradingActiveEvent)

	if tFsm.FSM.Current() != trader.WatchState {
		t.Error("Bad current state: ", tFsm.FSM.Current())
	} else {
		t.Log("State : ", tFsm.FSM.Current())
	}

	tFsm.FSM.Event(trader.ShutdownEvent)

	if tFsm.FSM.Current() != trader.ShutdownState {
		t.Error("Bad current state: ", tFsm.FSM.Current())
	} else {
		t.Log("State : ", tFsm.FSM.Current())
	}

}
