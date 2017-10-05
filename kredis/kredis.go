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
	connUpdateList redis.Conn
	server         string
	validKeys      map[string]bool
	size           int
	count          uint
	mu             *sync.Mutex
}

func NewKredis(size int) *Kredis {
	kr := &Kredis{}
	kr.size = size
	kr.server = ":6379"
	kr.mu = &sync.Mutex{}
	kr.validKeys = make(map[string]bool)
	kr.validKeys["CEXIO_BTCUSD"] = true

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

}

func (kr *Kredis) Start() {
	kr.dial()
}

func (kr *Kredis) GetCounter(exchange, pair string) (int, error) {

	key := fmt.Sprintf("%s_%s", exchange, pair)

	log.Debugf("GetCounter %s", key)

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

	log.Debug("LPUSH to key: ", key)

	return err
}

func (kr *Kredis) Update(exchange, pair string, value string) error {

	key := fmt.Sprintf("PRICE_%s_%s", exchange, pair)
	//keyp := fmt.Sprintf("PUB_%s_%s", exchange, pair)
	log.Info(key, value)
	kr.mu.Lock()
	_, err := kr.conn.Do("SET", key, value)
	kr.mu.Unlock()
	if err != nil {
		return err
	}

	//_, err = kr.conn.Do("PUBLISH", key, key)

	return err
}

func (kr *Kredis) UpdateList(exchange, pair string) error {

	currentKey := fmt.Sprintf("PRICE_%s_%s", exchange, pair)
	log.Debug("update list: ", currentKey)

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

	return err
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
			return retList, err
		}

		retList[ID] = value

	}
	return retList, err
}
