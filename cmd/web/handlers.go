package main

import (
	"context"
	"gobot.io/x/gobot/platforms/raspi"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

//seed data for the keyer - tutor
const (
	speed        = "15"
	floatSpeed   = 15.0
	farnspeed    = "18"
	floatFarn    = 18.0
	lsm          = "1.2"
	floatLsm     = 1.55
	wsm          = "1.3"
	floatWsm     = 1.3
	displayLines = 10
)

//for feeding dynamic data and error reports to templates
type templateData struct {
	FormData   url.Values
	FormErrors map[string]string
	Speed      string
	FarnSpeed  string
	Lsm        string
	Wsm        string
	Mode       string
	Top        LogsRow
	Table      []LogsRow
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "home.page.html", nil)
}

func (app *application) ktutor(w http.ResponseWriter, r *http.Request) {
	data := getBaseTemp()
	app.render(w, r, "ktutor.page.html", data)
}

func (app *application) qsolog(w http.ResponseWriter, r *http.Request) {
	var err error
	app.td.Table, err = app.stationModel.getLatestLogs(displayLines)
	if err != nil {
		app.serverError(w, err)
	}
	app.render(w, r, "log.page.html", app.td)
}

func (app *application) addlog(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
	}
	tr := LogsRow{}

	tr.Call = r.PostForm.Get("call")
	tr.Sent = r.PostForm.Get("sent")
	tr.Rcvd = r.PostForm.Get("rcvd")
	tr.Band = r.PostForm.Get("band")
	tr.Name = r.PostForm.Get("name")
	tr.Country = r.PostForm.Get("country")
	tr.Comment = ""
	tr.Lotwsent = ""
	tr.Lotwrcvd = ""

	mode := r.PostForm.Get("mode")
	if mode == "1" {
		tr.Mode = "USB"
	} else {
		tr.Mode = "CW"
	}

	_, err = app.stationModel.insertLog(&tr)
	if err != nil {
		app.serverError(w, err)
	}

	app.td.Table, err = app.stationModel.getLatestLogs(displayLines)
	if err != nil {
		app.serverError(w, err)
	}

	app.render(w, r, "log.page.html", app.td)
}

func (app *application) ant(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "ant.page.html", nil)
}

func (app *application) start(w http.ResponseWriter, r *http.Request) {
	//variable naming convention:
	//all constants are lower case and longer
	//thier numeric version has float in front ot it
	//data extracted from the form ends in X (they are all strings)
	//when converted to float64, they are shortened
	var err error
	var lf, wf, s, fs float64
	data := getBaseTemp()
	err = r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
	}

	speedX := r.PostForm.Get("speed")
	fspeedX := r.PostForm.Get("farnspeed")
	lsmX := r.PostForm.Get("lsm")   //letter spacing margin
	wsmX := r.PostForm.Get("wsm")   //word spacing margin
	modeX := r.PostForm.Get("mode") //tutor or keyer

	//to do:  will build a scalable validation solution next
	errors := make(map[string]string)

	if strings.TrimSpace(speedX) == "" {
		s = floatSpeed
		speedX = speed
	} else {
		s, err = strconv.ParseFloat(speedX, 64)
		if err != nil {
			errors["speed"] = "Sending speed must be a number"
		}
	}

	if strings.TrimSpace(fspeedX) == "" {
		fs = floatFarn
		fspeedX = farnspeed
	} else {
		fs, err = strconv.ParseFloat(fspeedX, 64)
		if err != nil {
			errors["farnspeed"] = "Farnsworth speed must be a number"
		}
	}

	if strings.TrimSpace(wsmX) == "" {
		wf = floatWsm
		wsmX = wsm
	} else {
		wf, err = strconv.ParseFloat(wsmX, 64)
		if err != nil {
			errors["wsm"] = "Word spacing margin must be a number"
		}
	}

	if strings.TrimSpace(lsmX) == "" {
		lf = floatLsm
		lsmX = lsm
	} else {
		lf, err = strconv.ParseFloat(lsmX, 64)
		if err != nil {
			errors["lsm"] = "Letter spacing margin must be a number"
		}
	}

	if modeX == "2" {
		errors["mode"] = "Keyer feature not yet implemented"
	}
	//check to make sure keyer has stopped
	if app.ktrunning {
		errors["ktrunning"] = "Keyer-tutor is running, stop it first"
	}

	if len(errors) > 0 {
		data.FormData = r.PostForm
		data.FormErrors = errors
		app.render(w, r, "ktutor.page.html", data)
		return
	}
	//get context with cancel so the keyer can be stopped when needed
	//note: cancel and ktrunning make this application stateful
	app.ctx, app.cancel = context.WithCancel(context.Background())
	cw := &cwDriver{
		dit:       raspi.NewAdaptor(),
		speed:     s,
		farnspeed: fs,
		lF:        lf,
		wF:        wf,
	}
	//so the keyer can't be run twice without stopping it first
	app.ktrunning = true
	go cw.work(app.ctx)

	data = &templateData{
		Speed:     speedX,
		FarnSpeed: fspeedX,
		Lsm:       lsmX,
		Wsm:       wsmX,
	}
	switch modeX {
	case "1":
		data.Mode = "Tutor"
	case "2":
		data.Mode = "keyer"
	default:
		data.Mode = "tutor"
	}
	app.render(w, r, "runkt.page.html", data)

}

func (app *application) stop(w http.ResponseWriter, r *http.Request) {
	app.cancel()
	app.ktrunning = false
	//to do: it would be better if the last data the user inputted was
	//used here.
	data := getBaseTemp()
	app.render(w, r, "ktutor.page.html", data)
}

func (app *application) stopcode(w http.ResponseWriter, r *http.Request) {
	data := getBaseTemp()
	app.render(w, r, "runkt.page.html", data)
}
