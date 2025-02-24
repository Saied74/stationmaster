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

func BandRead(bd *BandData) int {
	var n1, n2, n3 int
	for {
		tStart := time.Now()
		n1 = readOnce(bd)
		time.Sleep(time.Duration(200)*time.Millisecond)
		n2 = readOnce(bd)
		time.Sleep(time.Duration(200)*time.Millisecond)
		n3 = readOnce(bd)
		time.Sleep(time.Duration(200)*time.Millisecond)
		if n1 == n2 && n2 == n3 {
			return n1
		}
		tEnd := time.Now()
		if tEnd.Sub(tStart) > time.Duration(900)*time.Millisecond {
			return n3
		}
	}
}

func readOnce(bd *BandData) int {
		n := 0
		i, _ := bd.Adaptor.DigitalRead(sw2)
		n = i
		n *= 2
		i, _ = bd.Adaptor.DigitalRead(sw1)
		n += i
		n *= 2
		i, _ = bd.Adaptor.DigitalRead(sw0)
		n += i
		return n
}
	
