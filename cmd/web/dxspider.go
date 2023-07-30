package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
//    "os"
	"strings"
	"time"
)

var filters = []string{
	"reject/spot 0 on hf/rtty",
	"reject/spot 1 on hf/sstv",
	"reject/spot 2 not by_state nj,ny,pa",
}

type spider struct {
	r *bufio.Reader
	w *bufio.Writer
}

func getSpider() {
	//program flags
	dxspider := flag.String("spider", "dxc.ww1r.com:7300", "dxspider server ip:port address")
	call := flag.String("call", "AD2CC", "your call sign")
	flag.Parse()

	b := make([]byte, 500)

	c, err := net.Dial("tcp", *dxspider)
	if err != nil {
		log.Fatal(err)
	}

	defer c.Close()

	dx := spider{
		r: bufio.NewReader(c),
		w: bufio.NewWriter(c),
	}
	err = dx.logIn(c, *call)
	if err != nil {
		log.Fatal(err)
	}

	_, err = dx.w.WriteString("accept/spot 4 on 20m \n")
	if err != nil {
		log.Fatal(err)
	}
	err = dx.w.Flush()
	if err != nil {
		log.Fatal(err)
	}

	b = []byte{}

	for {
		bb, err := dx.r.ReadByte()
		if err != nil || err == io.EOF {
			break
		}
		b = append(b, bb)
		if strings.Contains(string(b), "ad2cc") {
			break
		}
	}
	for i := 0; i < 20; i++ {
		_, err = dx.w.WriteString("show/dx 20 filter \n")
		if err != nil {
			log.Fatal(err)
		}
		err = dx.w.Flush()
		if err != nil {
			log.Fatal(err)
		}

		b = []byte{}

		for {
			bb, err := dx.r.ReadByte()
			if err != nil || err == io.EOF {
				break
			}
			b = append(b, bb)
			if strings.Contains(string(b), "ad2cc") {
				break
			}
		}
		//fmt.Printf("Start Result \n%s\nEnd Result\n", string(b))
       	lexResults(string(b))
        //os.Exit(1)
        time.Sleep(time.Duration(3)*time.Second)
	}
	_, err = dx.w.WriteString("bye\n")
	if err != nil {
		log.Fatal(err)
	}
	err = dx.w.Flush()
	if err != nil {
		log.Fatal(err)
	}
}

func (s *spider) logIn(c net.Conn, call string) error {
	var err error
	b := make([]byte, 2000)

	for {
		bb, err := s.r.ReadByte()
		if bb == 32 {
			b = append(b, bb)
			break
		}
		if err != nil || err == io.EOF {
			return err
		}
		b = append(b, bb)
	}
	fmt.Println(string(b))

	_, err = s.w.WriteString(call + "\n")
	if err != nil {
		log.Fatal(err)
	}
	err = s.w.Flush()
	if err != nil {
		return err
	}
	for {
		bb, err := s.r.ReadByte()
		if err != nil || err == io.EOF {
			break
		}
		b = append(b, bb)
		if strings.Contains(string(b), "ad2cc") {
			break
		}
	}
	fmt.Println(string(b))

	return nil
}

func (s *spider) setFilters() error {
	var err error

	b := make([]byte, 300)

	for _, filter := range filters {
		_, err = s.w.WriteString(filter + "\n")
		if err != nil {
			log.Fatal(err)
		}
		err = s.w.Flush()
		if err != nil {
			return err
		}
		time.Sleep(time.Duration(3000) * time.Millisecond)

		for {
			bb, err := s.r.ReadByte()
			if err != nil || err == io.EOF {
				break
			}
			b = append(b, bb)
			if strings.Contains(string(b), "ad2cc") {
				break
			}
		}
		fmt.Println(string(b))
	}
	return nil
}
