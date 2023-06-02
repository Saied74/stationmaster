package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime/debug"
	"strings"
	"time"

	"github.com/k0swe/wsjtx-go/v4"
)

const (
	wsjtBuffer = 5
	wsjtShortAverage = 3
	wsjtLongAverage = 5
	wsjtMargin = 2
)


func (app *application) wsjtxServe() {
	log.Println("Listening for WSJT-X...")
	wsjtxServer, err := wsjtx.MakeServer()
	if err != nil {
		serverError(err)
		return
	}
	wsjtxChannel := make(chan interface{}, 5)
	errChannel := make(chan error, 5)
	go wsjtxServer.ListenToWsjtx(wsjtxChannel, errChannel)
	//toggle := true
	for {
		select {
		case err := <-errChannel:
			log.Printf("error: %v", err)
		case message := <-wsjtxChannel:
			err := app.handleServerMessage(message)
			if err != nil {
				serverError(err)
				break
			}
		default:
			//time.Sleep(3000*time.Millisecond)
		}
		//t := time.Now().Second()
		//if t % 15 < wsjtMargin && toggle == true {
			//fmt.Printf("Count of 1 round = %d\t", app.cqStat[app.wsjtPntr])
			//short := app.countCQ(wsjtShortAverage)
			//long  := app.countCQ(wsjtLongAverage)
			//fmt.Printf("\n")
			
			//if app.qsoStat[app.wsjtPntr] != 0 {
				//r := float64(app.cqStat[app.wsjtPntr]) / float64(app.qsoStat[app.wsjtPntr])
				//fmt.Printf("Prcnt of 1 round = %0.2f\t", r)
			//}
			//app.averageCQ(short, wsjtShortAverage)
			//app.averageCQ(long, wsjtLongAverage)
			
			//app.wsjtPntr++
			//if app.wsjtPntr >= wsjtBuffer {
				//app.wsjtPntr = 0
			//}
			//app.cqStat[app.wsjtPntr] = 0
			//app.qsoStat[app.wsjtPntr] = 0
			//fmt.Printf("\n\n")
			//toggle = false
		//}
		//if t % 15 > 10 {
			//toggle = true
		//}
	}
}

func (app * application)countCQ(n int) int{
	p := app.wsjtPntr
	cnt := 0
	for i := 0; i < n; i++ {
		cnt += app.cqStat[p]
		p--
		if p < 0 {
			p = wsjtBuffer - 1
		}
	}
	fmt.Printf("Count of %d rounds = %d\t", n, cnt)
	return cnt
}

func (app * application)averageCQ(l, n int) {
	p := app.wsjtPntr
	cnt := 0
	for i := 0; i < n; i++ {
		cnt += app.qsoStat[p]
		p--
		if p < 0 {
			p = wsjtBuffer - 1
		}
	}
	if cnt != 0 {
		r := float64(l) / float64(cnt)
		fmt.Printf("Prcnt of %d rounds = %0.2f\t", n, r)
	}
}
	


func (app *application)handleServerMessage(message interface{}) error {
	switch message.(type) {
	case wsjtx.HeartbeatMessage:
		//log.Println("Heartbeat:", message)
	case wsjtx.StatusMessage:
		//log.Println("Status:", message)
	case wsjtx.DecodeMessage:
		m := message.(wsjtx.DecodeMessage)
		mm := strings.TrimSpace(m.Message)
		msg := strings.Split(mm, " ")
		if len(msg) < 2 {
			app.infoLog.Printf("wsjt message did not have two components %s, %v\n", mm, msg)
		}
		for i, mm := range msg {
			msg[i] = strings.TrimSpace(mm)
		}
		t := m.Time
		now := time.Now()
		now = now.UTC()
		//fmt.Println("Hour: ", now.Hour())
		t = t - (uint32(now.Hour()) * (60 * 60 *1000))
		//fmt.Println("Minute: ", now.Minute())
		t = t - (uint32(now.Minute()) * (60 * 1000))
		t = t / 1000
		if msg[0] == "CQ" {
			app.cqStat[app.wsjtPntr]++
		}
		
	case wsjtx.ClearMessage:
		//log.Println("Clear:", message)
	case wsjtx.QsoLoggedMessage:
		err := app.logQSO(message)
		if err != nil {
			return err
		}
		//log.Println("QSO Logged:", message)
	case wsjtx.CloseMessage:
		//log.Println("Close:", message)
	case wsjtx.WSPRDecodeMessage:
		//log.Println("WSPR Decode:", message)
	case wsjtx.LoggedAdifMessage:
		//log.Println("Logged Adif:", message)
	default:
		log.Println("Other message type:", reflect.TypeOf(message), message)
	}
	app.qsoStat[app.wsjtPntr]++
	return nil
}

func (app *application)logQSO(message interface{}) error {
	m := message.(wsjtx.QsoLoggedMessage)
	
	band, err := app.otherModel.getDefault("band")
	if err != nil {
		return err
	}
	mode, err := app.otherModel.getDefault("mode")
	if err != nil {
		return err
	}
	
	call := m.DxCall
	
	c, err := app.qrzModel.getQRZ(call)
	if err != nil {
		if errors.Is(err, errNoRecord) {
			q, err := app.getHamInfo(call)
			if err != nil {
				return err
			}
			c = &q.Callsign
			c.QSOCount = 1
			err = app.qrzModel.insertQRZ(c)
			if err != nil {
				return err
			}
			
			lr := LogsRow{
				Call:  m.DxCall,
				Sent: m.ReportSent,
				Rcvd: m.ReportReceived,
				Band: band,
				Mode: mode,
				Name: fmt.Sprintf("%s %s", c.Fname, c.Lname),
				Country: c.Country,
				Comment: m.DxGrid,
				ExchSent: m.ExchangeSent,
				ExchRcvd: m.ExchangeReceived,
			}
			_, err = app.logsModel.insertLog(&lr)
			if err != nil {
				return err
			}
			return nil
			//This is the case that this is the first contact
		}
		return err
	}
	//this is the case that this the second of more contacts
	t, err := app.logsModel.getLogsByCall(call)
	if err != nil {
		return err
	}
	err = app.qrzModel.updateQSOCount(call, len(t)+1)
	if err != nil {
		return err
	}
	
	lr := LogsRow{
		Call:  m.DxCall,
		Sent: m.ReportSent,
		Rcvd: m.ReportReceived,
		Band: band,
		Mode: mode,
		Name: fmt.Sprintf("%s %s", c.Fname, c.Lname),
		Country: c.Country,
		Comment: m.DxGrid,
		ExchSent: m.ExchangeSent,
		ExchRcvd: m.ExchangeReceived,
	}
	_, err = app.logsModel.insertLog(&lr)
	if err != nil {
		return err
	}
	return nil
}

func serverError(err error) {
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.LUTC|log.Llongfile)
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	errorLog.Output(2, trace)
}

//<<==================== wsjt decodeMessage handler ==================>>

//type wsjtMsgType int

//type wsjtMsgItemType struct {
	//typ wsjtMsgType
	//val string
//}

//type wsjtMsgLexer struct {
	//name  string    // used only for error reports.
	//input string    // the string being scanned.
	//start int       // start position of this item.
	//pos   int       // current position in the input.
	//width int       // width of last rune read from input.
	//field wsjtMsgItemType  //field to emit into the chanel
	//items chan wsjtMsgItemType // channel of scanned items.
//}

//type wsjtMsgStateFn func(*wsjtMsgLexer) wsjtMsgStateFn

//const (
	//wsjtItemCQ wsjtMsgType = iota
	//wsjtItemDE
	//wsjtItemDX
	//wsjtItemEOR
	//wsjtItemError
//)

//type wsjtDecodeMsgType {
	//cq bool
	//dx string
	//de string
	//t  time.Time
//}
	

//unc (app *application) processDecodeMsg(m string) (wsjtDecodeMsgType, error) {
	//msg := wsjtDecodeMsgType{
		//t: time.Now(),
		//}

	//_, c := lex("wsjtDecodeMessage", m)
	
	//for {
		//d := <-c
		//switch d.typ {
		//case wsjtItemEOR:
			//return msg, nil
		//case wsjtItemCQ:
			//msg.cq = true
		//case wsjtItemDE:
			//msg.de = d.val
		//case wsjtItemDX:
			//msg.dx = d.val
		//case itemError:
			//return msg{}, fmt.Errorf("error from the decode message parser channel")
		//default:
			//return msg{}, fmt.Errorf("invalid type received from message parser channel %s", m)
		//}
	//}
	//return msg, fmt.Errorf("got to the end of processDecodeMsg %s", m)
//}

//func wsjtMsgLex(name, input string) (*wsjtMsgLexer, chan wsjtMsgItemType) {
	//l := &wsjtMsgLexer{
		//name:  name,
		//input: input,
		//items: make(chan item),
	//}
	//go l.run() // Concurrently run state machine.
	//return l, l.items
//}

//// run lexes the input by executing state functions until
//// the state is nil.
//func (l *wsjtMsgLexer) run() {
	//for state := wsjtMsgLexSpace; state != nil; {
		//state = state(l)
	//}
	//close(l.items) // No more tokens will be delivered.
//}

//func (l *wsjtMsgLexer) emit() {
	//l.items <- wsjtMsgItemType{l.field, l.input[l.start:l.pos]}
	//l.start = l.pos
//}

//// next returns the next rune in the input.
//func (l *wsjtMsgLexer) next() (r rune) {
	//if l.pos >= len(l.input) {
		//l.width = 0
		//return eof
	//}
	//r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	//l.pos += l.width
	//return r
//}

//func wsjtLexSpace(l *wsjtMsgLexer) stateFn {
	//for {
		//if !strings.HasPrefix(l.input[l.pos:], " ") {
			//l.start = l.pos
			//return wsjtLexCQ // Next state.
		//}
		//if l.next() == eof {
			//break
		//}
	//}
	//// Correctly reached EOF.
	//l.field = wsjtItemEOR
	//l.emit()   // Useful to make EOF a token.
	//return nil // Stop the run loop.
//}

//func wsjtLexCQ(l *wsjtMsgLexer) stateFn {
	//for {
		//if strings.HasPrefix(l.input[l.pos:], "CQ") {
			//l.start = l.pos
			//return wsjtLexCQ // Next state.
		//}
		//if l.next() == eof {
			//break
		//}
	//}
	//// Correctly reached EOF.
	//l.field = wsjtItemEOR
	//l.emit()   // Useful to make EOF a token.
	//return nil // Stop the run loop.
//}
