// +build !rpi

package bandselect

import (
	"time"

	"gobot.io/x/gobot/platforms/raspi"
)

type BandData struct {
	Band    chan int
	Adaptor *raspi.Adaptor
}

func BandRead(bd *BandData) {
	i := 3
	for {
		time.Sleep(2 * time.Second)

		bd.Band <- i
		// i++
		// if i == 8 {
		// 	i = 0
		// }
	}
}
