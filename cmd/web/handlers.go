package main

import (
	"context"
	"net/http"
	"os"

	"gobot.io/x/gobot/platforms/raspi"

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

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	app.render(w, r, "home.page.html", td)
}

//<++++++++++++++++++++++++++  Antenna  ++++++++++++++++++++++++++++>

func (app *application) ant(w http.ResponseWriter, r *http.Request) {
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
	modeX := r.PostForm.Get("mode") //tutor or keyer
	var whichOutput string
	var hi, low byte
	switch modeX {
	case "1":
		whichOutput = code.TutorOutput
		hi = byte(0)
		low = byte(1)
		td.Mode = "Tutor"
	case "2":
		whichOutput = code.KeyerOutput
		hi = byte(1)
		low = byte(0)
		td.Mode = "Keyer"
	default:
		hi = byte(0)
		low = byte(1)
		td.Mode = "Tutor"
		f.Errors.add("ktrunning", "No mode selected, set to Tutor")
	}
	cw := &code.CwDriver{
		Dit:       raspi.NewAdaptor(),
		Speed:     s,
		Farnspeed: fs,
		LF:        lf,
		WF:        wf,
		Output:    whichOutput,
		Hi:        hi,
		Low:       low,
	}

	go cw.Work(ctx)

	td.Speed = speedX
	td.FarnSpeed = fspeedX
	td.Lsm = lsmX
	td.Wsm = wsmX
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
