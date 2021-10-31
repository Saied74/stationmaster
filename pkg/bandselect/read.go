// +build !rpi

package bandselect

import (
	"time"
)

type BandData struct {
	Band int
}

func BandRead(b chan BandData) {
	i := 0
	for {
		time.Sleep(2 * time.Second)

		b <- BandData{i}
		i++
		if i == 8 {
			i = 0
		}
	}
}
