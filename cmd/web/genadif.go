package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"
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
		b.Write([]byte("<eor>\n\n"))
	}
	l := b.Len()
	p := make([]byte, l)
	b.Read(p)
	// fmt.Println(string(p))
	err := os.WriteFile(app.adifFile, p, 0644)
	if err != nil {
		return err
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
	userDef := "AD2CC stationmaster:github.com/Saied74/stationmaster"
	b.Write([]byte(fmt.Sprintf("<USERDEF1:%d:S>%s\n", len(userDef), userDef)))
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
