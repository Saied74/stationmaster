package main

import (
	"bytes"
	"os"
	"strings"
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
	filename  string
	name      string
	startTime time.Time
	endTime   time.Time
	score     string
}

func (app *application) genCabrilloFile(rows []LogsRow, cd *contestData) error {
	var b bytes.Buffer
	b = writeCabrilloHeader(b, cd.name, cd.score)
	b.Write([]byte("                              --------info sent------- -------info rcvd--------\n"))
	b.Write([]byte("QSO:  freq mo date       time call          rst exch   call          rst exch   t\n"))
	b.Write([]byte("QSO: ***** ** yyyy-mm-dd nnnn ************* nnn ****** ************* nnn ****** n\n"))
	for _, row := range rows {
		b.Write([]byte("QSO: "))
		b.Write(bandNormalize(row.Band))
		b.Write(modeNormalize(row.Mode))
		dd, dt := cookTime(row.Time)
		b.Write([]byte(dd + " "))
		t := strings.Split(strings.TrimSuffix(dt, "Z"), ":")
		b.Write([]byte(strings.Join(t[0:2], "") + " "))
		b.Write(lengthNormalize("AD2CC", callLen))
		b.Write(lengthNormalize(row.Sent, rstLen))
		b.Write(lengthNormalize(row.ExchSent, exchLen))
		b.Write(lengthNormalize(row.Call, callLen))
		b.Write(lengthNormalize(row.Rcvd, rstLen))
		b.Write(lengthNormalize(row.ExchRcvd, exchLen))
		b.Write([]byte("o\n"))
	}
	b.Write([]byte("START-OF-LOG: \n"))
	p := make([]byte, b.Len())
	b.Read(p)
	err := writeCab(cd.filename, p)
	if err != nil {
		return err
	}
	return nil
}

func bandNormalize(band string) []byte {
	band = strings.ToUpper(band)
	switch band {
	case "10M":
		return []byte("21000 ")
	case "20M":
		return []byte("14000 ")
	case "40M":
		return []byte(" 7000 ")
	case "80M":
		return []byte(" 3500 ")
	}
	return lengthNormalize(band, freqLen) // TODO: check length and normalize
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

func writeCabrilloHeader(b bytes.Buffer, contest, score string) bytes.Buffer {
	b.Write([]byte("START-OF-LOG: 3.0\n"))
	b.Write([]byte("CONTEST: " + contest + "\n"))
	b.Write([]byte("LOCATION: NJ\n"))
	b.Write([]byte("CALLSIGN: AD2CC\n"))
	b.Write([]byte("CATEGORY-OPERATOR: SINGLE-OP\n"))
	b.Write([]byte("CATEGORY-ASSISTED: NON-ASSISTED\n"))
	b.Write([]byte("CATEGORY-OVERLAY: ROOKIE\n"))
	b.Write([]byte("CATEGORY-BAND: 20M\n"))
	b.Write([]byte("CATEGORY-POWER: LOW\n"))
	b.Write([]byte("CATEGORY-MODE: SSB\n"))
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
