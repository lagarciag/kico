package kredis

import (
	"fmt"

	"strconv"

	"sync"

	"github.com/garyburd/redigo/redis"
	log "github.com/sirupsen/logrus"
)

var ExchangesList = []string{"CEXIO"}
var ExchangePairListMap = map[string][]string{"CEXIO": {"BTCUSD"}}

type Kredis struct {
	conn           redis.Conn
	psc            redis.PubSubConn
	connUpdateList redis.Conn
	server         string
	validKeys      map[string]bool
	size           int
	count          uint
	mu             *sync.Mutex
	subsChan       chan []string
}

func NewKredis(size int) *Kredis {
	kr := &Kredis{}
	kr.size = size
	kr.server = ":6379"
	kr.mu = &sync.Mutex{}
	kr.validKeys = make(map[string]bool)
	kr.validKeys["CEXIO_BTCUSD"] = true
	kr.subsChan = make(chan []string, 1000)

	return kr

}

func (kr *Kredis) dial() {
	log.Info("Dialing redis...")
	var err error
	kr.conn, err = redis.Dial("tcp", kr.server)
	if err != nil {
		log.Fatal("Could not dial redis: ", err.Error())
	}

	kr.connUpdateList, err = redis.Dial("tcp", kr.server)
	if err != nil {
		log.Fatal("Could not dial redis: ", err.Error())
	}

	kr.psc = redis.PubSubConn{kr.conn}

}

func (kr *Kredis) Start() {
	kr.dial()
}

func (kr *Kredis) GetCounter(exchange, pair string) (int, error) {

	key := fmt.Sprintf("%s_%s", exchange, pair)

	//log.Debugf("GetCounter %s", key)

	kr.mu.Lock()
	countUntype, err := kr.conn.Do("LLEN", key)
	kr.mu.Unlock()
	if err != nil {
		return 0, err
	}

	switch v := countUntype.(type) {

	case int64:
		return int(countUntype.(int64)), nil

	case []uint8:
		countString := countUntype.(string)
		countInt, err := strconv.ParseInt(countString, 10, 64)
		if err != nil {
			return 0, err
		}
		return int(countInt), nil

	case []interface{}:
		log.Info("Initializing...")
		return 0, nil

	default:
		thisError := fmt.Errorf("BAD GETCOUNTER TYPE: %T!\n", v)
		return 0, thisError

	}
	return 0, err
}

func (kr *Kredis) DeleteList(exchange, pair string) error {

	key := fmt.Sprintf("%s_%s", exchange, pair)

	kr.mu.Lock()
	_, err := kr.conn.Do("DEL", key)
	kr.mu.Unlock()
	if err != nil {
		return fmt.Errorf("While deleting list %s: %s", key, err.Error)
	}

	return nil

}

func (kr *Kredis) Add(exchange, pair string, value float64) error {

	key := fmt.Sprintf("%s_%s", exchange, pair)
	//keyp := fmt.Sprintf("PUB_%s_%s", exchange, pair)

	count, err := kr.GetCounter(exchange, pair)

	if err != nil {
		return fmt.Errorf("While extracting counter :%s", err.Error())
	}

	if count == kr.size {
		kr.mu.Lock()
		_, err := kr.conn.Do("RPOP", key)
		kr.mu.Unlock()
		if err != nil {
			return fmt.Errorf("While poping the last element: %s", err.Error)
		}

	}

	valueStr := strconv.FormatFloat(value, 'f', 6, 64)

	kr.mu.Lock()
	_, err = kr.conn.Do("LPUSH", key, valueStr)
	kr.mu.Unlock()
	if err != nil {
		return err
	}

	return err
}

func (kr *Kredis) AddString(exchange, pair string, value interface{}) error {

	key := fmt.Sprintf("%s_%s", exchange, pair)

	count, err := kr.GetCounter(exchange, pair)

	if err != nil {
		return fmt.Errorf("While extracting counter :%s", err.Error())
	}

	if count == kr.size {

		kr.mu.Lock()
		_, err := kr.connUpdateList.Do("RPOP", key)
		kr.mu.Unlock()
		if err != nil {
			return fmt.Errorf("While poping the last element: %s", err.Error)
		}

	}

	kr.mu.Lock()
	_, err = kr.connUpdateList.Do("LPUSH", key, value)
	kr.mu.Unlock()
	if err != nil {
		return err
	}

	kr.mu.Lock()
	_, err = kr.connUpdateList.Do("PUBLISH", key, value)
	kr.mu.Unlock()
	if err != nil {
		return err
	}

	//log.Debug("LPUSH to key: ", key)

	return err
}

func (kr *Kredis) AddStringLong(exchange, pair string, value interface{}) error {

	key := fmt.Sprintf("%s_%s", exchange, pair)

	kr.mu.Lock()
	//log.Info("LPUSH: ", key, value)
	_, err := kr.connUpdateList.Do("LPUSH", key, value)
	kr.mu.Unlock()
	if err != nil {
		return err
	}

	kr.mu.Lock()
	_, err = kr.connUpdateList.Do("PUBLISH", key, value)
	kr.mu.Unlock()
	if err != nil {
		return err
	}

	return err
}

func (kr *Kredis) Update(exchange, pair string, value string) error {

	key := fmt.Sprintf("PRICE_%s_%s", exchange, pair)
	kr.mu.Lock()
	_, err := kr.conn.Do("SET", key, value)
	kr.mu.Unlock()
	if err != nil {
		return err
	}

	//_, err = kr.conn.Do("PUBLISH", key, key)

	return err
}

func (kr *Kredis) UpdateList(exchange, pair string) (valueString string, err error) {

	currentKey := fmt.Sprintf("PRICE_%s_%s", exchange, pair)
	//log.Debug("update list: ", currentKey)

	kr.mu.Lock()
	currentValue, err := kr.connUpdateList.Do("GET", currentKey)
	kr.mu.Unlock()
	if err != nil {
		log.Errorf("UpdateList on GET %s: %s ", currentKey, err.Error())
	}

	err = kr.AddString(exchange, pair, currentValue)

	if err != nil {
		log.Error("UpdateList on AddString: ", err.Error())
	}

	valueString = string(currentValue.([]uint8))

	return valueString, err
}

func (kr *Kredis) GetPriceValue(exchange, pair string) (valueString interface{}, err error) {

	currentKey := fmt.Sprintf("PRICE_%s_%s", exchange, pair)
	//log.Debug("update list: ", currentKey)

	kr.mu.Lock()
	currentValue, err := kr.connUpdateList.Do("GET", currentKey)
	kr.mu.Unlock()
	if err != nil {
		log.Errorf("UpdateList on GET %s: %s ", currentKey, err.Error())
	}

	return currentValue, err

}

func (kr *Kredis) PushToPriceList(value interface{}, exchange, pair string) (retValue string, err error) {
	err = kr.AddString(exchange, pair, value)

	if err != nil {
		log.Error("UpdateList on AddString: ", err.Error())
	}

	retValue = string(value.([]uint8))

	return retValue, err
}

func (kr *Kredis) GetLatest(exchange, pair string) (float64, error) {

	key := fmt.Sprintf("%s_%s", exchange, pair)

	kr.mu.Lock()
	valueInt, err := kr.conn.Do("LINDEX", key, 0)
	kr.mu.Unlock()

	valueStr := string(valueInt.([]uint8))

	value, err := strconv.ParseFloat(valueStr, 64)

	if err != nil {
		err = fmt.Errorf("whie converting string to float:", err.Error())
	}

	return value, err

}

func (kr *Kredis) GetLatestValue(key string) (string, error) {
	kr.mu.Lock()
	valueInt, err := kr.conn.Do("LINDEX", key, 0)
	kr.mu.Unlock()

	if err != nil {
		err = fmt.Errorf("while getting value:", err.Error())
	}

	valueStr := string(valueInt.([]uint8))

	return valueStr, err
}

func (kr *Kredis) GetRange(key string, size int) (retList []string, err error) {

	kr.mu.Lock()
	countUntype, err := kr.conn.Do("LLEN", key)
	kr.mu.Unlock()
	if err != nil {
		return []string{}, err
	}

	countInt := int(countUntype.(int64))

	getSize := size

	if countInt < getSize {
		getSize = countInt
	}

	retList = make([]string, getSize)

	kr.mu.Lock()
	rawList, err := kr.conn.Do("LRANGE", key, 0, getSize-1)
	kr.mu.Unlock()

	if err != nil {
		err = fmt.Errorf("while getting value:", err.Error())
	}

	for ID, element := range rawList.([]interface{}) {
		valueStr := string(element.([]uint8))
		if err != nil {
			return retList, err
		}
		retList[ID] = valueStr
	}

	return retList, err
}

func (kr *Kredis) GetList(exchange, pair string) (retList []float64, err error) {

	key := fmt.Sprintf("%s_%s", exchange, pair)

	size, err := kr.GetCounter(exchange, pair)
	if err != nil {
		return retList, err
	}

	retList = make([]float64, size)
	kr.mu.Lock()
	rawList, err := kr.conn.Do("LRANGE", key, 0, -1)
	kr.mu.Unlock()

	if err != nil {
		return retList, err
	}

	for ID, element := range rawList.([]interface{}) {

		valueStr := string(element.([]uint8))
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			log.Error("Parsing error")
			return retList, err
		}

		retList[ID] = value

	}
	return retList, err
}

func (kr *Kredis) Subscribe(chanName string, foo func(value float64)) interface{} {
	kr.mu.Lock()
	psc := redis.PubSubConn{Conn: kr.conn}
	psc.Subscribe(chanName)
	kr.mu.Unlock()

	for {

		log.Info("RECIEVE")

		switch v := psc.Receive().(type) {

		case redis.Message:
			fmt.Printf("%s: message: %s\n", v.Channel, v.Data)
			valueString := string(v.Data)
			value, err := strconv.ParseFloat(valueString, 64)
			if err != nil {
				log.Error(err.Error())
			}
			foo(value)
			log.Info("END FOO")
		case redis.Subscription:
			log.Infof("Subscribed to price updates for : %s", v.Channel)
			//fmt.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
		case error:
			return v
		}
	}

}

func (kr *Kredis) SubscribeLookup(chanName string) {
	kr.mu.Lock()
	kr.psc.Subscribe(chanName)
	kr.mu.Unlock()
}

func (kr *Kredis) SubscriberChann() chan []string {
	return kr.subsChan
}

func (kr *Kredis) SubscriberMonitor() {
	return
	for {

		switch v := kr.psc.Receive().(type) {

		case redis.Message:
			//fmt.Printf("%s: message: %s\n", v.Channel, v.Data)
			//log.Infof("REDIS CHAN RECIEVE: %s , %s", string(v.Channel), string(v.Data))
			resp := make([]string, 2)
			resp[0] = string(v.Channel)
			resp[1] = string(v.Data)
			kr.subsChan <- resp
			//log.Info("END FOO")
		case redis.Subscription:
			log.Infof("Subscribed to data updates for : %s", v.Channel)
			//fmt.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
		case error:
			log.Error("SubscriberMonitor Error")
			return
		}
	}

}
