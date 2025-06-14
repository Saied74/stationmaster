package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"
	"unicode/utf8"
)

const (
	freqLen = 5
	modeLen = 2
	callLen = 13
	rstLen  = 3
	exchLen = 6
)

type contestData struct {
	filename    string
	name        string //contest name
	startTime   time.Time
	endTime     time.Time
	score       string
	fieldCount  int
	callWidth   int
	field1Width int
	field2Width int
	field3Width int
	field4Width int
	field5Width int
}

type cabBuffer []byte

var cabData = cabBuffer{}

func (c cabBuffer) Write(p []byte) (int, error) {
	for _, pp := range p {
		cabData = append(cabData, pp)
	}
	return len(p), nil
}

func (app *application) genCabrilloFile(rows []LogsRow, cd *contestData) error {
	cabData = cabBuffer{}
	dd := make(cabBuffer, 10)
	cd.score = ""
	tst, err := app.contestModel.getContest(cd.name)
	if err != nil {
		return err
	}
	header := writeCabrilloHeader(cabData, cd.name, cd.score)
	w := tabwriter.NewWriter(dd, 1, 2, 1, ' ', 0)
	for _, row := range rows {
		s := ""
		s1 := ""
		s += "QSO:\t"
		s += bandNormalize(row.Band) + "\t"
		s += row.Mode + "\t"
		dd, dt := cookTime(row.Time)
		s += dd + "\t"
		t := strings.Split(strings.TrimSuffix(dt, "Z"), ":")
		s += strings.Join(t[0:2], "") + "\t"
		s += myCall + "\t"
		s1 += row.Call + "\t"
		if tst.FieldCount < 3 {
			s += row.Field1Sent + "\t"
			s += row.Field2Sent + "\t"
			s1 += row.Field1Rcvd + "\t"
			s1 += row.Field2Rcvd + "\t"
		}
		if tst.FieldCount < 4 {
			s += row.Field3Sent + "\t"
			s1 += row.Field3Rcvd + "\t"
		}
		if tst.FieldCount < 5 {
			s += row.Field4Sent + "\t"
			s1 += row.Field4Rcvd + "\t"
		}
		if tst.FieldCount < 6 {
			s += row.Field5Sent + "\t"
			s1 += row.Field5Rcvd + "\t"
		}
		s += s1
		//		s += row.Sent + "\t"
		//		s += row.ExchSent + "\t"
		//		s += row.Call + "\t"
		//		s += row.Rcvd + "\t"
		//		s += row.ExchRcvd + "\t"
		fmt.Fprintln(w, s)
	}
	w.Flush()
	dd.Write([]byte("END-OF-LOG: \n"))
	fullData := append(header, cabData...)
	err = writeCab(cd.filename, []byte(fullData))
	if err != nil {
		return err
	}
	return nil

}

func (app *application) genNewCabrilloFile(rows []LogsRow, cd *contestData) error {
	b := new(bytes.Buffer)
	//cabData = cabBuffer{}
	//dd := make(cabBuffer, 10)
	cd.score = ""
	writeNewCabrilloHeader(b, cd.name, cd.score)
	//w := tabwriter.NewWriter(dd, 1, 2, 1, ' ', 0)
	for _, row := range rows {
		s := ""
		s += "QSO: "
		band := bandNormalize(row.Band)
		bandLen := utf8.RuneCountInString(band)
		if bandLen == 5 {
			s += band + " "
		} else {
			s += " " + band + " "
		}
		s += row.Mode + " "
		dd, dt := cookTime(row.Time)
		s += dd + " "
		t := strings.Split(strings.TrimSuffix(dt, "Z"), ":")
		s += strings.Join(t[0:2], "") + " "
		callGap := cd.callWidth - utf8.RuneCountInString(myCall)
		if callGap < 0 {
			return fmt.Errorf("caller call sign too wide by: %d", callGap)
		}
		s += myCall + strings.Repeat(" ", callGap+1)
		if cd.fieldCount >= 2 {
			gap1 := cd.field1Width - utf8.RuneCountInString(row.Field1Sent)
			if gap1 < 0 {
				return fmt.Errorf("field 1 gap is negative by: %d", gap1)
			}
			gap2 := cd.field2Width - utf8.RuneCountInString(row.Field2Sent)
			if gap2 < 0 {
				return fmt.Errorf("field 2 gap is negative by: %d", gap2)
			}
			s += row.Field1Sent + strings.Repeat(" ", gap1+1)
			s += row.Field2Sent + strings.Repeat(" ", gap2+1)
		}
		if cd.fieldCount >= 3 {
			gap3 := cd.field3Width - utf8.RuneCountInString(row.Field3Sent)
			if gap3 < 0 {
				return fmt.Errorf("field 3 gap is negative by: %d", gap3)
			}
			s += row.Field1Sent + strings.Repeat(" ", gap3+1)
		}
		if cd.fieldCount >= 4 {
			gap4 := cd.field4Width - utf8.RuneCountInString(row.Field4Sent)
			if gap4 < 0 {
				return fmt.Errorf("field 4 gap is negative by: %d", gap4)
			}
			s += row.Field1Sent + strings.Repeat(" ", gap4+1)
		}
		if cd.fieldCount == 5 {
			gap5 := cd.field5Width - utf8.RuneCountInString(row.Field5Sent)
			if gap5 < 0 {
				return fmt.Errorf("field 5 gap is negative by: %d", gap5)
			}
			s += row.Field1Sent + strings.Repeat(" ", gap5+1)
		}
		callGap = cd.callWidth - utf8.RuneCountInString(row.Call)
		if callGap < 0 {
			return fmt.Errorf("called call sign too wide by: %d", callGap)
		}
		s += row.Call + strings.Repeat(" ", callGap+1)
		if cd.fieldCount >= 2 {
			gap1 := cd.field1Width - utf8.RuneCountInString(row.Field1Rcvd)
			if gap1 < 0 {
				return fmt.Errorf("field 1 gap is negative by: %d", gap1)
			}
			gap2 := cd.field2Width - utf8.RuneCountInString(row.Field2Rcvd)
			if gap2 < 0 {
				return fmt.Errorf("field 2 gap is negative by: %d", gap2)
			}
			s += row.Field1Rcvd + strings.Repeat(" ", gap1+1)
			s += row.Field2Rcvd + strings.Repeat(" ", gap2+1)
		}
		if cd.fieldCount >= 3 {
			gap3 := cd.field3Width - utf8.RuneCountInString(row.Field3Rcvd)
			if gap3 < 0 {
				return fmt.Errorf("field 3 gap is negative by: %d", gap3)
			}
			s += row.Field1Rcvd + strings.Repeat(" ", gap3+1)
		}
		if cd.fieldCount >= 4 {
			gap4 := cd.field4Width - utf8.RuneCountInString(row.Field4Rcvd)
			if gap4 < 0 {
				return fmt.Errorf("field 4 gap is negative by: %d", gap4)
			}
			s += row.Field1Rcvd + strings.Repeat(" ", gap4+1)
		}
		if cd.fieldCount == 5 {
			gap5 := cd.field5Width - utf8.RuneCountInString(row.Field5Rcvd)
			if gap5 < 0 {
				return fmt.Errorf("field 5 gap is negative by: %d", gap5)
			}
			s += row.Field1Rcvd + strings.Repeat(" ", gap5+1)
		}
		s += "\n"

		b.WriteString(s)
	}

	b.WriteString("END-OF-LOG: \n")
	f, err := os.Create(cd.filename)
	defer f.Close()
	if err != nil {
		return err
	}
	bb := make([]byte, 10000)
	b.Read(bb)
	_, err = f.Write(bb)
	if err != nil {
		return err
	}

	return nil

}

func bandNormalize(band string) string {
	band = strings.ToUpper(band)
	switch band {
	case "10M":
		return "28000"
	case "15M":
		return "21000"
	case "20M":
		return "14000"
	case "40M":
		return "7000"
	case "80M":
		return "3500"
	}
	return band // lengthNormalize(band, freqLen) // TODO: check length and normalize
}

func modeNormalize(mode string) []byte {
	mode = strings.ToUpper(mode)
	switch mode {
	case "USB":
		return []byte("PH ")
	case "LSB":
		return []byte("PH ")
	}
	return lengthNormalize(mode, modeLen) //// TODO: check length and normamize
}

func modeNorm(mode string) string {
	mode = strings.ToUpper(mode)
	switch mode {
	case "USB":
		return "PH\t"
	case "LSB":
		return "PH\t"
	}
	return mode + "\t"
}

func lengthNormalize(x string, l int) []byte {
	if len(x) == l {
		return []byte(x + " ")
	}
	if len(x) > l {
		return []byte(x[0:l] + " ")
	}
	y := []byte(x)
	space := 0x20
	for i := 0; i < l-len(x); i++ {
		y = append(y, byte(space))
	}
	return append(y, byte(space))
}

func writeCabrilloHeader(b cabBuffer, contest, score string) cabBuffer {
	b.Write([]byte("START-OF-LOG: 3.0\n"))
	b.Write([]byte("CONTEST: " + contest + "\n"))
	b.Write([]byte("LOCATION: NJ\n"))
	b.Write([]byte("CALLSIGN: " + myCall + "\n"))
	b.Write([]byte("CATEGORY-OPERATOR: SINGLE-OP\n"))
	b.Write([]byte("CATEGORY-ASSISTED: NON-ASSISTED\n"))
	b.Write([]byte("CATEGORY-BAND: All\n"))
	b.Write([]byte("CATEGORY-POWER: LOW\n"))
	b.Write([]byte("CATEGORY-MODE: CW\n"))
	b.Write([]byte("CATEGORY-STATION: FIXED\n"))
	b.Write([]byte("CATEGORY-TRANSMITTER: ONE\n"))
	b.Write([]byte("CLAIMED-SCORE: " + score + "\n"))
	b.Write([]byte("CLUB: DVRA\n"))
	b.Write([]byte("NAME: Asadolah Seghatoleslami\n"))
	b.Write([]byte("ADDRESS: 11 Silvers Lane, Cranbury NJ 08512\n"))
	return b
}

func writeCab(filename string, c []byte) error {
	f, err := os.Create(filename)
	defer f.Close()
	if err != nil {
		return err
	}
	_, err = f.Write(c)
	if err != nil {
		return err
	}
	return nil
}

func writeNewCabrilloHeader(b *bytes.Buffer, contest, score string) {
	b.Write([]byte("START-OF-LOG: 3.0\n"))
	b.Write([]byte("CONTEST: " + contest + "\n"))
	b.Write([]byte("LOCATION: NJ\n"))
	b.Write([]byte("CALLSIGN: " + myCall + "\n"))
	b.Write([]byte("CATEGORY-OPERATOR: SINGLE-OP\n"))
	b.Write([]byte("CATEGORY-ASSISTED: NON-ASSISTED\n"))
	b.Write([]byte("CATEGORY-BAND: All\n"))
	b.Write([]byte("CATEGORY-POWER: LOW\n"))
	b.Write([]byte("CATEGORY-MODE: CW\n"))
	b.Write([]byte("CATEGORY-STATION: FIXED\n"))
	b.Write([]byte("CATEGORY-TRANSMITTER: ONE\n"))
	b.Write([]byte("CLAIMED-SCORE: " + score + "\n"))
	b.Write([]byte("CLUB: DVRA\n"))
	b.Write([]byte("NAME: Asadolah Seghatoleslami\n"))
	b.Write([]byte("ADDRESS: 11 Silvers Lane, Cranbury NJ 08512\n"))
}
