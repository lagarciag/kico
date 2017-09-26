package kredis

import (
	"fmt"

	"strconv"

	"github.com/garyburd/redigo/redis"
	log "github.com/sirupsen/logrus"
)

type Kredis struct {
	conn      redis.Conn
	server    string
	validKeys map[string]bool
	size      int
}

func NewKredis(size int) *Kredis {
	kr := &Kredis{}
	kr.size = size
	kr.server = ":6379"

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
}

func (kr *Kredis) GetCounter(exchange, pair string) (int, error) {

	key := fmt.Sprintf("%s_%_", exchange, pair)
	countInt, err := kr.conn.Do("LLEN", key)

	if err != nil {
		return 0, err
	}

	return int(countInt.(int64)), nil
}

func (kr *Kredis) DeleteList(exchange, pair string) error {

	key := fmt.Sprintf("%s_%_", exchange, pair)

	_, err := kr.conn.Do("DEL", key)

	if err != nil {
		return fmt.Errorf("While deleting list %s: %s", key, err.Error)
	}

	return nil

}

func (kr *Kredis) Add(exchange, pair string, value float64) error {

	key := fmt.Sprintf("%s_%_", exchange, pair)

	count, err := kr.GetCounter(exchange, pair)

	if err != nil {
		return fmt.Errorf("While extracting counter :%s", err.Error())
	}

	if count == kr.size {

		_, err := kr.conn.Do("RPOP", key)

		if err != nil {
			return fmt.Errorf("While poping the last element: %s", err.Error)
		}

	}

	valueStr := strconv.FormatFloat(value, 'f', 6, 64)

	_, err = kr.conn.Do("LPUSH", key, valueStr)

	return err
}

func (kr *Kredis) GetLatest(exchange, pair string) (float64, error) {

	key := fmt.Sprintf("%s_%_", exchange, pair)

	valueInt, err := kr.conn.Do("LINDEX", key, 0)

	valueStr := string(valueInt.([]uint8))

	value, err := strconv.ParseFloat(valueStr, 64)

	if err != nil {
		err = fmt.Errorf("whie converting string to float:", err.Error())
	}

	return value, err

}

func (kr *Kredis) GetList(exchange, pair string) (retList []float64, err error) {

	key := fmt.Sprintf("%s_%_", exchange, pair)

	size, err := kr.GetCounter(exchange, pair)
	if err != nil {
		return retList, err
	}

	retList = make([]float64, size)

	rawList, err := kr.conn.Do("LRANGE", key, 0, -1)
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
