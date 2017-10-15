package taynibot

type Automata interface {
	Start()
	Stop()
	PublicStart()
	PublicRestart()
	UpdatePriceLists(exchange, pair string)
	MonitorPrice()
}
