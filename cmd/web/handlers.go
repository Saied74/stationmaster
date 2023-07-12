package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	//"time"

	//	"gobot.io/x/gobot/platforms/raspi"

	"github.com/Saied74/stationmaster/pkg/code"
	"github.com/Saied74/stationmaster/pkg/vfo"
)

//seed data for the keyer - tutor
const (
	speed      = "20"
	floatSpeed = 20.0
	farnspeed  = "18"
	floatFarn  = 18.0
	lsm        = "1.0"
	floatLsm   = 1.0
	wsm        = "1.0"
	floatWsm   = 1.0
	dxLines    = 200
	cqZone     = "5" //Eastern US
	maxDXLines = 20
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
		Dit:       app.vfoAdaptor, //raspi.NewAdaptor(),
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

//<++++++++++++++++++++++++++++  VFO  +++++++++++++++++++++++++++++++>

type VFO struct {
	Band       string `json:"Band"`
	Mode       string `json:"Mode"`
	RFreq      string `json:"RFreq"`
	XFreq      string `json:"XFreq"`
	UpperLimit string `json:"UpperLimit"`
	CWBoundary string `json:"CWBoundary"`
	LowerLimit string `json:"LowerLimit"`
	Split      string `json:"Split"`
	VFOBase    string `json:"VFOBase"`
	DX         []DXClusters
	//	Offset     float64 `json:"Offset"`
}

var vfoMemory = map[string]*VFO{
	"10m":  &VFO{UpperLimit: "29.700000", LowerLimit: "28.000000", CWBoundary: "28.300000", VFOBase: "5.010847"},
	"15m":  &VFO{UpperLimit: "21.450000", LowerLimit: "21.000000", CWBoundary: "21.200000", VFOBase: "5.010382"},
	"Aux":  &VFO{UpperLimit: "10.500000", LowerLimit: "10.000000", CWBoundary: "10.500000", VFOBase: "5.000000"},
	"20m":  &VFO{UpperLimit: "14.350000", LowerLimit: "14.000000", CWBoundary: "14.150000", VFOBase: "5.000305"},
	"WWV":  &VFO{UpperLimit: "10.500000", LowerLimit: "10.000000", CWBoundary: "10.500000", VFOBase: "5.011585"},
	"40m":  &VFO{UpperLimit: "7.300000", LowerLimit: "7.000000", CWBoundary: "7.125000", VFOBase: "5.000200"},
	"80m":  &VFO{UpperLimit: "4.000000", LowerLimit: "3.500000", CWBoundary: "3.600000", VFOBase: "5.000000"},
	"160m": &VFO{UpperLimit: "2.000000", LowerLimit: "1.800000", CWBoundary: "1.900000", VFOBase: "5.000000"},
}

func (app *application) startVFO(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	v, err := app.getVFOUpdate()
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.VFO = v
	band, err := app.otherModel.getDefault("band")
	if err != nil {
		app.serverError(w, err)
		return
	}
	//mode, err := app.otherModel.getDefault("mode")
	//if err != nil {
		//app.serverError(w, err)
		//return
	//}
	//fmt.Println("x",  mode)
	noDXData := strings.Contains(band, "Aux") || strings.Contains(band, "WWV")
	if noDXData {
		app.render(w, r, "vfo.page.html", td)
		return
	}
	dx, err := clusters(band, dxLines)
	if err != nil {
		if errors.Is(err, errNoDXSpots) {
			app.render(w, r, "vfo.page.html", td) //data)
			return
		}
		app.infoLog.Printf("error from calling clusters in startVFO %v\n", err)
	}
	dx, err = app.pickZone(cqZone, dx)
	if err != nil {
		app.serverError(w, err)
	}
	dx, err = app.logsModel.findNeed(dx)
	if err != nil {
		app.serverError(w, err)
		app.render(w, r, "vfo.page.html", td)
		return
	}
	//mode, err = app.otherModel.getDefault("mode")
	//if err != nil {
		//app.serverError(w, err)
		//return
	//}
	//fmt.Println("y",  mode)
	td.VFO.DX = dx
	app.render(w, r, "vfo.page.html", td) //data)
}

func (app *application) updateVFO(w http.ResponseWriter, r *http.Request) {
	var v VFO
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	err = json.NewDecoder(r.Body).Decode(&v)
	if err != nil {
		app.serverError(w, err)
		return
	}
	band, err := app.otherModel.getDefault("band")
	if err != nil {
		app.serverError(w, err)
		return
	}
	xf := band + "xfreq"
	err = app.otherModel.updateDefault(xf, v.XFreq)
	if err != nil {
		app.serverError(w, err)
		return
	}
	rf := band + "rfreq"
	err = app.otherModel.updateDefault(rf, v.RFreq)
	if err != nil {
		app.serverError(w, err)
		return
	}
	err = app.otherModel.updateDefault("split", v.Split)
	if err != nil {
		app.serverError(w, err)
		return
	}
	vfoSet := vfoMemory[band]
	lowerLimit, err := strconv.ParseFloat(vfoSet.LowerLimit, 64)
	if err != nil {
		app.serverError(w, err)
	}
	rFreq, err := strconv.ParseFloat(v.RFreq, 64)
	if err != nil {
		app.serverError(w, err)
	}
	//	rFreq += vfoSet.Offset
	xFreq, err := strconv.ParseFloat(v.XFreq, 64)
	if err != nil {
		app.serverError(w, err)
	}
	b, err := strconv.ParseFloat(vfoSet.VFOBase, 64)
	if err != nil {
		app.serverError(w, err)
	}
	xFreq = xFreq - lowerLimit + b
	xPhase := (xFreq * math.Pow(2.0, 32.0)) / 125.0
	xP := int(math.Round(xPhase))
	rFreq = rFreq - lowerLimit + b
	rPhase := (rFreq * math.Pow(2.0, 32.0)) / 125.0
	rP := int(math.Round(rPhase))
	//fmt.Println("before runvfo", band, xf, rf)
	vfo.Runvfo(app.vfoAdaptor, xP, rP)
}

type BandUpdate struct {
	Band    string `json:"Band"`
	Mode    string `json:"Mode"`
	DXTable []DXClusters
}

var switchTable = map[int]BandUpdate{
	0: BandUpdate{Band: "10m", Mode: "USB"},
	1: BandUpdate{Band: "15m", Mode: "USB"},
	2: BandUpdate{Band: "Aux", Mode: "USB"},
	3: BandUpdate{Band: "20m", Mode: "USB"},
	4: BandUpdate{Band: "WWV", Mode: "USB"},
	5: BandUpdate{Band: "40m", Mode: "LSB"},
	6: BandUpdate{Band: "80m", Mode: "LSB"},
	7: BandUpdate{Band: "160m", Mode: "LSB"},
}

func (app *application) updateBand(w http.ResponseWriter, r *http.Request) {
	var b int
	update := BandUpdate{}
	mode, err := app.otherModel.getDefault("mode")
	if err != nil {
		app.serverError(w, err)
		return
	}
	band, err := app.otherModel.getDefault("band")
	if err != nil {
		app.serverError(w, err)
		return
	}
	select {
	case b = <-app.bandData.Band:
		update, ok := switchTable[b]
		if !ok {
			app.serverError(w, fmt.Errorf("bad data from the switch %d", b))
			return
		}
		//if mode != "FT8" || mode != "FT4" {
			//err = app.otherModel.updateDefault("mode", update.Mode)
			//if err != nil {
				//app.serverError(w, err)
				//return
			//}
		//}
		err = app.otherModel.updateDefault("band", update.Band)
			if err != nil {
				app.serverError(w, err)
				return
			}	
	default:
		update = BandUpdate{
			Mode: mode,
			Band: band,
		}
	}

	u, err := json.Marshal(update)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(u)
}

func (app *application) updateDX(w http.ResponseWriter, r *http.Request) {
	band, err := app.otherModel.getDefault("band")
	if err != nil {
		app.serverError(w, err)
		return
	}
	dx, err := clusters(band, dxLines)
	if err != nil {
		if errors.Is(err, errNoDXSpots) {
			return
		}
		app.infoLog.Printf("error from calling clusters in updateDX %v\n", err)
		return
	}
	dx, err = app.pickZone(cqZone, dx)
	if err != nil {
		app.serverError(w, err)
		return
	}
	dx, err = app.logsModel.findNeed(dx)
	if err != nil {
		app.serverError(w, err)
		return
	}
	update := BandUpdate{
		DXTable: dx,
	}
	u, err := json.Marshal(update)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(u)
}
