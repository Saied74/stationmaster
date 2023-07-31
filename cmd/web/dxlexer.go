package main

import (
	"fmt"
//	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

/*
var pattern = `AD2CC de W1NR 28-Jul-2023 2108Z dxspider >

  14043.0 KE0YDN      28-Jul-2023 2100Z CW                             <KC3M>
  14042.7 WB9HFK      28-Jul-2023 2059Z CW                             <KC3M>
  14037.0 KV0I        28-Jul-2023 2058Z CW                             <KC3M>
  21260.0 PR6T        28-Jul-2023 2058Z SA080 Tinhare Is.            <KA2NUE>
  14033.1 N7AUE       28-Jul-2023 2057Z CW                             <KC3M>
  14031.4 AA2IL       28-Jul-2023 2057Z CW                             <KC3M>
  14041.7 DL4CF       28-Jul-2023 2054Z CW                             <KC3M>
  14030.8 KR2Q        28-Jul-2023 2052Z CW                             <KC3M>
  14240.0 S51DX       28-Jul-2023 2039Z 59 in EPA                    <KC3UTT>
  14035.7 VE6RST      28-Jul-2023 2029Z CW                             <KC3M>
  14029.3 KG5U        28-Jul-2023 2026Z CW                             <KC3M>
  50313.0 KB9NKM      28-Jul-2023 2023Z FT8 FN20wt -> EN70            <AC2PB>
  18069.1 OK1WQ       28-Jul-2023 1957Z Lot of CQ little listening :-(<K8WHA>
  14031.1 9A/S54W     28-Jul-2023 1953Z CW                             <KC3M>
  18117.0 MI0TLG      28-Jul-2023 1951Z                                <KX3C>
  18077.6 HB9CVQ      28-Jul-2023 1945Z Big sig Grid FN20             <K8WHA>
  14041.0 KW5CW       28-Jul-2023 1931Z POTA TX                       <KG2GL>
  14197.0 W9ISF       28-Jul-2023 1920Z op Ken S/E   Human decode    <KA2NUE>
  14197.0 W9ISF       28-Jul-2023 1907Z                               <W2EJR>
  21301.5 1A0C        28-Jul-2023 1858Z Good signal NY                 <N2KI>
AD2CC de W1NR 28-Jul-2023 2108Z dxspider >`
*/

type dxItemType int

type dxItem struct {
	typ dxItemType
	val string
}

const (
	dxItemError dxItemType = iota
	dxItemEOF
	dxItemBegin
	dxItemEnd
	dxItemText
	dxItemFreq
	dxItemDX
	dxItemDate
	dxItemTime
	dxItemInfo
	dxItemDE
)

const (
	dtLength     = 17
	startPattern = "ad2cc"
	deStart      = "<"
	deEnd        = ">"
	myCall       = "ad2cc"
	dxeof          = -1
)

type dxLexer struct {
	name  string    // used only for error reports.
	input string    // the string being scanned.
	start int       // start position of this item.
	pos   int       // current position in the input.
	width int       // width of last rune read from input.
	dxItems chan dxItem // channel of scanned items.
}

type dxStateFn func(*dxLexer) dxStateFn

func (i dxItem) String() string {
	switch i.typ {
	case dxItemEOF:
		return "EOF"
	case dxItemError:
		return i.val
	}
	if len(i.val) > 10 {
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

func (l *dxLexer) run() {
	for state := dxLexPreFreq; state != nil; {
		state = state(l)
	}
	close(l.dxItems) // No more tokens will be delivered.
}

func dxLex(name, input string) (*dxLexer, chan dxItem) {
	l := &dxLexer{
		name:  name,
		input: input,
		dxItems: make(chan dxItem),
	}
	go l.run() // Concurrently run state machine.
	return l, l.dxItems
}

func (l *dxLexer) emit(t dxItemType) {
	l.dxItems <- dxItem{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

// next returns the next rune in the input.
func (l *dxLexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return dxeof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// ignore skips over the pending input before this point.
func (l *dxLexer) ignore() {
	l.start = l.pos
}

// backup steps back one rune.
// Can be called only once per call of next.
func (l *dxLexer) backup() {
	l.pos -= l.width
}

// error returns an error token and terminates the scan
// by passing back a nil pointer that will be the next
// state, terminating l.run.
func (l *dxLexer) errorf(format string, args ...interface{}) dxStateFn {
	l.dxItems <- dxItem{dxItemError, fmt.Sprintf(format, args...)}
	return nil
}

func isSpace(r rune) bool {
	space, _, _, _ := strconv.UnquoteChar(` `, '"')
	tab, _, _, _ := strconv.UnquoteChar(`\t`, '"')
	switch r {
	case space:
		return true
	case tab:
		return true
	}
	return false
}

func isNumber(r rune) bool {
	dot, _, _, _ := strconv.UnquoteChar(`.`, '"')
	if unicode.IsNumber(r) || r == dot {
		return true
	}
	return false
}

func isAlphanumeric(r rune) bool {
	if isNumber(r) || unicode.IsLetter(r) {
		return true
	}
	return false
}

func isDate(r rune) bool {
	dash, _, _, _ := strconv.UnquoteChar(`-`, '"')
	if isAlphanumeric(r) || r == dash {
		return true
	}
	return false
}

func isTime(r rune) bool {
	zee, _, _, _ := strconv.UnquoteChar(`Z`, '"')
	if isNumber(r) || r == zee {
		return true
	}
	return false
}

func isCall(r rune) bool {
	slash, _, _, _ := strconv.UnquoteChar(`/`, '"')
	if isAlphanumeric(r) || r == slash {
		return true
	}
	return false
}

// look for the start pattern (end of the dxspiders prompt + newline)
func dxLexText(l *dxLexer) dxStateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], startPattern) {
			if l.pos > l.start {
				l.emit(dxItemBegin)
			}
			return dxLexPreFreq // Next state.
		}
		if l.next() == dxeof {
			break
		}
	}
	// Correctly reached EOF.
	if l.pos > l.start {
		l.emit(dxItemText)
	}
	l.emit(dxItemEOF) // Useful to make EOF a token.
	return nil      // Stop the run loop.

}

// skip spaces before frequency
func dxLexPreFreq(l *dxLexer) dxStateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], myCall) {
			if l.pos > l.start {
				l.emit(dxItemEOF)
				return nil
			}
		}

		switch r := l.next(); {
		case r == dxeof: // || r == '\n':
			return l.errorf("no frequency field")
		case isSpace(r):
			l.ignore()
		case isNumber(r):
			l.backup()
			return dxLexFreq
		default:
			continue
		}
	}
	return nil
}

// Frequency is composed of numbers and a dot
// munch numbers (includes dot) until space, emit frequency
func dxLexFreq(l *dxLexer) dxStateFn {
	for {
		switch r := l.next(); {
		case r == dxeof || r == '\n':
			return l.errorf("no frequency field")
		case isSpace(r):
			l.backup()
			l.emit(dxItemFreq)
			return dxLexPostFreq
		case isNumber(r):
			continue
		default:
			return l.errorf("bad item in freq field %v", r)
		}
	}
	return nil
}

// skip over spaces until alphanumeric
func dxLexPostFreq(l *dxLexer) dxStateFn {
	for {
		switch r := l.next(); {
		case r == dxeof || r == '\n':
			return l.errorf("no DX call field")
		case isSpace(r):
			l.ignore()
		case isCall(r):
			l.backup()
			return dxLexDX
		default:
			return l.errorf("fell out of the bottom of lexPostFreq")
		}
	}
	return nil
}

// call sign is alphanumerics
// munch alphanumerics until space, emit DX call
func dxLexDX(l *dxLexer) dxStateFn {
	for {
		switch r := l.next(); {
		case r == dxeof || r == '\n':
			return l.errorf("no DX call field")
		case isSpace(r):
			l.backup()
			l.emit(dxItemDX)
			return dxLexPostDX
		case isCall(r):
			continue
		default:
			return l.errorf("fell out of the bottom of lexDX")
		}
	}
	return nil
}

// skip over space until reach alphanumeric
func dxLexPostDX(l *dxLexer) dxStateFn {
	for {
		switch r := l.next(); {
		case r == dxeof || r == '\n':
			return l.errorf("no date field")
		case isSpace(r):
			l.ignore()
		case isDate(r):
			l.backup()
			return dxLexDate
		default:
			return l.errorf("fell out of the bottom of lexPostDX")
		}
	}
	return nil
}

// date is alphanumeric + "-"
// munch date elements until reach space, thene emit date
func dxLexDate(l *dxLexer) dxStateFn {
	for {
		switch r := l.next(); {
		case r == dxeof || r == '\n':
			return l.errorf("no date field")
		case isSpace(r):
			l.backup()
			l.emit(dxItemDate)
			return dxLexPostDate
		case isDate(r):
			continue
		default:
			return l.errorf("fell out of the bottom of lexDate")
		}
	}
	return nil

}

// skip over space until reach time element
func dxLexPostDate(l *dxLexer) dxStateFn {
	for {
		switch r := l.next(); {
		case r == dxeof || r == '\n':
			return l.errorf("no time field")
		case isSpace(r):
			l.ignore()
		case isTime(r):
			l.backup()
			return dxLexTime
		default:
			return l.errorf("fell out of the bottom of lexPostDate")
		}
	}
	return nil
}

// time is numeric + Z
// munch time elements until space, emit time
func dxLexTime(l *dxLexer) dxStateFn {
	for {
		switch r := l.next(); {
		case r == dxeof || r == '\n':
			return l.errorf("no time field")
		case isSpace(r):
			l.backup()
			l.emit(dxItemTime)
			return dxLexInfo
		case isTime(r):
			continue
		default:
			return l.errorf("fell out of the bottom of LexTime")
		}
	}
	return nil
}

// much everything until reach deStart which is <; emit info
func dxLexInfo(l *dxLexer) dxStateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], deStart) {
			if l.pos > l.start {
				l.emit(dxItemInfo)
			}
			return dxLexDE // Next state.
		}
		if l.next() == dxeof {
			break
		}
	}
	// Correctly reached EOF.
	if l.pos > l.start {
		l.emit(dxItemText)
	}
	l.emit(dxItemEOF) // Useful to make EOF a token.
	return nil      // Stop the run loop.

}

// munch everything until read deEnd which is >; emit DE call.
// then jump to lexPreFreq to start processing the next line
func dxLexDE(l *dxLexer) dxStateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], deEnd) {
			if l.pos > l.start {
				l.start++
				l.emit(dxItemDE)
			}
			return dxLexPreFreq // Next state.
		}
		if l.next() == dxeof {
			break
		}
	}
	// Correctly reached EOF.
	if l.pos > l.start {
		l.emit(dxItemText)
	}
	l.emit(dxItemEOF)
	return nil
}

type DXClusters struct {
	DE        string
	DXStation string
	Country   string
	Frequency string
	Date	  string
	Time	  string
	Info	  string
	Need      string
}


//type lineType struct {
	//dxCall string
	//freq   string
	//date   string
	//time   string
	//info   string
	//deCall string
//}

func lexResults(pattern string) ([]DXClusters, error) {
	dx := []DXClusters{}
	l := DXClusters{}
	_, c := dxLex("dxspiders", pattern)
	//l := lineType{}
    b := false
	for {
		d := <-c
		switch d.typ {
		case dxItemError:
			return []DXClusters{}, errNoDXSpots
			//fmt.Println("Error: ", d.val)
			//os.Exit(1)
		case dxItemText:
			continue
			//fmt.Println("Text: ", d.val)
		case dxItemBegin:
			continue
			//fmt.Println("Item Begin: ", d.val)
		case dxItemEnd:
			return dx, nil
			//fmt.Println("END")
		case dxItemFreq:
			l.Frequency = d.val
		case dxItemDX:
			l.DXStation = d.val
		case dxItemDate:
			l.Date = d.val
		case dxItemTime:
			l.Time = d.val
		case dxItemInfo:
			l.Info = d.val
		case dxItemDE:
			l.DE = d.val
			dx = append(dx, l)
			//fmt.Printf("DX Call: %s Freq: %s Date: %s Time: %s Info: %s DE Call: %s\n",
				//l.dxCall, l.freq, l.date, l.time, l.info, l.deCall)
		case dxItemEOF:
            b = true
			break
		default:
			break
		}
        if b {break}
	}
	return dx, nil
}
