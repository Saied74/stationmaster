package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)


type spider struct {
	r *bufio.Reader
	w *bufio.Writer
}

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
	err = dx.logIn(c, app.call)
	if err != nil {
		return spider{}, err
	}
	return dx, nil
}

func (app *application) changeBand(band string) error {
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
	return nil
}

func (app *application) getSpider(band string, lineCnt int) ([]DXClusters, error) {
	b := make([]byte, 500)

	_, err := app.sp.w.WriteString(fmt.Sprintf("show/dx %d filter\n", lineCnt))
	if err != nil {
		if e, ok := err.(net.Error); ok && e.Timeout() {
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
	
	b = []byte{}

	for {
		bb, err := app.sp.r.ReadByte()
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Timeout() {
				return []DXClusters{}, errTimeout
			}
			return []DXClusters{}, err
		}
		b = append(b, bb)
		if strings.Contains(string(b), myCall) {
			break
		}
	}
	//fmt.Printf("Start Result\n%s\nEnd Result\n", string(b))
	dxData, err := lexResults(string(b))
	if err != nil {
		return []DXClusters{}, err
	}
	dxData, err = app.logsModel.findNeed(dxData)
	if err != nil {
		return dxData, nil
	}
	return dxData, nil
}

func (s *spider) logIn(c net.Conn, call string) error {
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


