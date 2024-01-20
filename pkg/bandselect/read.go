//go:build !rpi
// +build !rpi

package bandselect

import (

	"gobot.io/x/gobot/platforms/raspi"
)

type BandData struct {
	Band    chan int
	Adaptor *raspi.Adaptor
}

func BandRead(bd *BandData) int {
	return 3
}
