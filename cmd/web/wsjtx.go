package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime/debug"
//	"time"

	"github.com/k0swe/wsjtx-go/v4"
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
	}
}


func (app *application)handleServerMessage(message interface{}) error {
	switch message.(type) {
	case wsjtx.HeartbeatMessage:
		//log.Println("Heartbeat:", message)
	case wsjtx.StatusMessage:
		//log.Println("Status:", message)
	case wsjtx.DecodeMessage:
		//m := message.(wsjtx.DecodeMessage)
		//log.Println("ID:", m.Id)
		//log.Println("New: ", m.New)
		//log.Println("Time: ", m.Time)
		//log.Println("SNR: ", m.Snr)
		//log.Println("DT: ", m.DeltaTimeSec)
		//log.Println("DF: ", m.DeltaFrequencyHz)
		//log.Println("Mode: ", m.Mode)
		//log.Println("Message: ", m.Message)
		//log.Println("Low Conf: ", m.LowConfidence)
		//log.Println("OffAir: ", m.OffAir)
		//log.Println("Decode:", message)
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
