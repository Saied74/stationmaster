// +build !rpi

package code

import (
	"context"

	"gobot.io/x/gobot/platforms/raspi"
)

const (
	dotLen     = "Dot Length"
	letterFact = "Letter Spacing Factor"
	wordFact   = "Word Spacing Factor"
	eol        = "\n"
	lineLength = 100  //number of letters and spaces per printed line
	ditInput   = "11" // GPIO physical input pin 11
	dahInput   = "7"  // GPIO physical input pin 7
	// output       = "10" // GPIO physical output pin 10
	TutorOutput  = "10" // GPIO physical output pin 10
	KeyerOutput  = "8"  //GPIO Pin 8
	debounceTime = 2    //in milliseconds
)

type CwDriver struct {
	letter    string
	cnt       int
	Dit       *raspi.Adaptor
	Speed     float64
	Farnspeed float64
	dL        float64 //dot length
	LF        float64 //letter factor
	WF        float64 //word factor
	Output    string
}

func (cw *CwDriver) Work(ctx context.Context) {
	return
}
