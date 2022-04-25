// +build rpi

package code

import (
	"context"
	"fmt"
	"time"

	"gobot.io/x/gobot/platforms/raspi"
)

const (
	dotLen       = "Dot Length"
	letterFact   = "Letter Spacing Factor"
	wordFact     = "Word Spacing Factor"
	eol          = "\n"
	lineLength   = 100  //number of letters and spaces per printed line
	ditInput     = "3"  // GPIO physical input pin 11
	dahInput     = "7"  // GPIO physical input pin 7
	TutorOutput  = "16" // GPIO physical output pin 10
	KeyerOutput  = "12"  //GPIO Pin 8
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
	Hi        byte
	Low       byte
}

func (cw *CwDriver) Work(ctx context.Context) {
	//the next few lines and the method "calcSpacing implement
	//the Farnsworth code speed model which can be found here:
	//http://www.arrl.org/files/file/Technology/x9004008.pdf

	fmt.Printf("speed: %f, farnspeed: %f, lf: %f, wf: %f\n", cw.Speed,
		cw.Farnspeed, cw.LF, cw.WF)
	if cw.Speed >= 18.0 {
		cw.dL = 1200.0 / cw.Speed
	} else {
		cw.dL = 1200.0 / cw.Farnspeed
	}
	uwm, ulm := cw.calcSpacing()

	fmt.Printf("DL: %f, uwm: %v, ulm: %v\n", cw.dL, uwm, ulm)

	cw.Dit.DigitalWrite(cw.Output, cw.Low)
	letterTimer := time.Now()
	wordTimer := time.Now()
	setL := false
	setW := false
	for {
		//read the paddle - note dots take precedent
		dit, _ := cw.Dit.DigitalRead(ditInput)
		dah, _ := cw.Dit.DigitalRead(dahInput)
		//if dot, close contact one dot length, open one dot length
		if dit == 0 && debounce(cw.Dit, 0, cw.Output) {

			cw.emit("0")
			setL = true
			

			cw.Dit.DigitalWrite(cw.Output, cw.Hi)
			time.Sleep(time.Duration(cw.dL) * time.Millisecond)
			cw.Dit.DigitalWrite(cw.Output, cw.Low)
			time.Sleep(time.Duration(cw.dL) * time.Millisecond)
			letterTimer = time.Now()
			wordTimer = time.Now()
		}
		//if dash, close contact for three dot lengths, open for one.
		if dah == 0 && debounce(cw.Dit, 0, "7") {
			cw.emit("1")
			setL = true
			
			cw.Dit.DigitalWrite(cw.Output, cw.Hi)
			time.Sleep(time.Duration(cw.dL*3) * time.Millisecond)
			cw.Dit.DigitalWrite(cw.Output, cw.Low)
			time.Sleep(time.Duration(cw.dL) * time.Millisecond)
			letterTimer = time.Now()
			wordTimer = time.Now()

		}
		//if nothing happens longer than upper letter margin,
		//emit the letter
		if time.Now().After(letterTimer.Add(ulm)) && setL {
			cw.emit("L")
			setL = false
			setW = true
		}
		//if nothing happens longer than upper word margin,
		//emit the word
		if time.Now().After(wordTimer.Add(uwm)) && setW {
			cw.emit("W")
			setW = false
		}
		//if the Done channel of the context is closed, return
		select {
		case <-ctx.Done():
			return
		default:
			continue
		}
	}
}

//See the Farnsworth reference above
func (cw *CwDriver) calcSpacing() (uwm, ulm time.Duration) {
	if cw.Speed >= 18.0 {
		uwm := time.Duration(cw.WF*cw.dL*6) * time.Millisecond
		ulm := time.Duration(cw.LF*cw.dL*2) * time.Millisecond
		return uwm, ulm
	}
	dL := 1200 / cw.Speed
	ta := (60.0*cw.Farnspeed - 32.7*cw.Speed) / (cw.Farnspeed * cw.Speed)
	tc := (3 * ta) / 19
	tw := (7 * ta) / 19
	uwm = time.Duration(cw.WF*(tw+7*dL)) * time.Millisecond
	ulm = time.Duration(cw.LF*(tc+3*dL)) * time.Millisecond
	return uwm, ulm
}

//given how the work function loop works, I am not sure if this is needed
//but here it is.
func debounce(ada *raspi.Adaptor, state int, pin string) bool {
	time.Sleep(debounceTime * time.Millisecond)
	newRead, _ := ada.DigitalRead(pin)
	if newRead == state {
		return true
	}
	return false
}

func newCwDriver() *CwDriver {
	return &CwDriver{}
}

func (cw *CwDriver) emit(s string) {

	switch s {
	case "0":
		cw.letter += "0"
		cw.cnt++
		return
	case "1":
		cw.letter += "1"
		cw.cnt++
		return
	case "L":
		fmt.Printf("%s", decode(cw.letter))
		cw.letter = ""
		return
	case "W":
		fmt.Printf(" ")
		if cw.cnt > lineLength {
			fmt.Printf("\n")
			cw.cnt = 0
			return
		}
	default:
		fmt.Printf("Bad bug, got %s when called", s)
		return
	}
}

func decode(s string) string {

	code := map[string]string{
		"01":     "a",
		"1000":   "b",
		"1010":   "c",
		"100":    "d",
		"0":      "e",
		"0010":   "f",
		"110":    "g",
		"0000":   "h",
		"00":     "i",
		"0111":   "j",
		"101":    "k",
		"0100":   "l",
		"11":     "m",
		"10":     "n",
		"111":    "o",
		"0110":   "p",
		"1101":   "q",
		"010":    "r",
		"000":    "s",
		"1":      "t",
		"001":    "u",
		"0001":   "v",
		"011":    "w",
		"1001":   "x",
		"1011":   "y",
		"1100":   "z",
		"01111":  "1",
		"00111":  "2",
		"00011":  "3",
		"00001":  "4",
		"00000":  "5",
		"10000":  "6",
		"11000":  "7",
		"11100":  "8",
		"11110":  "9",
		"11111":  "0",
		"010101": ".",
		"110011": ",",
		"001100": "?",
		"10101":  ";",
		"111000": ":",
		"10010":  "/",
		"100001": "-",
		"10001":  "=",
		"10110":  "kn",
	}
	s, ok := code[s]
	if !ok {
		s = "?"
	}
	return s

}
