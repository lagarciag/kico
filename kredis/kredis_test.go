package kredis

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
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
	formatter.FullTimestamp = false
	formatter.ForceColors = true
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(formatter)

	os.Exit(m.Run())
}

func TestKredisDial(t *testing.T) {

	kr := NewKredis(10)

	kr.dial()

	_, err := kr.conn.Do("PING")

	if err != nil {
		t.Error("Could not connect to redis", err.Error())
		t.FailNow()
	}

}

func TestKredisGetCounter(t *testing.T) {

	kr := NewKredis(10)

	kr.dial()

	value, err := kr.GetCounter("CEXIO", "BTCUSD")

	if err != nil {
		t.Error("ERROR: ", err.Error())
	}

	t.Log("VALUE: ", value)
}

func TestKredisAdd(t *testing.T) {

	kr := NewKredis(10)

	kr.dial()

	exchange := "CEXIO"
	pair := "BTCUSD"

	err := kr.DeleteList(exchange, pair)

	if err != nil {
		t.Error(err.Error())
	}

	counter, err := kr.GetCounter(exchange, pair)

	t.Log("Counter: ", counter)

	if err != nil {
		t.Errorf(err.Error())
	}

	if counter != 0 {
		t.Error("Counter should be 0 : ", counter)
	}

	for n := 0; n < 20; n++ {
		err := kr.Add(exchange, pair, float64(2323.232))
		if err != nil {
			t.Error("Could not connect to redis", err.Error())
			t.FailNow()
		}
	}

	counter, err = kr.GetCounter(exchange, pair)

	if err != nil {
		t.Errorf(err.Error())
	}

	if counter != 10 {
		t.Error("counter mismatch: ", counter)
	}

	t.Log("Counter: ", counter)
}

func TestKredisGet(t *testing.T) {

	size := 1000

	kr := NewKredis(size)

	kr.dial()

	exchange := "CEXIO"
	pair := "BTCUSD"

	err := kr.DeleteList(exchange, pair)

	if err != nil {
		t.Error(err.Error())
	}

	counter, err := kr.GetCounter(exchange, pair)

	t.Log("Counter: ", counter)

	if err != nil {
		t.Errorf(err.Error())
	}

	if counter != 0 {
		t.Error("Counter should be 0 : ", counter)
	}

	for n := 0; n < size*2; n++ {
		err := kr.Add(exchange, pair, float64(n))
		if err != nil {
			t.Error("Could not connect to redis", err.Error())
			t.FailNow()
		}

		value, err := kr.GetLatest(exchange, pair)

		if err != nil {
			t.Error(err.Error())
		}

		if n != int(value) {
			t.Error("No match:", n, value)
		}

	}

	counter, err = kr.GetCounter(exchange, pair)

	if err != nil {
		t.Errorf(err.Error())
	}

	if counter != size {
		t.Error("counter mismatch: ", counter)
	}

	t.Log("Counter: ", counter)
}

func TestKredisGetList(t *testing.T) {

	size := 1000

	kr := NewKredis(size)

	kr.dial()

	exchange := "CEXIO"
	pair := "BTCUSD"

	err := kr.DeleteList(exchange, pair)

	if err != nil {
		t.Error(err.Error())
	}

	counter, err := kr.GetCounter(exchange, pair)

	t.Log("Counter: ", counter)

	if err != nil {
		t.Errorf(err.Error())
	}

	if counter != 0 {
		t.Error("Counter should be 0 : ", counter)
	}

	for n := 0; n < size*2; n++ {
		err := kr.Add(exchange, pair, float64(n))
		if err != nil {
			t.Error("Could not connect to redis", err.Error())
			t.FailNow()
		}

	}

	counter, err = kr.GetCounter(exchange, pair)

	if err != nil {
		t.Errorf(err.Error())
	}

	if counter != size {
		t.Error("counter mismatch: ", counter)
	}

	t.Log("Counter: ", counter)

	list, err := kr.GetList(exchange, pair)

	if err != nil {
		t.Errorf(err.Error())
	}

	//t.Log(list)

	if len(list) != size {
		t.Log("Bad Size")
	}

	last := list[size-1]

	if last != float64(size*2-size) {
		t.Error("last should be 0", last)
	}

	if list[0] != float64((size*2 - 1)) {
		t.Error("first should be :", size-1)
	}

}

func TestKredisPubSub(t *testing.T) {

	kr := NewKredis(10)

	kr.dial()

	exchange := "CEXIO"
	pair := "BTCUSD"

	err := kr.DeleteList(exchange, pair)

	if err != nil {
		t.Error(err.Error())
	}

	counter, err := kr.GetCounter(exchange, pair)

	t.Log("Counter: ", counter)

	if err != nil {
		t.Errorf(err.Error())
	}

	if counter != 0 {
		t.Error("Counter should be 0 : ", counter)
	}

	pubSubClient, err := redis.Dial("tcp", ":6379")

	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	psc := redis.PubSubConn{Conn: pubSubClient}

	//key := fmt.Sprintf("%s_%_", exchange, pair)
	keyp := fmt.Sprintf("%s_%s", exchange, pair)
	psc.Subscribe(keyp)

	go pubSub(psc)

	time.Sleep(time.Second)

	for n := 0; n < 10; n++ {
		err := kr.Add(exchange, pair, float64(2323.232))
		if err != nil {
			t.Error("Could not connect to redis", err.Error())
			t.FailNow()
		}
	}

}

func pubSub(psc redis.PubSubConn) {

	for {

		switch v := psc.Receive().(type) {
		case redis.Message:
			fmt.Printf("MESSG: %v: message: %s\n", v.Channel, v.Data)
		case redis.Subscription:
			fmt.Printf("SUBSC: %s: %s %d\n", v.Channel, v.Kind, v.Count)

		case error:
			panic("BAD!!")
		}
	}
}
