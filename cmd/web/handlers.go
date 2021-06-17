package main

import (
	"context"
	"gobot.io/x/gobot/platforms/raspi"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	speed      = "15"
	floatSpeed = 15.0
	farnspeed  = "18"
	floatFarn  = 18.0
	lsm        = "1.2"
	floatLsm   = 1.55
	wsm        = "1.3"
	floatWsm   = 1.3
)

type templateData struct {
	FormData   url.Values
	FormErrors map[string]string
	Speed      string
	FarnSpeed  string
	Lsm        string
	Wsm        string
	Mode       string
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "home.page.html", nil)
}

func (app *application) ktutor(w http.ResponseWriter, r *http.Request) {
	data := getBaseTemp()
	app.render(w, r, "ktutor.page.html", data)
}

func (app *application) qsolog(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "log.page.html", nil)
}

func (app *application) ant(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "ant.page.html", nil)
}

func (app *application) start(w http.ResponseWriter, r *http.Request) {
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

	if app.ktrunning {
		errors["ktrunning"] = "Keyer-tutor is running, stop it first"
	}

	if len(errors) > 0 {
		data.FormData = r.PostForm
		data.FormErrors = errors
		app.render(w, r, "ktutor.page.html", data)
		return
	}

	app.ctx, app.cancel = context.WithCancel(context.Background())
	cw := &cwDriver{
		dit:       raspi.NewAdaptor(),
		speed:     s,
		farnspeed: fs,
		lF:        lf,
		wF:        wf,
	}

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
	data := getBaseTemp()
	app.render(w, r, "ktutor.page.html", data)
}

func (app *application) stopcode(w http.ResponseWriter, r *http.Request) {
	data := getBaseTemp()
	app.render(w, r, "runkt.page.html", data)
}
