package multiema_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/lagarciag/multiema"
)

func TestMain(m *testing.M) {
	seed := time.Now().UTC().UnixNano()
	fmt.Println("SEED:", seed)
	rand.Seed(seed)
	os.Exit(m.Run())
}

func TestMultiEmaSmoke(t *testing.T) {

	mema := multiema.NewMultiEma(9, 6)

	size := 9 * 6 * 10

	for i := 0; i < size; i++ {
		mema.Add(float64(i))
		t.Logf("Adding %d, ema: %f", i, mema.Value())
	}

}
