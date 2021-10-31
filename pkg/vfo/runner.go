// +build !rpi

package vfo

import (
	"gobot.io/x/gobot/platforms/raspi"
)

const (
	reset      = "36"
	wordClock  = "37"
	freqUpdate = "38"
	data       = "40"
	high       = byte(1)
	low        = byte(0)
)

func Initvfo(n int) *raspi.Adaptor {
	return raspi.NewAdaptor()
}

func Runvfo(ad *raspi.Adaptor, xmt, rcv int) {
	return
}
