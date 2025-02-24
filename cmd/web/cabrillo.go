package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

const (
	freqLen = 5
	modeLen = 2
	callLen = 13
	rstLen  = 3
	exchLen = 6
)

type contestData struct {
	filename   string
	name       string //contest name
	startTime  time.Time
	endTime    time.Time
	score      string
	fieldCount int
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
	cabData = cabBuffer{}
	dd := make(cabBuffer, 10)
	cd.score = ""
	header := writeCabrilloHeader(cabData, cd.name, cd.score)
	w := tabwriter.NewWriter(dd, 1, 2, 1, ' ', 0)
	for _, row := range rows {
		s := ""
		s += "QSO:\t"
		s += bandNormalize(row.Band) + "\t"
		s += row.Mode + "\t"
		dd, dt := cookTime(row.Time)
		s += dd + "\t"
		t := strings.Split(strings.TrimSuffix(dt, "Z"), ":")
		s += strings.Join(t[0:2], "") + "\t"
		s += myCall + "\t"
		if cd.fieldCount >= 2 {
			s += row.Field1Sent + "\t"
			s += row.Field2Sent + "\t"
		}
		if cd.fieldCount >= 3 {
			s += row.Field3Sent + "\t"
		}
		if cd.fieldCount >= 4 {
			s += row.Field4Sent + "\t"
		}
		if cd.fieldCount == 5 {
			s += row.Field5Sent + "\t"
		}
		s += row.Call + "\t"
		if cd.fieldCount >= 2 {
			s += row.Field1Rcvd + "\t"
			s += row.Field2Rcvd + "\t"
		}
		if cd.fieldCount >= 3 {
			s += row.Field3Rcvd + "\t"
		}
		if cd.fieldCount >= 4 {
			s += row.Field4Rcvd + "\t"
		}
		if cd.fieldCount == 5 {
			s += row.Field5Rcvd + "\t"
		}

		fmt.Fprintln(w, s)
	}
	w.Flush()
	dd.Write([]byte("END-OF-LOG: \n"))
	fullData := append(header, cabData...)
	err := writeCab(cd.filename, []byte(fullData))
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
