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
	FormData  *formData
	Speed     string
	FarnSpeed string
	Lsm       string
	Wsm       string
	Mode      string
	Top       headRow
	Table     []LogsRow
	LogEdit   *LogsRow
	Show      bool
	Edit      bool
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "home.page.html", nil)
}

//<++++++++++++++++++++++++++++  Logger  ++++++++++++++++++++++++++++++>

func (app *application) qsolog(w http.ResponseWriter, r *http.Request) {
	var err error
	app.td.Table, err = app.stationModel.getLatestLogs(displayLines)
	if err != nil {
		app.serverError(w, err)
	}
	app.td.Top = tableHead
	f := newForm(url.Values{})
	app.td.FormData = f
	app.td.Show = false
	app.td.Edit = false
	app.render(w, r, "log.page.html", app.td)
}

func (app *application) addlog(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
	}

	f := newForm(r.PostForm)

	f.required("call", "sent", "rcvd", "band")

	f.maxLength("call", 10)
	f.maxLength("sent", 3)
	f.maxLength("rcvd", 3)
	f.maxLength("band", 8)
	f.maxLength("name", 25)
	f.maxLength("country", 25)
	f.maxLength("comment", 75)
	f.maxLength("lotwrcvd", 10)
	f.maxLength("lotwsent", 10)

	f.minLength("sent", 2)
	f.minLength("rcvd", 2)

	f.isInt("sent")
	f.isInt("rcvd")

	if !f.valid() {
		var err error
		app.td.Table, err = app.stationModel.getLatestLogs(displayLines)
		if err != nil {
			app.serverError(w, err)
		}
		app.td.Top = tableHead

		app.td.FormData = f
		app.td.Show = true
		app.td.Edit = false
		app.render(w, r, "log.page.html", app.td)
		return
	}

	tr := LogsRow{}

	tr.Call = strings.ToUpper(r.PostForm.Get("call"))
	tr.Sent = r.PostForm.Get("sent")
	tr.Rcvd = r.PostForm.Get("rcvd")
	tr.Band = strings.ToLower(r.PostForm.Get("band"))
	tr.Name = r.PostForm.Get("name")
	tr.Country = r.PostForm.Get("country")
	tr.Comment = r.PostForm.Get("comment")
	tr.Lotwsent = r.PostForm.Get("lotwsent")
	tr.Lotwrcvd = r.PostForm.Get("lotwrcvd")

	mode := r.PostForm.Get("mode")
	if mode == "1" {
		tr.Mode = "USB"
	} else {
		tr.Mode = "CW"
	}

	_, err = app.stationModel.insertLog(&tr)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.td.Table, err = app.stationModel.getLatestLogs(displayLines)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.td.Top = tableHead
	app.td.Show = false
	app.td.Edit = false
	app.render(w, r, "log.page.html", app.td)
}

func (app *application) editlog(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	tr, err := app.stationModel.getLogByID(id)
	if err != nil {
		app.serverError(w, err)
	}

	app.td.LogEdit = tr
	f := newForm(url.Values{})
	app.td.FormData = f
	app.td.Show = true
	app.td.Edit = true
	app.render(w, r, "log.page.html", app.td)
}

func (app *application) updatedb(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
	}

	f := newForm(r.PostForm)

	f.required("call", "sent", "rcvd", "band")

	f.maxLength("call", 10)
	f.maxLength("sent", 3)
	f.maxLength("rcvd", 3)
	f.maxLength("band", 8)
	f.maxLength("name", 25)
	f.maxLength("country", 25)
	f.maxLength("comment", 75)
	f.maxLength("lotwrcvd", 10)
	f.maxLength("lotwsent", 10)

	f.minLength("sent", 2)
	f.minLength("rcvd", 2)

	f.isInt("sent")
	f.isInt("rcvd")

	if !f.valid() {
		var err error
		app.td.Table, err = app.stationModel.getLatestLogs(displayLines)
		if err != nil {
			app.serverError(w, err)
		}
		app.td.Top = tableHead

		app.td.FormData = f
		app.td.Show = true
		app.td.Edit = false
		app.render(w, r, "log.page.html", app.td)
		return
	}

	tr := LogsRow{}

	tr.Call = strings.ToUpper(r.PostForm.Get("call"))
	tr.Sent = r.PostForm.Get("sent")
	tr.Rcvd = r.PostForm.Get("rcvd")
	tr.Band = strings.ToLower(r.PostForm.Get("band"))
	tr.Name = r.PostForm.Get("name")
	tr.Country = r.PostForm.Get("country")
	tr.Comment = r.PostForm.Get("comment")
	tr.Lotwsent = r.PostForm.Get("lotwsent")
	tr.Lotwrcvd = r.PostForm.Get("lotwrcvd")

	mode := r.PostForm.Get("mode")
	if mode == "1" {
		tr.Mode = "USB"
	} else {
		tr.Mode = "CW"
	}

	err = app.stationModel.updateLog(&tr, app.td.LogEdit.Id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.td.Table, err = app.stationModel.getLatestLogs(displayLines)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.td.Top = tableHead
	app.td.Show = false
	app.td.Edit = false
	app.render(w, r, "log.page.html", app.td)
}

//<++++++++++++++++++++++++++  Antenna  ++++++++++++++++++++++++++++>

func (app *application) ant(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "ant.page.html", nil)
}

//<++++++++++++++++++++++++  Keyer - Tutor  ++++++++++++++++++++++++++>

func (app *application) ktutor(w http.ResponseWriter, r *http.Request) {
	data := getBaseTemp()
	f := newForm(url.Values{})
	data.FormData = f
	app.render(w, r, "ktutor.page.html", data)
}

func (app *application) start(w http.ResponseWriter, r *http.Request) {
	//variable naming convention:
	//all constants are lower case and longer
	//thier numeric version has float in front ot it
	//data extracted from the form ends in X (they are all strings)
	//when converted to float64, they are shortened

	data := getBaseTemp()
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
	}
	f := newForm(r.PostForm)

	s, speedX := f.extractFloat("speed", speed, floatSpeed)
	fs, fspeedX := f.extractFloat("farnspeed", farnspeed, floatFarn)
	wf, wsmX := f.extractFloat("wsm", wsm, floatWsm)
	lf, lsmX := f.extractFloat("lsm", lsm, floatLsm)

	modeX := r.PostForm.Get("mode") //tutor or keyer
	if modeX == "2" {
		f.Errors.add("mode", "Keyer feature not yet implemented")
	}
	//check to make sure keyer has stopped
	if app.ktrunning {
		f.Errors.add("ktrunning", "Keyer-tutor is running, stop it first")
	}

	if !f.valid() {
		data.FormData = f
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
		FormData:  f,
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
	f := newForm(url.Values{})
	data.FormData = f

	app.render(w, r, "ktutor.page.html", data)
}

func (app *application) stopcode(w http.ResponseWriter, r *http.Request) {
	data := getBaseTemp()
	app.render(w, r, "runkt.page.html", data)
}
