package main

import (
	"fmt"
	"os"
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

type itemType int

type item struct {
	typ itemType
	val string
}

const (
	itemError itemType = iota
	itemEOF
	itemBegin
	itemEnd
	itemText
	itemFreq
	itemDX
	itemDate
	itemTime
	itemInfo
	itemDE
)

const (
	dtLength     = 17
	startPattern = "ad2cc"
	deStart      = "<"
	deEnd        = ">"
	myCall       = "ad2cc"
	eof          = -1
)

type lexer struct {
	name  string    // used only for error reports.
	input string    // the string being scanned.
	start int       // start position of this item.
	pos   int       // current position in the input.
	width int       // width of last rune read from input.
	items chan item // channel of scanned items.
}

type stateFn func(*lexer) stateFn

func (i item) String() string {
	switch i.typ {
	case itemEOF:
		return "EOF"
	case itemError:
		return i.val
	}
	if len(i.val) > 10 {
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

func (l *lexer) run() {
	for state := lexPreFreq; state != nil; {
		state = state(l)
	}
	close(l.items) // No more tokens will be delivered.
}

func lex(name, input string) (*lexer, chan item) {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}
	go l.run() // Concurrently run state machine.
	return l, l.items
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

// next returns the next rune in the input.
func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// backup steps back one rune.
// Can be called only once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// error returns an error token and terminates the scan
// by passing back a nil pointer that will be the next
// state, terminating l.run.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, fmt.Sprintf(format, args...)}
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
func lexText(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], startPattern) {
			if l.pos > l.start {
				l.emit(itemBegin)
			}
			return lexPreFreq // Next state.
		}
		if l.next() == eof {
			break
		}
	}
	// Correctly reached EOF.
	if l.pos > l.start {
		l.emit(itemText)
	}
	l.emit(itemEOF) // Useful to make EOF a token.
	return nil      // Stop the run loop.

}

// skip spaces before frequency
func lexPreFreq(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], myCall) {
			if l.pos > l.start {
				l.emit(itemEOF)
				return nil
			}
		}

		switch r := l.next(); {
		case r == eof: // || r == '\n':
			return l.errorf("no frequency field")
		case isSpace(r):
			l.ignore()
		case isNumber(r):
			l.backup()
			return lexFreq
		default:
			continue
		}
	}
	return nil
}

// Frequency is composed of numbers and a dot
// munch numbers (includes dot) until space, emit frequency
func lexFreq(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof || r == '\n':
			return l.errorf("no frequency field")
		case isSpace(r):
			l.backup()
			l.emit(itemFreq)
			return lexPostFreq
		case isNumber(r):
			continue
		default:
			return l.errorf("bad item in freq field %v", r)
		}
	}
	return nil
}

// skip over spaces until alphanumeric
func lexPostFreq(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof || r == '\n':
			return l.errorf("no DX call field")
		case isSpace(r):
			l.ignore()
		case isCall(r):
			l.backup()
			return lexDX
		default:
			return l.errorf("fell out of the bottom of lexPostFreq")
		}
	}
	return nil
}

// call sign is alphanumerics
// munch alphanumerics until space, emit DX call
func lexDX(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof || r == '\n':
			return l.errorf("no DX call field")
		case isSpace(r):
			l.backup()
			l.emit(itemDX)
			return lexPostDX
		case isCall(r):
			continue
		default:
			return l.errorf("fell out of the bottom of lexDX")
		}
	}
	return nil
}

// skip over space until reach alphanumeric
func lexPostDX(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof || r == '\n':
			return l.errorf("no date field")
		case isSpace(r):
			l.ignore()
		case isDate(r):
			l.backup()
			return lexDate
		default:
			return l.errorf("fell out of the bottom of lexPostDX")
		}
	}
	return nil
}

// date is alphanumeric + "-"
// munch date elements until reach space, thene emit date
func lexDate(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof || r == '\n':
			return l.errorf("no date field")
		case isSpace(r):
			l.backup()
			l.emit(itemDate)
			return lexPostDate
		case isDate(r):
			continue
		default:
			return l.errorf("fell out of the bottom of lexDate")
		}
	}
	return nil

}

// skip over space until reach time element
func lexPostDate(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof || r == '\n':
			return l.errorf("no time field")
		case isSpace(r):
			l.ignore()
		case isTime(r):
			l.backup()
			return lexTime
		default:
			return l.errorf("fell out of the bottom of lexPostDate")
		}
	}
	return nil
}

// time is numeric + Z
// munch time elements until space, emit time
func lexTime(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof || r == '\n':
			return l.errorf("no time field")
		case isSpace(r):
			l.backup()
			l.emit(itemTime)
			return lexInfo
		case isTime(r):
			continue
		default:
			return l.errorf("fell out of the bottom of LexTime")
		}
	}
	return nil
}

// much everything until reach deStart which is <; emit info
func lexInfo(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], deStart) {
			if l.pos > l.start {
				l.emit(itemInfo)
			}
			return lexDE // Next state.
		}
		if l.next() == eof {
			break
		}
	}
	// Correctly reached EOF.
	if l.pos > l.start {
		l.emit(itemText)
	}
	l.emit(itemEOF) // Useful to make EOF a token.
	return nil      // Stop the run loop.

}

// munch everything until read deEnd which is >; emit DE call.
// then jump to lexPreFreq to start processing the next line
func lexDE(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], deEnd) {
			if l.pos > l.start {
				l.start++
				l.emit(itemDE)
			}
			return lexPreFreq // Next state.
		}
		if l.next() == eof {
			break
		}
	}
	// Correctly reached EOF.
	if l.pos > l.start {
		l.emit(itemText)
	}
	l.emit(itemEOF)
	return nil
}

type lineType struct {
	dxCall string
	freq   string
	date   string
	time   string
	info   string
	deCall string
}

func lexResults(pattern string) {
	_, c := lex("dxspiders", pattern)
	l := lineType{}
    b := false
	for {
		d := <-c
		switch d.typ {
		case itemError:
			fmt.Println("Error: ", d.val)
			os.Exit(1)
		case itemText:
			fmt.Println("Text: ", d.val)
		case itemBegin:
			fmt.Println("Item Begin: ", d.val)
		case itemEnd:
			fmt.Println("END")
		case itemFreq:
			l.freq = d.val
		case itemDX:
			l.dxCall = d.val
		case itemDate:
			l.date = d.val
		case itemTime:
			l.time = d.val
		case itemInfo:
			l.info = d.val
		case itemDE:
			l.deCall = d.val
			fmt.Printf("DX Call: %s Freq: %s Date: %s Time: %s Info: %s DE Call: %s\n",
				l.dxCall, l.freq, l.date, l.time, l.info, l.deCall)
		case itemEOF:
            b = true
			break
		default:
			break
		}
        if b {break}
	}
}
