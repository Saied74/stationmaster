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
	//TutorOutput is the output driving the tutor circuit.
	TutorOutput = "10" // GPIO physical output pin 10
	//KeyerOutput is the pin driving the radio keyer, tone is coming from the radio
	KeyerOutput  = "8" //GPIO Pin 8
	debounceTime = 2   //in milliseconds
)

//CwDriver injects the required data into the CW work function.
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
	Hi        byte
	Low       byte
    RcvFreq	  float64
	Band      string

}

//Work in this file does nothing.  It is for outside of RPI build
func (cw *CwDriver) Work(ctx context.Context) {
	return
}

	
func (cw *CwDriver) BeSilent() {
	return
}
