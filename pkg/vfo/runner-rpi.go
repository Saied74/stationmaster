// +build rpi

package vfo

//this is a test program for the Analog Devices AD9850 device and the
//nooElec board using it.  It is in anticipation of incorporating it
//into  the stationmaster program and driving the VPO of  the Ten Tec
//Omni D.

//Initially, the AD9850 will be driven from the serial  interface.
//This interface uses the following pins:
//Pin 22 Reset Pin 
//Pin 7 Word Clock for colcking in the data
//Pin 8 Freqquency Update for updating the DDS with newly loaded data
//Pin 25 (D7) serial input pin

//Currently, pins 3, 7, 8, and 10 are used for the keyer
//For the DDS, we will  use pins 36, 37, 38, and 40.
//These numbers are all physical GPIO pin numbers (not logical)
//Pin assignment is as follows:
//Pin 36: Reset (reset)
//Pin 37: Word Clock (wordClock)
//Pin 38: Frequency update (freqUpdate
//Pin 40: Data input (dataInput)

//For variable names, we have adopted the golang adaptation of
//the Analaog devices pin names as shown in the parantheses above


import (
//	"context"
//	"fmt"
	"math"
	"time"

	"gobot.io/x/gobot/platforms/raspi"
)

const(
reset = "26"
wordClock = "24"
freqUpdate = "18"
data = "22"
high = byte(1)
low = byte(0)
)

func Initvfo(n int) *raspi.Adaptor{
	ad := raspi.NewAdaptor()
	
	//set all pins to  low
	ad.DigitalWrite(reset, byte(0))
	ad.DigitalWrite(wordClock, byte(0))
	ad.DigitalWrite(freqUpdate, byte(0))
	ad.DigitalWrite(data, byte(0))
	
	//reset the chip
	ad.DigitalWrite(reset, byte(1))
	time.Sleep(time.Duration(1)*time.Microsecond)
	ad.DigitalWrite(reset, byte(0))
	s := serialize(n)
    streamData(ad, s)
    streamData(ad, s)
	return ad
}


func Runvfo(ad *raspi.Adaptor, xFreq, rFreq float64) {
	//These two lines commented until I implement the split feature
	//xPhase := (xFreq * math.Pow(2.0, 32.0)) / 125.0
	//xP := int(math.Round(xPhase))
	rPhase := (rFreq * math.Pow(2.0, 32.0)) / 125.0
	rP := int(math.Round(rPhase))
	//number := rcv
	s := serialize(rP)

	streamData(ad, s)
//	fmt.Println("returning from RunVFO")
}

func streamData(ad *raspi.Adaptor, d []byte) {
	for _, dd := range d {
		ad.DigitalWrite(data, dd)
		ad.DigitalWrite(wordClock, high)
		ad.DigitalWrite(wordClock, low)
//		fmt.Println(dd)
	}
	ad.DigitalWrite(freqUpdate, high)
	ad.DigitalWrite(freqUpdate, low)
//	fmt.Println("returning from streamData")
}


//Serizalize takes in an integer and returns a binary slice of length 40
//That is the input word for the AD9850 input.  LSB will be in address
//zero of the slice.  The first 32 enteries will be the binary digits
//of the decimal number.  The last byte will  be the control byte that
//will be returned as all zeros
func serialize(n int) []byte {
	var y int 
	d := make([]byte, 40)
	for i := range d {
		y = n & 0x1
		d[i] = byte(y)
		n = n >> 1
		if n == 0 {break}
	}
	return d
}
