package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	//	"github.com/Saied74/stationmaster/pkg/code"

	"github.com/Saied74/stationmaster/pkg/vfo"
)

// seed data for the keyer - tutor
const (
	speed      = 22
	tone       = 650
	volume     = 5
	cwMode     = "Keyer"
	dxLines    = 20
	cqZone     = "5" //Eastern US
	maxDXLines = 20
)

type cwData struct {
	speed   int8
	volume  int8
	tone    int16
	cmd     byte
	RcvFreq float64 //dummy for now
	Band    string  //dummy for now
}

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

// This function (and related areas in the code) have been modified to work
// with the new Arduino based CW keyer and practice oscillator
func (app *application) start(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	cmd := &cwData{}
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
	}
	f := newForm(r.PostForm)

	s, err := f.extractCWParameter("speed")
	if err != nil {
		cmd.speed = int8(1200 / speed)
		s = speed
	} else {
		cmd.speed = int8(1200 / s)
	}
	t, err := f.extractCWParameter("tone")
	if err != nil {
		cmd.tone = int16(tone)
	} else {
		cmd.tone = int16(t)
	}
	v, err := f.extractCWParameter("volume")
	if err != nil {
		cmd.volume = int8(volume)
	} else {
		cmd.volume = int8(-14*v + 141)
	}

	if !f.valid() {
		td.FormData = f
		td.StopCode = true
		app.render(w, r, "ktutor.page.html", td)
		return
	}
	modeX := r.PostForm.Get("mode") //tutor or keyer
	switch modeX {
	case "1":
		td.Mode = "Tutor"
		cmd.cmd = tutor
	case "2":
		td.Mode = "Keyer"
		cmd.cmd = keyer
	default:
		td.Mode = cwMode
		if cwMode == "Tutor" {
			cmd.cmd = tutor
		} else {
			cmd.cmd = keyer
		}
	}
	err = app.issueCWCmd(cmd)
	if err != nil {
		f.Errors.add("ktrunning", fmt.Sprintf("error from new CW %v", err))
	}

	td.Speed = int8(s)
	td.Tone = cmd.tone
	td.Volume = cmd.volume
	//td.Wsm = wsmX
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
	FT8Freq    string `json:"FT8Freq"`
	FT4Freq    string `json:"FT4Freq"`
	Split      string `json:"Split"`
	VFOBase    string `json:"VFOBase"`
	DX         []DXClusters
	//	Offset     float64 `json:"Offset"`
}

var vfoMemory = map[string]*VFO{
	"10m":  &VFO{UpperLimit: "29.700000", LowerLimit: "28.000000", CWBoundary: "28.300000", VFOBase: "5.010847", FT8Freq: "28.074000", FT4Freq: "28.180000"},
	"15m":  &VFO{UpperLimit: "21.450000", LowerLimit: "21.000000", CWBoundary: "21.200000", VFOBase: "5.010382", FT8Freq: "21.074000", FT4Freq: "21.140000"},
	"Aux":  &VFO{UpperLimit: "10.500000", LowerLimit: "10.000000", CWBoundary: "10.500000", VFOBase: "5.000000", FT8Freq: "", FT4Freq: ""},
	"20m":  &VFO{UpperLimit: "14.350000", LowerLimit: "14.000000", CWBoundary: "14.150000", VFOBase: "5.000305", FT8Freq: "14.074000", FT4Freq: "14.080000"},
	"WWV":  &VFO{UpperLimit: "10.500000", LowerLimit: "10.000000", CWBoundary: "10.500000", VFOBase: "5.011585", FT8Freq: "", FT4Freq: ""},
	"40m":  &VFO{UpperLimit: "7.300000", LowerLimit: "7.000000", CWBoundary: "7.125000", VFOBase: "5.000200", FT8Freq: "7.074000", FT4Freq: "7.047500"},
	"80m":  &VFO{UpperLimit: "4.000000", LowerLimit: "3.500000", CWBoundary: "3.600000", VFOBase: "5.000000", FT8Freq: "3.573000", FT4Freq: "3.575000"},
	"160m": &VFO{UpperLimit: "2.000000", LowerLimit: "1.800000", CWBoundary: "1.900000", VFOBase: "5.000000", FT8Freq: "", FT4Freq: ""},
}

//var bandUpdateError error

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

	noDXData := strings.Contains(band, "Aux") || strings.Contains(band, "WWV")
	if noDXData {
		app.render(w, r, "vfo.page.html", td)
		return
	}
	var dx []DXClusters
	dx, err = app.getSpider(band, dxLines)
	if err != nil {
		err = app.spiderError(err)
		if err != nil {
			app.serverError(w, err)
			//app.render(w, r, "vfo.page.html", td)
			return
		}
		dx, err = app.getSpider(band, dxLines)
		if err != nil {
			app.serverError(w, err)
			//app.render(w, r, "vfo.page.html", td)
			return
		}
	}

	dx, err = app.logsModel.findNeed(dx)
	if err != nil {
		app.serverError(w, err)
		//app.render(w, r, "vfo.page.html", td)
		return
	}
	td.VFO.Band = band
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
	rFreq = rFreq - lowerLimit + b
	app.cw.RcvFreq = rFreq
	app.cw.Band = band
	//app.setSplit("noSplit")
	rPhase := (rFreq * math.Pow(2.0, 32.0)) / 125.0
	rP := uint32(rPhase)
	//fmt.Println("rP: ", rP)
	app.setFrequency(rP, "tx")
	vfo.Runvfo(app.vfoAdaptor, xFreq, rFreq)
}

type BandUpdate struct {
	Band    string       `json:"Band"`
	Mode    string       `json:"Mode"`
	DXTable []DXClusters `json:"DXTable"`
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

// triggered by regular update requests from the web page vfo.page.html
func (app *application) updateBand(w http.ResponseWriter, r *http.Request) {
	var err error
	var v = &VFO{}
	var badV = false
	v, err = app.getUpdateBand() //reads the band switch and updates DB
	if err != nil && !errors.Is(err, noPortMatch) {
		badV = true
		app.serverError(w, err)
	}
	if errors.Is(err, noPortMatch) {
		badV = true
	}
	err = app.getUpdateMode(v) //calculates mode from band and xmit freq, updates DB
	if err != nil {
		badV = true
		app.serverError(w, err)
	}
	var u = []byte{}
	if !badV {
		u, err = json.Marshal(*v)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}
	app.cw.Band = v.Band
	w.Header().Set("Content-Type", "application/json")
	w.Write(u)
}

func (app *application) updateDX(w http.ResponseWriter, r *http.Request) {
	update := BandUpdate{}
	var validDX = true
	band, err := app.otherModel.getDefault("band")
	if err != nil {
		app.serverError(w, err)
		return
	}
	dx, err := app.getSpider(band, dxLines)
	if err != nil {
		err = app.spiderError(err)
		if err != nil {
			log.Println("EOF Error: ", err)
			log.Printf("EOF Error Type %T\n", err)
			app.serverError(w, err)
			validDX = false
		}
		if validDX {
			dx, err = app.getSpider(band, dxLines)
			if err != nil {
				app.serverError(w, err)
				validDX = false
			}
		}
	}
	if validDX {
		update.DXTable = dx
	}
	u, err := json.Marshal(update)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(u)
}
