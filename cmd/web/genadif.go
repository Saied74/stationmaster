package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

//ADIF file format can be found at https://adif.org/312/ADIF_312.htm#QSO_Fields
//As I enhance the program (and my skills at ham radio), this will be re-written

func (app *application) genADIFFile(rows []LogsRow) error {
	var b bytes.Buffer
	b = writeHeader(b)

	for _, row := range rows {
		b = writeDateTime(b, row.Time)
		b.Write([]byte(fmt.Sprintf("<call:%d>%s\n", len(row.Call), row.Call)))
		b.Write([]byte(fmt.Sprintf("<band:%d>%s\n", len(row.Band), strings.ToUpper(row.Band))))
		mode := normalizeMode(row.Mode)
		b.Write([]byte(fmt.Sprintf("<mode:%d>%s\n", len(mode), mode)))
		b.Write([]byte(fmt.Sprintf("<rst_sent:%d>%s\n", len(row.Sent), row.Sent)))
		b.Write([]byte(fmt.Sprintf("<rst_rcvd:%d>%s\n", len(row.Rcvd), row.Rcvd)))
		b.Write([]byte(fmt.Sprintf("<name:%d>%s\n", len(row.Name), row.Name)))
		b.Write([]byte("<eor>\n\n"))
	}
	l := b.Len()
	p := make([]byte, l)
	b.Read(p)

	err := writeADIFOutput(app.adifFile, p)
	if err != nil {
		return err
	}
	return nil
}

var testBuffer = []byte{}

func writeADIFOutput(fileName string, b []byte) error {
	switch v := cF.(type) {
	case func(string) (*os.File, error):
		f, err := v(fileName)
		if err != nil {
			return err
		}
		defer f.Close()
		f.Write(b)
	case []byte:
		testBuffer = b
	default:
		return fmt.Errorf("Bad type to writeADIFFile")
	}
	return nil
}

func cookTime(t time.Time) (string, string) {
	tt := fmt.Sprintf("%v", t)
	times := strings.Split(tt, " ")
	ztimes := strings.Split(times[1], ".")
	return times[0], ztimes[0] + "Z"
}

func topLine() string {
	t := time.Now()
	dt, tt := cookTime(t)
	return fmt.Sprintf("Generated on %s at %s for AD2CC\n", dt, tt)
}

func writeHeader(b bytes.Buffer) bytes.Buffer {
	line := topLine()
	b.Write([]byte(line))
	b.Write([]byte("\n"))
	b.Write([]byte("<adif_ver:5>3.0.5\n"))
	programID := "AD2CC Stationmaster"
	b.Write([]byte(fmt.Sprintf("<programid:%d>%s\n", len(programID), programID)))
	// userDef := "AD2CC stationmaster:github.com/Saied74/stationmaster"
	// b.Write([]byte(fmt.Sprintf("<USERDEF1:%d:S>%s\n", len(userDef), userDef)))
	b.Write([]byte("<EOH>\n"))
	b.Write([]byte("\n"))
	return b
}

func writeDateTime(b bytes.Buffer, t time.Time) bytes.Buffer {
	qsoD, qsoT := cookTime(t)
	qsoDate := strings.Join(strings.Split(qsoD, "-"), "")
	qsoTime := strings.Join(strings.Split(qsoT, ":")[0:2], "")
	b.Write([]byte(fmt.Sprintf("<qso_date:%d>%s\n", len(qsoDate), qsoDate)))
	b.Write([]byte(fmt.Sprintf("<time_on:%d>%s\n", len(qsoTime), qsoTime)))
	return b
}

func normalizeMode(s string) string {
	switch s {
	case "USB":
		return "SSB"
	case "LSB":
		return "SSB"
	case "CW":
		return "CW"
	default:
		return "SSB"
	}
}

//<<============== Read, Parse and Return Confirmed QSLs ====================>>

type itemType int

type item struct {
	typ itemType
	val string
}

type lexer struct {
	name  string    // used only for error reports.
	input string    // the string being scanned.
	start int       // start position of this item.
	pos   int       // current position in the input.
	width int       // width of last rune read from input.
	field itemType  //field to emit into the chanel
	items chan item // channel of scanned items.
}

type stateFn func(*lexer) stateFn

const (
	call         = "<CALL:"
	band         = "<BAND:"
	mode         = "<MODE:"
	qsoTimeStamp = "<APP_LoTW_QSO_TIMESTAMP:"
	qslRcvd      = "<QSL_RCVD:"
	rxQSO        = "<APP_LoTW_RXQSO:"
	rxQSL        = "<APP_LoTW_RXQSL:"
	eor          = "<eor>"
	lotwEof      = "<APP_LoTW_EOF>"
)

const (
	itemCall itemType = iota
	itemBand
	itemMode
	itemQSOTimeStamp
	itemQSLrcvd
	itemRxQSO
	itemRxQSL
	itemEOR
	itemEOF
	itemError
)

const eof = -1

var printSeq = []itemType{itemCall, itemBand, itemMode, itemQSOTimeStamp, itemQSLrcvd, itemRxQSO, itemRxQSL}

func (app *application) getQSLData(fileName string) ([]map[itemType]string, error) {
	output := []map[itemType]string{}
	fileName = filepath.Join(app.qslDir, fileName)
	records, err := os.ReadFile(fileName)
	if err != nil {
		return []map[itemType]string{}, err
	}

	_, c := lex("adif", string(records))
	var b bool
	row := map[itemType]string{}
	for {
		d := <-c
		switch d.typ {
		case itemEOR:
			output = append(output, row)
			row = map[itemType]string{}
		case itemEOF:
			b = true
			break
		case itemError:
			reportError(d)
		default:
			row[d.typ] = d.val
		}
		if b {
			break
		}
	}

	return output, nil
}

func reportError(d item) {
	fmt.Printf("Error reported from the channel %v \n", d)
	os.Exit(1)
}

func timeIt(s string) (time.Time, error) {
	// layout := "2021-07-27T22:12:00Z"
	s = strings.Join(strings.Split(s, " "), "T")
	s += "Z"
	return time.Parse(time.RFC3339, s)
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

// run lexes the input by executing state functions until
// the state is nil.
func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.items) // No more tokens will be delivered.
}

func (l *lexer) emit() {
	l.items <- item{l.field, l.input[l.start:l.pos]}
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

func lexText(l *lexer) stateFn {
	const eoh = "<eoh>"
	for {
		if strings.HasPrefix(l.input[l.pos:], eoh) {
			l.start = l.pos
			return lexFields // Next state.
		}
		if l.next() == eof {
			break
		}
	}
	// Correctly reached EOF.
	l.field = itemEOF
	l.emit()   // Useful to make EOF a token.
	return nil // Stop the run loop.
}

func lexFields(l *lexer) stateFn {

	for {
		switch {
		case strings.HasPrefix(l.input[l.pos:], call):
			l.field = itemCall
			l.pos += len(call)
			l.start = l.pos
			return lexRightMeta
		case strings.HasPrefix(l.input[l.pos:], band):
			l.field = itemBand
			l.pos += len(band)
			l.start = l.pos
			return lexRightMeta
		case strings.HasPrefix(l.input[l.pos:], mode):
			l.field = itemMode
			l.pos += len(mode)
			l.start = l.pos
			return lexRightMeta
		case strings.HasPrefix(l.input[l.pos:], qsoTimeStamp):
			l.field = itemQSOTimeStamp
			l.pos += len(qsoTimeStamp)
			l.start = l.pos
			return lexRightMeta
		case strings.HasPrefix(l.input[l.pos:], qslRcvd):
			l.field = itemQSLrcvd
			l.pos += len(qslRcvd)
			l.start = l.pos
			return lexRightMeta
		case strings.HasPrefix(l.input[l.pos:], rxQSO):
			l.field = itemRxQSO
			l.pos += len(rxQSO)
			l.start = l.pos
			return lexRightMeta
		case strings.HasPrefix(l.input[l.pos:], rxQSL):
			l.field = itemRxQSL
			l.pos += len(rxQSL)
			l.start = l.pos
			return lexRightMeta
		case strings.HasPrefix(l.input[l.pos:], eor):
			l.field = itemEOR
			l.pos += len(eor)
			l.emit()
			return lexFields
		case strings.HasPrefix(l.input[l.pos:], lotwEof):
			l.field = itemEOF
			l.pos += len(lotwEof)
			l.emit()
			return nil
		}
		if l.next() == eof {
			break
		}
	}
	// Correctly reached EOF.
	l.field = itemEOF
	l.emit()   // Useful to make EOF a token.
	return nil // Stop the run loop.

}

func lexRightMeta(l *lexer) stateFn {
	rightMeta := ">"
	for {
		if strings.HasPrefix(l.input[l.pos:], rightMeta) {
			if l.pos > l.start {
				w := l.input[l.start:l.pos]
				fieldLen, err := strconv.Atoi(w)
				if err != nil {
					l.field = itemError
					l.emit()
					return nil
				}
				l.start = l.pos + 1
				l.pos = l.start + fieldLen
				l.emit()
			}
			return lexFields // Next state.
		}
		if l.next() == eof {
			break
		}
	}
	l.field = itemEOF
	l.emit() // Useful to make EOF a token.
	return nil
}
