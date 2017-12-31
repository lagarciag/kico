package taynibot

type Automata interface {
	Stop()
	Start()
	Restart()
	UpdatePriceLists(exchange, pair string)
	MonitorPrice()
}
