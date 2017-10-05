package taynibot

import "sync"

type balance struct {
	usd float64
	btc float64
	btcUsdBase float64
	mu *sync.Mutex
}

func NewBalance() (bal *balance) {
	bal = &balance{}
	bal.mu = &sync.Mutex{}
	return bal
}

func (bal *balance) SetUSD(usd float64) {
	bal.mu.Lock()
	bal.usd = usd
	bal.mu.Unlock()
}

func (bal *balance) SetBTC(btc, price float64) {
	bal.mu.Lock()
	bal.btc = btc
	bal.btcUsdBase =  price
	bal.mu.Unlock()
}

func (bal *balance) USD() float64 {
	return bal.usd
}

func (bal *balance) BTC() float64 {
	return bal.btc
}

func (bal *balance) BtcUSDBase() float64 {
	return bal.btcUsdBase
}

