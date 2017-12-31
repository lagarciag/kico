package cexio

type subscriberType interface{}

type requestAuthAction struct {
	E    string          `json:"e"`
	Auth requestAuthData `json:"auth"`
}

type requestAuthData struct {
	Key       string `json:"key"`
	Signature string `json:"signature"`
	Timestamp int64  `json:"timestamp"`
}

type responseAction struct {
	Action string `json:"e"`
}

type responseAuth struct {
	E    string           `json:"e"`
	Data responseAuthData `json:"data"`
	OK   string           `json:"ok"`
}

type responseAuthData struct {
	Error string `json:"error"`
	OK    string `json:"ok"`
}

type requestPong struct {
	E string `json:"e"`
}

type requestTicker struct {
	E    string   `json:"e"`
	Data []string `json:"data"`
	Oid  string   `json:"oid"`
}

type requestTickerSub struct {
	E     string   `json:"e"`
	Rooms []string `json:"rooms"`
}

type requestInitOhlcvNew struct {
	E     string   `json:"e"`
	I     string   `json:"i"`
	Rooms []string `json:"rooms"`
}

type requestGetBalance struct {
	E    string `json:"e"`
	Data string `json:"data"`
	Oid  string `json:"oid"`
}

type ResponseTicker struct {
	E    string             `json:"e"`
	Data ResponseTickerData `json:"data"`
	OK   string             `json:"ok"`
	Oid  string             `json:"oid"`
}

type ResponseTickerSub struct {
	E    string                `json:"e"`
	Data ResponseTickerSubData `json:"data"`
	OK   string                `json:"ok"`
	Oid  string                `json:"oid"`
}

type ResponseTickerSubData struct {
	Symbol1 string `json:"symbol1"`
	Symbol2 string `json:"symbol2"`
	Price   string `json:"price"`
}

type ResponseTickerData struct {
	Bid   float64  `json:"bid"`
	Ask   float64  `json:"ask"`
	Pair  []string `json:"pair"`
	Error string   `json:"error"`
}

type requestOrderBookSubscribe struct {
	E    string                        `json:"e"`
	Data requestOrderBookSubscribeData `json:"data"`
	Oid  string                        `json:"oid"`
}

type requestOrderBookSubscribeData struct {
	Pair      []string `json:"pair"`
	Subscribe bool     `json:"subscribe"`
	Depth     int64    `json:"depth"`
}

type responseOrderBookSubscribe struct {
	E    string                         `json:"e"`
	Data responseOrderBookSubscribeData `json:"data"`
	OK   string                         `json:"ok"`
	Oid  string                         `json:"oid"`
}

type responseGetBalance struct {
	E    string      `json:"e"`
	Data BalanceData `json:"data"`
	Time int64       `json:"time"`
	Oid  string      `json:"oid"`
	OK   string      `json:"ok"`
}

type BalanceData struct {
	Balance  map[string]string `json:"balance"`
	OBalance map[string]string `json:"o_balance"`
}

type balanceData struct {
	Balance  BalanceS  `json:"balance"`
	Obalance ObalanceS `json:"obalance"`
}

type BalanceS struct {
	LTC string `json:"LTC"`
	USD string `json:"USD"`
	RUB string `json:"RUB"`
	EUR string `json:"EUR"`
	GHS string `json:"GHS"`
	BTC string `json:"BTC"`
}

type ObalanceS struct {
	BTC string `json:"BTC"`
	USD string `json:"USD"`
}

type responseOrderBookSubscribeData struct {
	Timestamp int64       `json:"timestamp"`
	Bids      [][]float64 `json:"bids"`
	Asks      [][]float64 `json:"asks"`
	Pair      string      `json:"pair"`
	ID        int64       `json:"id"`
}

type responseOrderBookUpdate struct {
	E    string                      `json:"e"`
	Data responseOrderBookUpdateData `json:"data"`
}

type responseOrderBookUpdateData struct {
	ID        int64       `json:"id"`
	Pair      string      `json:"pair"`
	Timestamp int64       `json:"time"`
	Bids      [][]float64 `json:"bids"`
	Asks      [][]float64 `json:"asks"`
}

//OrderBookUpdateData data of order book update
type OrderBookUpdateData struct {
	ID        int64
	Pair      string
	Timestamp int64
	Bids      [][]float64
	Asks      [][]float64
}

//SubscriptionHandler subscription update handler type
type SubscriptionHandler func(updateData OrderBookUpdateData)

type orderBookPair struct {
	Pair  []string `json:"pair"`
	Error string   `json:"error,omitempty"`
}

type requestOrderBookUnsubscribe struct {
	E    string        `json:"e"`
	Data orderBookPair `json:"data"`
	Oid  string        `json:"oid"`
}

type responseOrderBookUnsubscribe struct {
	E    string        `json:"e"`
	Data orderBookPair `json:"data"`
	OK   string        `json:"ok"`
	Oid  string        `json:"oid"`
}

type RequestOrderPlacement struct {
	E    string                    `json:"e"`
	Data RequestOrderPlacementData `json:"data"`
	Oid  string                    `json:"oid"`
}

type ResponseOrderPlacement struct {
	E    string                     `json:"e"`
	Data ResponseOrderPlacementData `json:"data"`
	Oid  string                     `json:"oid"`
	OK   string                     `json:"ok"`
}

type RequestOrderPlacementData struct {
	Pair   []string `json:"pair"`
	Amount string   `json:"amount"`
	Price  string   `json:"price"`
	Type   string   `json:"type"`
}

type ResponseOrderPlacementData struct {
	Complete bool    `json:"complete"`
	ID       string  `json:"id"`
	Time     float64 `json:"time"`
	Pending  string  `json:"pending"`
	Amount   string  `json:"amount"`
	Type     string  `json:"type"`
	Price    string  `json:"price"`
	Error    string  `json:"error"`
}

type RequestOpenOrders struct {
	E    string                `json:"e"`
	Data RequestOpenOrdersData `json:"data"`
	Oid  string                `json:"oid"`
}

type RequestOpenOrdersData struct {
	Pair []string `json:"pair"`
}

type ResponseOpenOrders struct {
	E    string                   `json:"e"`
	Data []ResponseOpenOrdersData `json:"data"`
	Oid  string                   `json:"oid"`
	OK   string                   `json:"ok"`
}

type ResponseOpenOrdersData struct {
	ID      string `json:"id"`
	Time    string `json:"time"`
	Pending string `json:"pending"`
	Amount  string `json:"amount"`
	Type    string `json:"type"`
	Price   string `json:"price"`
}

type ResponseOrder struct {
	E    string            `json:"e"`
	Data ResponseOrderData `json:"data"`
}

type ResponseOrderData struct {
	ID       string            `json:"id"`
	Remains  string            `json:"remains"`
	Fremains string            `json:"fremains"`
	Cancel   bool              `json:"cancel"`
	Pair     map[string]string `json:"pair"`
}
