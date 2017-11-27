package ringbuffer_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/lagarciag/ringbuffer"
)

func TestMain(m *testing.M) {
	seed := time.Now().UTC().UnixNano()
	fmt.Println("SEED:", seed)
	rand.Seed(seed)
	os.Exit(m.Run())
}

func TestSimple(t *testing.T) {

	rb := ringbuffer.NewBuffer(10, true, 0, 0)

	rb.Push(4)
	rb.Push(50)
	rb.Push(10)
	rb.Push(1)
	rb.Push(1)
	rb.Push(60)

	if rb.Low() != 0 {
		t.Error("mistmatch, Low is", rb.Low())
	}

	rb.Push(1)
	rb.Push(1)
	rb.Push(1)

	if rb.High() != 60 {
		t.Error("mistmatch, high is 60")
	}

	if rb.Low() != 0 {
		t.Error("mistmatch, Low is", rb.Low())
	}

	rb.Push(1)

	if rb.High() != 60 {
		t.Error("mistmatch, high is 60")
	}

	if rb.Low() != 1 {
		t.Error("mistmatch, Low is", rb.Low())
	}

}

func TestSimple2(t *testing.T) {

	rb := ringbuffer.NewBuffer(10, true, 0, 0)

	rb.Push(1)
	rb.Push(2)
	rb.Push(3)
	rb.Push(4)
	rb.Push(5)
	rb.Push(6)
	rb.Push(7)
	rb.Push(8)
	rb.Push(9)
	rb.Push(10)
	rb.Push(11)
	rb.Push(12)
	rb.Push(13)
	rb.Push(14)
	rb.Push(15)
	rb.Push(16)
	rb.Push(17)
	rb.Push(18)
	rb.Push(19)
	//rb.Push(20)

	t.Log(rb.Oldest())
	t.Log(rb.MostRecent())
	t.Log(rb.Low())
	t.Log(rb.High())

}

func TestDownTrend(t *testing.T) {

	rb := ringbuffer.NewBuffer(10, true, 0, 0)

	rb.Push(1000)

	t.Log("Low:", rb.Low())

	rb.Push(500)
	rb.Push(400)
	rb.Push(100)
	rb.Push(99)
	rb.Push(60)
	rb.Push(60)
	rb.Push(60)
	rb.Push(60)
	rb.Push(60)
	rb.Push(2)

	t.Log("Low:", rb.Low())

	rb.Push(1)

	if rb.High() != 400 {
		t.Error("mistmatch, high shouldbe 400", rb.High())
	}

	if rb.Low() != 1 {
		t.Error("mistmatch, Low is 1", rb.Low())
	}

	t.Log("HIGH:", rb.High())
	t.Log("Low:", rb.Low())

}

func TestDownTrend2(t *testing.T) {

	rb := ringbuffer.NewBuffer(10, true, 0, 0)

	rb.Push(1000)

	t.Log("Low:", rb.Low())

	rb.Push(500)
	rb.Push(400)
	rb.Push(100)
	rb.Push(1)
	rb.Push(6)
	rb.Push(60)
	rb.Push(60)
	rb.Push(60)
	rb.Push(60)
	rb.Push(23)
	rb.Push(11)
	rb.Push(24)
	rb.Push(2200)
	rb.Push(223)
	t.Log("Low2:", rb.Low())

	rb.Push(23)
	rb.Push(55)
	rb.Push(2)
	rb.Push(2)
	rb.Push(2)
	rb.Push(2)
	rb.Push(2)

	t.Log("Low:", rb.Low())

	rb.Push(1)

	if rb.High() != 2200 {
		t.Error("mistmatch, high shouldbe 400", rb.High())
	}

	if rb.Low() != 1 {
		t.Error("mistmatch, Low is 1", rb.Low())
	}

	t.Log("HIGH:", rb.High())
	t.Log("Low:", rb.Low())

}

func TestUpTrend(t *testing.T) {

	rb := ringbuffer.NewBuffer(10, true, 0, 0)

	rb.Push(1)

	if rb.High() != 1 {
		t.Error("mistmatch, high is 1")
	}

	rb.Push(5)

	if rb.High() != 5 {
		t.Error("mistmatch, high is 5", rb.High())
	}

	rb.Push(4)
	rb.Push(100)
	rb.Push(999)
	rb.Push(6000)
	rb.Push(6000)
	rb.Push(6000)
	rb.Push(6000)
	rb.Push(6000)
	rb.Push(20000)
	rb.Push(100000)

	if rb.High() != 100000 {
		t.Error("mistmatch, high should be 10000", rb.High())
	}

	if rb.Low() != 4 {
		t.Error("mistmatch, Low should be 4", rb.Low())
	}

	t.Log("HIGH:", rb.High())
	t.Log("Low:", rb.Low())

}
