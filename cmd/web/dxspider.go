package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
//    "os"
	"strings"
//	"time"
)

//var filters = []string{
	//"reject/spot 0 on hf/rtty",
	//"reject/spot 1 on hf/sstv",
	//"reject/spot 2 not by_state nj,ny,pa",
//}

type spider struct {
	r *bufio.Reader
	w *bufio.Writer
}

func (app *application)initSpider() (spider, error){

	c, err := net.Dial("tcp", app.dxspider)
	if err != nil {
		return spider{}, err
	}
	//defer c.Close()

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

func (app * application)getSpider(band string, lineCnt int) ([]DXClusters, error) {
	b := make([]byte, 500)
	
	_, err := app.sp.w.WriteString(fmt.Sprintf("accept/spot 4 on %s \n", band))
	if err != nil {
		return []DXClusters{}, err
	}
	err = app.sp.w.Flush()
	if err != nil {
		return []DXClusters{}, err
	}

	b = []byte{}

	for {
		bb, err := app.sp.r.ReadByte()
		if err != nil || err == io.EOF {
			break
		}
		b = append(b, bb)
		if strings.Contains(string(b), "ad2cc") { //to do: fix this
			break
		}
	}
		_, err = app.sp.w.WriteString(fmt.Sprintf("show/dx %d filter \n", lineCnt))
		if err != nil {
			return []DXClusters{}, err
		}
		err = app.sp.w.Flush()
		if err != nil {
			return []DXClusters{}, err
		}

		b = []byte{}

		for {
			bb, err := app.sp.r.ReadByte()
			if err != nil || err == io.EOF {
				break
			}
			b = append(b, bb)
			if strings.Contains(string(b), "ad2cc") { //to do: fix this
				break
			}
		}
		//fmt.Printf("Start Result \n%s\nEnd Result\n", string(b))
       	dxData, err := lexResults(string(b))
       	if err != nil {
			return []DXClusters{}, err
		}
        //os.Exit(1)
	return dxData, nil
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

//func (s *spider) setFilters() error {
	//var err error

	//b := make([]byte, 300)

	//for _, filter := range filters {
		//_, err = s.w.WriteString(filter + "\n")
		//if err != nil {
			//log.Fatal(err)
		//}
		//err = s.w.Flush()
		//if err != nil {
			//return err
		//}
		//time.Sleep(time.Duration(3000) * time.Millisecond)

		//for {
			//bb, err := s.r.ReadByte()
			//if err != nil || err == io.EOF {
				//break
			//}
			//b = append(b, bb)
			//if strings.Contains(string(b), "ad2cc") {
				//break
			//}
		//}
		//fmt.Println(string(b))
	//}
	//return nil
//}
