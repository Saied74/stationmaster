package main

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)

//<<<--------------------   Generating ADIF file     ---------------------->>>
type parsPat struct {
	testName string
	call     []string
	band     []string
	mode     []string
	sent     []string
	rcvd     []string
	name     []string
	contains []string
}

var errMismatch = errors.New("parsPat fields mismatch")

func (tp *parsPat) testRows() ([]LogsRow, error) {
	l := len(tp.call)
	switch {
	case len(tp.band) != l:
		return []LogsRow{}, errMismatch
	case len(tp.mode) != l:
		return []LogsRow{}, errMismatch
	case len(tp.sent) != l:
		return []LogsRow{}, errMismatch
	case len(tp.rcvd) != l:
		return []LogsRow{}, errMismatch
	case len(tp.name) != l:
		return []LogsRow{}, errMismatch
	}
	rows := []LogsRow{}
	for i := 0; i < l; i++ {
		row := LogsRow{}
		row.Call = tp.call[i]
		row.Band = tp.band[i]
		row.Mode = tp.mode[i]
		row.Sent = tp.sent[i]
		row.Rcvd = tp.rcvd[i]
		row.Name = tp.name[i]
		rows = append(rows, row)
	}
	return rows, nil
}

type mockWrite struct {
	testBuffer bytes.Buffer
}

type mockRead struct {
	testBuffer []byte
}

func (m *mockWrite) write(filename string, c []byte) error {
	n, _ := m.testBuffer.Write(c)
	fmt.Println("written to buffer", n)
	return nil
}

func (m *mockWrite) read(filename string) ([]byte, error) {
	p := make([]byte, m.testBuffer.Len())
	m.testBuffer.Read(p)
	return p, nil
}

func (mr *mockRead) read(filename string) ([]byte, error) {
	return mr.testBuffer, nil
}

func TestGenADIFFile(t *testing.T) {
	var tps = []parsPat{
		{
			testName: "clean test",
			call:     []string{"AD2CC", "W2MMT", "N2KF"},
			band:     []string{"20m", "20m", "40m"},
			mode:     []string{"SSB", "CW", "SSB"},
			sent:     []string{"59", "488", "55"},
			rcvd:     []string{"59", "599", "48"},
			name:     []string{"Joe", "Frank", "Harry"},
			contains: []string{"AD2CC", "20M", "SSB", "59", "59", "Joe",
				"W2MMT", "20M", "CW", "488", "599", "Frank",
				"N2KF", "40M", "SSB", "55", "48", "Harry",
			},
		},
		{
			testName: "USB LSB test",
			call:     []string{"AD2CC", "W2MMT", "N2KF"},
			band:     []string{"20m", "20m", "40m"},
			mode:     []string{"USB", "CW", "LSB"},
			sent:     []string{"59", "488", "55"},
			rcvd:     []string{"59", "599", "48"},
			name:     []string{"Joe", "Frank", "Harry"},
			contains: []string{"AD2CC", "20M", "SSB", "59", "59", "Joe",
				"W2MMT", "20M", "CW", "488", "599", "Frank",
				"N2KF", "40M", "SSB", "55", "48", "Harry",
			},
		},
	}
	app := &application{}
	for _, tp := range tps {
		writeControl = &mockWrite{}
		t.Run(tp.testName, func(t *testing.T) {
			rows, err := tp.testRows()
			if err != nil {
				t.Errorf("test rows did not map into logs row %v", err)
			}
			err = app.genADIFFile(rows)
			if err != nil {
				t.Errorf("Error from getADIFFile %v", err)
			}
			testBuffer, _ := writeControl.read("abc")
			for _, item := range tp.contains {
				if !bytes.Contains(testBuffer, []byte(item)) {
					t.Errorf("ADIF file did not contain %s", item)
				}
			}
		})
	}
}

//<<<------------------   Testing parsing ADIF file  ---------------------->>>
type genPat struct {
	name  string
	input string
	start int      // start position of this item.
	pos   int      // current position in the input.
	width int      // width of last rune read from input.
	field itemType //field to emit into the chanel
	items item
}

func (tp *genPat) testLexer() *lexer {
	return &lexer{
		input: tp.input,
		start: tp.start,
		pos:   tp.pos,
		width: tp.width,
		items: make(chan item),
	}
}

func TestGetQSLData(t *testing.T) {
	testBuffer := []byte(`ARRL Logbook of the World Status Report
Generated at 2021-08-23 04:17:35
for ad2cc
Query:
    QSL ONLY: YES
QSL RX SINCE: 2021-08-16 00:00:00 (user supplied value)

<PROGRAMID:4>LoTW
<APP_LoTW_LASTQSL:19>2021-08-23 04:13:11

<APP_LoTW_NUMREC:1>7

<eoh>

<APP_LoTW_OWNCALL:5>AD2CC
<STATION_CALLSIGN:5>AD2CC
<CALL:4>AC0W
<BAND:3>20M
<MODE:3>SSB
<APP_LoTW_MODEGROUP:5>PHONE
<QSO_DATE:8>20210821
<APP_LoTW_RXQSO:19>2021-08-23 04:13:11 // QSO record inserted/modified at LoTW
<TIME_ON:6>190600
<APP_LoTW_QSO_TIMESTAMP:20>2021-08-21T19:06:00Z // QSO Date & Time; ISO-8601
<QSL_RCVD:1>Y
<QSLRDATE:8>20210823
<APP_LoTW_RXQSL:19>2021-08-23 04:13:11 // QSL record matched/modified at LoTW
<eor>

<APP_LoTW_OWNCALL:5>AD2CC
<STATION_CALLSIGN:5>AD2CC
<CALL:4>KE4Q
<BAND:3>20M
<MODE:3>SSB
<APP_LoTW_MODEGROUP:5>PHONE
<QSO_DATE:8>20210822
<APP_LoTW_RXQSO:19>2021-08-23 04:13:11 // QSO record inserted/modified at LoTW
<TIME_ON:6>004700
<APP_LoTW_QSO_TIMESTAMP:20>2021-08-22T00:47:00Z // QSO Date & Time; ISO-8601
<QSL_RCVD:1>Y
<QSLRDATE:8>20210823
<APP_LoTW_RXQSL:19>2021-08-23 04:13:11 // QSL record matched/modified at LoTW
<eor>`)
	app := &application{}
	readControl = &mockRead{testBuffer: testBuffer}
	adifData, err := app.getQSLData("abc")
	if err != nil {
		t.Fatal("getQSLData returned error ", err)
	}
	if len(adifData) != 2 {
		t.Fatal("length of parsed adif data was not two, it was", len(adifData))
	}
	adi0 := adifData[0]
	if adi0[itemCall] != "AC0W" {
		t.Errorf("getQSLData %v expected AC0W got %s", itemCall, adi0[itemCall])
	}
	if adi0[itemBand] != "20M" {
		t.Errorf("getQSLData %v expected 20M got %s", itemBand, adi0[itemBand])
	}
	if adi0[itemMode] != "SSB" {
		t.Errorf("getQSLData %v expected SSB got %s", itemMode, adi0[itemMode])
	}
	if adi0[itemQSOTimeStamp] != "2021-08-21T19:06:00Z" {
		t.Errorf("getQSLData %v expected 2021-08-21T19:06:00Z got %s", itemQSOTimeStamp, adi0[itemQSOTimeStamp])
	}
	if adi0[itemQSLrcvd] != "Y" {
		t.Errorf("getQSLData %v expected Y got %s", itemQSLrcvd, adi0[itemQSLrcvd])
	}
	if adi0[itemRxQSO] != "2021-08-23 04:13:11" {
		t.Errorf("getQSLData %v expected 2021-08-23 04:13:11 got %s", itemRxQSO, adi0[itemRxQSO])
	}
	if adi0[itemRxQSL] != "2021-08-23 04:13:11" {
		t.Errorf("getQSLData %v expected 2021-08-23 04:13:11 got %s", itemRxQSL, adi0[itemRxQSL])
	}
}

func TestLexTest(t *testing.T) {
	var tps = []genPat{
		{
			name:  "valid eoh",
			input: "this that or the other thing <eoh> and more",
			items: item{itemEOF, ""},
		},
		{
			name:  "end of file",
			input: "this that or the other thing and more",
			items: item{itemEOF, ""},
		},
		{
			name:  "empty input",
			input: "",
			items: item{itemEOF, ""},
		},
	}
	for i, tp := range tps {
		switch i {
		case 0:
			l := tp.testLexer()
			f := lexText(l)
			if f == nil {
				t.Errorf("expectd not nil returned, got nil")
			}
		default:
			l := tp.testLexer()
			go lexText(l)
			x := <-l.items
			if x.typ != tp.items.typ {
				t.Errorf("expected %v got %v", tp.items.typ, x.typ)
			}
		}
	}
}

func TestLexFields(t *testing.T) {
	var tps = []genPat{
		{
			name:  "call",
			input: "<CALL: 5>AD2CC",
			items: item{itemCall, ""},
		},
		{
			name:  "band",
			input: "<BAND: 3>20m testing",
			items: item{itemBand, ""},
		},
		{
			name:  "mode",
			input: "<MODE: 3>SSB xxx",
			items: item{itemMode, ""},
		},
		{
			name:  "rxQSO",
			input: "<APP_LoTW_RXQSO: 3>YEX xxx",
			items: item{itemRxQSO, ""},
		},
		{
			name:  "rxQSOTimeStamp",
			input: "<APP_LoTW_QSO_TIMESTAMP: 8>    ",
			items: item{itemQSOTimeStamp, ""},
		},
		{
			name:  "rxQSL",
			input: "<APP_LoTW_RXQSL: 8>    ",
			items: item{itemRxQSL, ""},
		},
		{
			name:  "QSL Rcvd",
			input: "<QSL_RCVD: 8>    ",
			items: item{itemQSLrcvd, ""},
		},
	}

	for _, tp := range tps {
		t.Run(tp.name, func(t *testing.T) {
			l := tp.testLexer()
			lexFields(l)
			if l.field != tp.items.typ {
				t.Errorf("expected item type %v got %v", tp.items.typ, l.field)
			}
		})
	}
}

func TestLexFields2(t *testing.T) {
	var tps = []genPat{
		{
			name:  "End of record",
			input: "<eor>", // " <eor> ",
			items: item{itemEOR, ""},
		},
		{
			name:  "End of file",
			input: "",
			items: item{itemEOF, ""},
		},
		{
			name:  "log of the world end of file",
			input: "   <APP_LoTW_EOF>   ",
			items: item{itemEOF, ""},
		},
	}

	for _, tp := range tps {
		t.Run(tp.name, func(t *testing.T) {
			l := tp.testLexer()
			go lexFields(l)
			x := <-l.items
			if x.typ != tp.items.typ {
				t.Errorf("expected item type %v got %v", tp.items.typ, x.typ)
			}
		})
	}
}

func TestLexRightMeta(t *testing.T) {
	var tps = []genPat{
		{
			name:  "good pattern",
			input: "5>AD2CC this is a test",
			items: item{itemCall, "AD2CC"},
		},
		{
			name:  "bad number",
			input: "b>AD2CC this is a test",
			items: item{itemError, ""},
		},
		{
			name:  "no data",
			input: "",
			items: item{itemEOF, ""},
		},
	}
	for _, tp := range tps {
		t.Run(tp.name, func(t *testing.T) {
			l := &lexer{}
			l.input = tp.input
			l.start = tp.start
			l.pos = tp.pos
			l.width = tp.width
			l.items = make(chan item)
			go lexRightMeta(l)
			x := <-l.items
			if x.typ != tp.items.typ {
				t.Errorf("expected item type %v got %v", tp.items.typ, x.typ)
			}
			if x.typ != itemEOF && x.typ != itemEOR && x.typ != itemError {
				if x.val != tp.items.val {
					t.Errorf("expected item value %v got %v", tp.items.val, x.val)
				}
			}
		})
	}

}
