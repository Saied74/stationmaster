package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"
	"syscall"
	"time"
)

type spider struct {
	r *bufio.Reader
	w *bufio.Writer
}

const (
	lineLength = 74
	msgLength  = lineLength * dxLines
	disconnect = "disconnected"
)

var errNoDXSpots = errors.New("no dx spots")
var errTimeout = errors.New("dx spider timeout error")
var errDisconnect = errors.New("dx spider disconnected")

func (app *application) initSpider() (spider, error) {

	dlr := net.Dialer{
		Timeout: time.Duration(2) * time.Second,
	}

	c, err := dlr.Dial("tcp", app.dxspider)
	if err != nil {
		return spider{}, err
	}

	dx := spider{
		r: bufio.NewReader(c),
		w: bufio.NewWriter(c),
	}
	//fmt.Println(app.call)
	err = dx.logIn( /*c, */ app.call)
	if err != nil {
		return spider{}, err
	}
	return dx, nil
}

func (app *application) changeBand(band string) error {
	if band == "WWV" || band == "AUX" || band == "160m" {
		return nil
	}
	b := make([]byte, 500)

	_, err := app.sp.w.WriteString(fmt.Sprintf("accept/spot 4 on %s\n", band))
	if err != nil {
		return err
	}
	err = app.sp.w.Flush()
	if err != nil {
		return err
	}
	b = []byte{}

	for {
		bb, err := app.sp.r.ReadByte()
		if err != nil {
			return err
		}
		b = append(b, bb)
		if strings.Contains(string(b), myCall) {
			break
		}
	}
	return nil
}

func (app *application) getSpider(band string, lineCnt int) ([]DXClusters, error) {
	b := make([]byte, 500)
	if app.sp.w == nil {
		return []DXClusters{}, nil
	}
	_, err := app.sp.w.WriteString(fmt.Sprintf("show/dx %d filter\n", lineCnt))
	if err != nil {
		return []DXClusters{}, err
	}
	err = app.sp.w.Flush()
	if err != nil {
		return []DXClusters{}, err
	}
	var sB string
	m := 0
	for {
		m++
		if m == 2048 {
			break
		}
		bb, err := app.sp.r.ReadByte()
		if err != nil {
			return []DXClusters{}, err
		}
		b = append(b, bb)
		sB = string(b)

		if strings.HasSuffix(sB, myCall) && len(sB) >= msgLength {
			break
		}
		if strings.Contains(sB, disconnect) {
			return []DXClusters{}, errDisconnect
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
		if n == -1 {
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

	return dx, nil
}

func (s *spider) logIn(call string) error {
	var err error
	b := make([]byte, 2000)

	for {
		bb, err := s.r.ReadByte()
		if err != nil {
			return err
		}
		if bb == 32 {
			b = append(b, bb)
			break
		}
		b = append(b, bb)
	}
	//fmt.Println(string(b))
	//fmt.Println(string(call))
	_, err = s.w.WriteString(call + "\n")
	if err != nil {
		return err
	}
	err = s.w.Flush()
	if err != nil {
		return err
	}
	//fmt.Println("logged in")
	for {
		bb, err := s.r.ReadByte()
		if err != nil {
			return err
		}
		b = append(b, bb)
		if strings.Contains(string(b), myCall) {
			break
		}
	}
	//fmt.Println(string(b))

	return nil
}

func (app *application) spiderError(err error) error {
	if errors.Is(err, errTimeout) {
		app.infoLog.Printf("timeout error from calling getSpider in updateDX %v\n", err)
		err = app.sp.logIn(app.call)
		if err != nil {
			return err
		}
	}
	if errors.Is(err, errDisconnect) {
		err = app.sp.logIn(app.call)
		if err != nil {
			return err
		}
	}
	if errors.Is(err, syscall.EPIPE) {
		sp, err := app.initSpider()
		fmt.Println("EPIPE: ", err)
		if err != nil {
			return err
		}
		app.sp = sp
		return nil
	}
	if errors.Is(err, syscall.ECONNRESET) {
		err = app.sp.logIn(app.call)
		if err != nil {
			return err
		}
	}
	if err != nil {
		fmt.Errorf("error from calling spiderError %v\n", err)
		return err
	}
	return nil
}

func (app *application) byeSpider() error {
	_, err := app.sp.w.WriteString("bye\n")
	if err != nil {
		return err
	}
	return nil
}
