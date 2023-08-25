package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
//	"os"
	"strings"
//	"text/tabwriter"
	"time"
)


type spider struct {
	r *bufio.Reader
	w *bufio.Writer
}

const (
	lineLength = 74
	msgLength = lineLength * dxLines
	disconnect = "disconnected"
)

var errNoDXSpots = errors.New("no dx spots")
var errTimeout = errors.New("dx spider timeout error")

func (app *application) initSpider() (spider, error) {
	
	dlr := net.Dialer{
		Timeout: time.Duration(2)*time.Second,
	}

	c, err := dlr.Dial("tcp", app.dxspider)
	if err != nil {
		if e, ok := err.(net.Error); ok && e.Timeout() {
			return spider{}, errTimeout
		}
		return spider{}, err
	}

	dx := spider{
		r: bufio.NewReader(c),
		w: bufio.NewWriter(c),
	}
	err = dx.logIn(/*c, */app.call)
	if err != nil {
		return spider{}, err
	}
	return dx, nil
}

func (app *application) changeBand(band string) error {
	if band == "WWV" || band == "AUX" {
		return nil
	}
	b := make([]byte, 500)

	_, err := app.sp.w.WriteString(fmt.Sprintf("accept/spot 4 on %s\n", band))
	if err != nil {
		if e, ok := err.(net.Error); ok && e.Timeout() {
			return errTimeout
		}
		return err
	}
	err = app.sp.w.Flush()
	if err != nil {
		if e, ok := err.(net.Error); ok && e.Timeout() {
			return errTimeout
		}
		return err
	}

	b = []byte{}

	for {
		bb, err := app.sp.r.ReadByte()
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Timeout() {
				return errTimeout
			}
			return err
		}
		b = append(b, bb)
		if strings.Contains(string(b), myCall) { //to do: fix this
			break
		}
	}
	//app.infoLog.Printf("%s", string(b))
	return nil
}

func (app *application) getSpider(band string, lineCnt int) ([]DXClusters, error) {
	b := make([]byte, 500)
    
	_, err := app.sp.w.WriteString(fmt.Sprintf("show/dx %d filter\n", lineCnt))
	if err != nil {
		if e, ok := err.(net.Error); ok && e.Timeout() {
			fmt.Println("Timed Out")
			return []DXClusters{}, errTimeout
		}
		return []DXClusters{}, err
	}
	err = app.sp.w.Flush()
	if err != nil {
		if e, ok := err.(net.Error); ok && e.Timeout() {
			return []DXClusters{}, errTimeout
		}
		return []DXClusters{}, err
	}
	
	var sB string
	m := 0
	for {
		m++
		if m >= 4095 {
			break
		}
		bb, err := app.sp.r.ReadByte()
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Timeout() {
				return []DXClusters{}, errTimeout
			}
			return []DXClusters{}, err
		}
		b = append(b, bb)
		sB = string(b)

		if strings.HasSuffix(sB, myCall) && len(sB) >= msgLength  {
			break
		}
		if strings.Contains(sB, disconnect) {
			app.sp.logIn(app.call)
			
		}
	}
	
	var splitLines = []string{}
	lines := strings.Split(sB, "\n")
	
	for _, line := range lines {
		if len(line) > 70 {
			splitLines = append(splitLines, strings.TrimSpace(line))
		}
	}
	var n int
	var dx = []DXClusters{}
	var l = DXClusters{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		n = strings.Index(line, " ")
		if n == -1 {
			continue
		}
		l.Frequency = line[:n]
		line = strings.TrimPrefix(line, l.Frequency)
		line = strings.TrimSpace(line)
		n = strings.Index(line, " ")
		if n == -1 {
			continue
		}
		l.DXStation = line[:n]
		line = strings.TrimPrefix(line, l.DXStation)
		line = strings.TrimSpace(line)
		n = strings.Index(line, " ")
		if n  == -1 {
			continue
		}
		l.Date = line[:n]
		line = strings.TrimPrefix(line, l.Date)
		line = strings.TrimSpace(line)
		n = strings.Index(line, " ")
		if n == -1 {
			continue
		}
		l.Time = line[:n]
		line = strings.TrimPrefix(line, l.Time)
		line = strings.TrimSpace(line)
		n = strings.LastIndex(line, "<")
		if n == -1 {
			continue
		}
		l.Info = line[:n]
		line = strings.TrimPrefix(line, l.Info)
		n = strings.LastIndex(line, ">")
		if n == -1 {
			continue
		}
		l.DE = line[:n]
		l.DE = strings.TrimPrefix(l.DE, "<")
		dx = append(dx, l)
	}
	//app.infoLog.Printf("\nReturn:\n%s\n", sB)
	//fmt.Printf("\n")
	//w := new(tabwriter.Writer)
	//w.Init(os.Stdout, 8, 8, 0, '\t', 0)
	//fmt.Fprintf(w, "DX Call\tFrequency\tDate\tTime\tInfo\tDE Call\n")
	//for _, line := range dx {
		//fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", line.DXStation, line.Frequency, line.Date, line.Time, line.Info, line.DE)
	//}
	//w.Flush()
	return dx, nil
}

func (s *spider) logIn(/*c net.Conn, */call string) error {
	var err error
	b := make([]byte, 2000)

	for {
		bb, err := s.r.ReadByte()
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Timeout() {
				return errTimeout
			}
			return err
		}
		if bb == 32 {
			b = append(b, bb)
			break
		}
		b = append(b, bb)
	}
	fmt.Println(string(b))

	_, err = s.w.WriteString(call + "\n")
	if err != nil {
		if e, ok := err.(net.Error); ok && e.Timeout() {
			return errTimeout
		}
		return err
	}
	err = s.w.Flush()
	if err != nil {
		if e, ok := err.(net.Error); ok && e.Timeout() {
			return errTimeout
		}
		return err
	}
	for {
		bb, err := s.r.ReadByte()
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Timeout() {
				return errTimeout
			}
			return err
		}
		b = append(b, bb)
		if strings.Contains(string(b), "ad2cc") {
			break
		}
	}
	fmt.Println(string(b))

	return nil
}


