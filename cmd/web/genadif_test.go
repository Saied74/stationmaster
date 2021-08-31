package main

import (
	"bytes"
	"errors"
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

type mockOSF struct {
	record []byte
}

func (m mockOSF) Close() error {
	return nil
}

func (m mockOSF) Write(p []byte) (int, error) {
	p = m.record
	return len(p), nil
}

func (m mockOSF) Read(p []byte) (int, error) {
	return 0, nil
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
	}
	app := &application{}
	for _, tp := range tps {
		cF = []byte{}
		t.Run(tp.testName, func(t *testing.T) {
			rows, err := tp.testRows()
			if err != nil {
				t.Errorf("test rows did not map into logs row %v", err)
			}
			err = app.genADIFFile(rows)
			if err != nil {
				t.Errorf("Error from getADIFFile %v", err)
			}
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
