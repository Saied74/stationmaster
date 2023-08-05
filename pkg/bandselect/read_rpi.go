//go:build rpi
// +build rpi

package bandselect

import (
	"time"

	"gobot.io/x/gobot/platforms/raspi"
)

const (
	sw2          = "40"
	sw1          = "38"
	sw0          = "36"
	debounceCnt  = 10
	debounceTime = 1 //in milliseconds
)

type BandData struct {
	Band    chan int
	Adaptor *raspi.Adaptor
}

func BandRead(bd *BandData) {
	for {
		n := 0
		i, _ := bd.Adaptor.DigitalRead(sw2)
		n = i
		n *= 2
		i, _ = bd.Adaptor.DigitalRead(sw1)
		n += i
		n *= 2
		i, _ = bd.Adaptor.DigitalRead(sw0)
		n += i
		bd.Band <- n
		time.Sleep(time.Duration(500) * time.Millisecond)
	}
}
