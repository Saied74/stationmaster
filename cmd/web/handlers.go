package main

import (
	"context"
	"gobot.io/x/gobot/platforms/raspi"
	"net/http"
	"os"
	"strconv"

	"github.com/Saied74/stationmaster/pkg/code"
)

//seed data for the keyer - tutor
const (
	speed      = "15"
	floatSpeed = 15.0
	farnspeed  = "18"
	floatFarn  = 18.0
	lsm        = "1.2"
	floatLsm   = 1.2
	wsm        = "1.3"
	floatWsm   = 1.3
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
	StopCode  bool
	Logger    bool
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	app.render(w, r, "home.page.html", td)
}

//<++++++++++++++++++++++++++++  Logger  ++++++++++++++++++++++++++++++>

func (app *application) qsolog(w http.ResponseWriter, r *http.Request) {
	var err error

	td := initTemplateData()
	td.Logger = true
	td.Table, err = app.stationModel.getLatestLogs(app.displayLines)
	if err != nil {
		app.serverError(w, err)
	}
	app.render(w, r, "log.page.html", td)
}

func (app *application) addlog(w http.ResponseWriter, r *http.Request) {
	var err error
	td := initTemplateData()
	td.Logger = true
	err = r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
	}

	f := newForm(r.PostForm)

	f.required("call", "sent", "rcvd", "band")
	f.checkAllLogMax()
	f.minLength("sent", 2)
	f.minLength("rcvd", 2)
	f.isInt("sent")
	f.isInt("rcvd")

	if !f.valid() {
		var err error
		td.Table, err = app.stationModel.getLatestLogs(app.displayLines)
		if err != nil {
			app.serverError(w, err)
		}

		td.FormData = f
		td.Show = true
		td.Edit = false
		app.render(w, r, "log.page.html", td)
		return
	}

	tr := copyPostForm(r)

	_, err = app.stationModel.insertLog(&tr)
	if err != nil {
		app.serverError(w, err)
		return
	}

	td.Table, err = app.stationModel.getLatestLogs(app.displayLines)
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Show = false
	td.Edit = false
	app.render(w, r, "log.page.html", td)
}

func (app *application) editlog(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	td.Logger = true
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	tr, err := app.stationModel.getLogByID(id)
	if err != nil {
		app.serverError(w, err)
	}

	td.Table, err = app.stationModel.getLatestLogs(app.displayLines)
	if err != nil {
		app.serverError(w, err)
	}
	app.putId(id)
	td.LogEdit = tr
	td.Show = true
	td.Edit = true
	app.render(w, r, "log.page.html", td)
}

func (app *application) updatedb(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	td.Logger = true
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
	}

	f := newForm(r.PostForm)

	f.required("call", "sent", "rcvd", "band")
	f.checkAllLogMax()
	f.minLength("sent", 2)
	f.minLength("rcvd", 2)
	f.isInt("sent")
	f.isInt("rcvd")

	if !f.valid() {
		var err error
		td.Table, err = app.stationModel.getLatestLogs(app.displayLines)
		if err != nil {
			app.serverError(w, err)
		}
		td.FormData = f
		td.Show = true
		td.Edit = false
		app.render(w, r, "log.page.html", td)
		return
	}
	tr := copyPostForm(r)

	id := app.getId()
	err = app.stationModel.updateLog(&tr, id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	td.Table, err = app.stationModel.getLatestLogs(app.displayLines)
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Show = false
	td.Edit = false
	app.render(w, r, "log.page.html", td)
}

//<++++++++++++++++++++++++++  Antenna  ++++++++++++++++++++++++++++>

func (app *application) ant(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	app.render(w, r, "ant.page.html", td)
}

func (app *application) keyDown(w http.ResponseWriter, r *http.Request) {
	
	cw := &code.CwDriver{
		Dit:       raspi.NewAdaptor(),
	}
	cw.KeyDown()
	
	td := initTemplateData()
	app.render(w, r, "ant.page.html", td)
}

func (app *application) keyUp(w http.ResponseWriter, r *http.Request) {

	cw := &code.CwDriver{
		Dit:       raspi.NewAdaptor(),
	}
	cw.KeyUp()
	
	td := initTemplateData()
	app.render(w, r, "ant.page.html", td)
}

//<++++++++++++++++++++++++  Keyer - Tutor  ++++++++++++++++++++++++++>

func (app *application) ktutor(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	td.StopCode = true
	app.render(w, r, "ktutor.page.html", td) //data)
}

func (app *application) start(w http.ResponseWriter, r *http.Request) {
	//variable naming convention:
	//all constants are lower case and longer
	//thier numeric version has float in front ot it
	//data extracted from the form ends in X (they are all strings)
	//when converted to float64, they are shortened
	td := initTemplateData()
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
	_, _, ktrunning := app.getCancel()
	if ktrunning {
		f.Errors.add("ktrunning", "Keyer-tutor is running, stop it first")
	}

	if !f.valid() {
		td.FormData = f
		td.StopCode = true
		app.render(w, r, "ktutor.page.html", td)
		return
	}
	//get context with cancel so the keyer can be stopped when needed
	ctx, cancel := context.WithCancel(context.Background())
	app.putCancel(ctx, cancel, true)
	cw := &code.CwDriver{
		Dit:       raspi.NewAdaptor(),
		Speed:     s,
		Farnspeed: fs,
		LF:        lf,
		WF:        wf,
	}

	go cw.Work(ctx)

	td.Speed = speedX
	td.FarnSpeed = fspeedX
	td.Lsm = lsmX
	td.Wsm = wsmX

	switch modeX {
	case "1":
		td.Mode = "Tutor"
	case "2":
		td.Mode = "keyer"
	default:
		td.Mode = "tutor"
	}
	td.StopCode = true
	app.render(w, r, "runkt.page.html", td)

}

func (app *application) stop(w http.ResponseWriter, r *http.Request) {
	ctx, cancel, _ := app.getCancel()
	if cancel != nil {
		cancel()
	}
	app.putCancel(ctx, cancel, false)
	//to do: it would be better if the last data the user inputted was
	//used here.
	td := initTemplateData()
	td.StopCode = true
	app.render(w, r, "ktutor.page.html", td)
}

func (app *application) stopcode(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	td.StopCode = true
	app.render(w, r, "runkt.page.html", td)
}

//<++++++++++++++++++++++++++++  Quit  ++++++++++++++++++++++++++++++>

func (app *application) quit(w http.ResponseWriter, r *http.Request) {
	os.Exit(1)
}
