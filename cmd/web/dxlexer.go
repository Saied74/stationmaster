package main

import (
	"fmt"
	//	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

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
	infoMin      = 25
	startPattern = "ad2cc"
	deStart      = "<"
	deEnd        = ">"
	//	myCall       = "ad2cc"
	dxeof = -1
)

type dxLexer struct {
	name    string      // used only for error reports.
	input   string      // the string being scanned.
	start   int         // start position of this item.
	pos     int         // current position in the input.
	width   int         // width of last rune read from input.
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
	for state := dxLexText; state != nil; {
		state = state(l)
	}
	close(l.dxItems) // No more tokens will be delivered.
}

func dxLex(name, input string) (*dxLexer, chan dxItem) {
	l := &dxLexer{
		name:    name,
		input:   input,
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

func isFreq(l *dxLexer) bool {
	pos := l.pos
	var r rune
	next := func() rune {
		if pos >= len(l.input) {
			return eof
		}
		r, l.width = utf8.DecodeRuneInString(l.input[pos:])
		pos += l.width
		return r
	}
	for i := 0; i < 5; i++ {
		r = next()
		if r == eof {
			return false
		}
		if !isNumber(r) {
			return false
		}
	}
	return true
}

// look for the start pattern (end of the dxspiders prompt + newline)
func dxLexText(l *dxLexer) dxStateFn {
	for {
		if isFreq(l) {
			l.emit(dxItemText)
			return dxLexPreFreq
		}
		if r := l.next(); r == eof {
			break
		}
	}

	//for {
	//if strings.HasPrefix(l.input[l.pos:], startPattern) {
	//if l.pos > l.start {
	//l.emit(dxItemBegin)
	////l.pos += len(startPattern)
	////l.emit(dxItemBegin)
	//}
	//return dxLexPreFreq // Next state.
	//}
	//if l.next() == dxeof {
	//break
	//}
	//}
	//// Correctly reached EOF.
	//if l.pos > l.start {
	//l.emit(dxItemText)
	//}
	l.emit(dxItemEOF) // Useful to make EOF a token.
	return nil        // Stop the run loop.

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
		case r == dxeof:
			return l.errorf("no frequency field %v", l.input)
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
		case r == dxeof:
			return l.errorf("no frequency field %v", l.input)
		case isSpace(r):
			l.backup()
			l.emit(dxItemFreq)
			return dxLexPostFreq
		case isNumber(r):
			continue
		default:
			return l.errorf("bad item in freq field %v", l.input)
		}
	}
	return nil
}

// skip over spaces until alphanumeric
func dxLexPostFreq(l *dxLexer) dxStateFn {
	for {
		switch r := l.next(); {
		case r == dxeof:
			return l.errorf("no DX call field %v", l.input)
		case isSpace(r):
			l.ignore()
		case isCall(r):
			l.backup()
			return dxLexDX
		default:
			return l.errorf("fell out of the bottom of lexPostFreq %v", l.input)
		}
	}
	return nil
}

// call sign is alphanumerics
// munch alphanumerics until space, emit DX call
func dxLexDX(l *dxLexer) dxStateFn {
	for {
		switch r := l.next(); {
		case r == dxeof:
			return l.errorf("no DX call field %v", l.input)
		case isSpace(r):
			l.backup()
			l.emit(dxItemDX)
			return dxLexPostDX
		case isCall(r):
			continue
		default:
			return l.errorf("fell out of the bottom of lexDX %v", l.input)
		}
	}
	return nil
}

// skip over space until reach alphanumeric
func dxLexPostDX(l *dxLexer) dxStateFn {
	for {
		switch r := l.next(); {
		case r == dxeof:
			return l.errorf("no date field %v", l.input)
		case isSpace(r):
			l.ignore()
		case isDate(r):
			l.backup()
			return dxLexDate
		default:
			return l.errorf("fell out of the bottom of lexPostDX %v", l.input)
		}
	}
	return nil
}

// date is alphanumeric + "-"
// munch date elements until reach space, thene emit date
func dxLexDate(l *dxLexer) dxStateFn {
	for {
		switch r := l.next(); {
		case r == dxeof:
			return l.errorf("no date field %v", l.input)
		case isSpace(r):
			l.backup()
			l.emit(dxItemDate)
			return dxLexPostDate
		case isDate(r):
			continue
		default:
			return l.errorf("fell out of the bottom of lexDate %v", l.input)
		}
	}
	return nil

}

// skip over space until reach time element
func dxLexPostDate(l *dxLexer) dxStateFn {
	for {
		switch r := l.next(); {
		case r == dxeof:
			return l.errorf("no time field %v", l.input)
		case isSpace(r):
			l.ignore()
		case isTime(r):
			l.backup()
			return dxLexTime
		default:
			return l.errorf("fell out of the bottom of lexPostDate %v", l.input)
		}
	}
	return nil
}

// time is numeric + Z
// munch time elements until space, emit time
func dxLexTime(l *dxLexer) dxStateFn {
	for {
		switch r := l.next(); {
		case r == dxeof:
			return l.errorf("no time field %v", l.input)
		case isSpace(r):
			l.backup()
			l.emit(dxItemTime)
			return dxLexInfo
		case isTime(r):
			continue
		default:
			return l.errorf("fell out of the bottom of LexTime %v", l.input)
		}
	}
	return nil
}

// much everything until reach deStart which is <; emit info
func dxLexInfo(l *dxLexer) dxStateFn {
	l.pos += infoMin
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
	return nil        // Stop the run loop.

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
	Date      string
	Time      string
	Info      string
	Need      string
}

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
			return []DXClusters{}, fmt.Errorf("%v", d.val)
		case dxItemText:
			continue
		case dxItemBegin:
			continue
		case dxItemEnd:
			return dx, nil
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
			//l.DXStation, l.Frequency, l.Date, l.Time, l.Info, l.DE)
		case dxItemEOF:
			b = true
			break
		default:
			break
		}
		if b {
			break
		}
	}
	return dx, nil
}
