package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

//seed data for the keyer - tutor

//for feeding dynamic data and error reports to templates
type templateData struct {
	FormData  *formData //for form validation error handling
	LookUp    *Ctype    //Full suite of QRZ individual ham data
	Speed     string    //code sending speed
	FarnSpeed string    //Farnsworth sending speed
	Lsm       string    //Letter spacing modifier
	Wsm       string    //word spacing modifier
	Mode      string    //keying mode, tutor or keyer
	Top       headRow   //Log table column titles
	Table     []LogsRow //full set of log table rows
	LogEdit   *LogsRow  //single row of the log table for editing
	Show      bool
	Edit      bool
	StopCode  bool
	Logger    bool
}

type LogType struct {
	Name    string `json:"Name"`
	Country string `json:"Country"`
	Band    string `json:"Band"`
	Mode    string `json:"Mode"`
}

type QRZType struct {
	QRZMsg   string `json:"QRZMsg"`
	Call     string `json:"Call"`
	Name     string `json:"Name"`
	Born     string `json:"Born"`
	Addr1    string `json:"Addr1"`
	Addr2    string `json:"Addr2"`
	Country  string `json:"QRZCountry"`
	GeoLoc   string `json:"GeoLoc"`
	Class    string `json:"Class"`
	TimeZone string `json:"TimeZone"`
	QSLCount string `json:"QSOCount"`
}

//<++++++++++++++++++++++++++++  Logger  ++++++++++++++++++++++++++++++>

func (app *application) qsolog(w http.ResponseWriter, r *http.Request) {
	var err error

	td := initTemplateData()
	td.Logger = true
	td.Table, err = app.stationModel.getLatestLogs(app.displayLines)
	if err != nil {
		app.serverError(w, err)
		return
	}
	v, err := app.stationModel.getDefault("band")
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.FormData.Set("band", v)
	v, err = app.stationModel.getDefault("mode")
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Mode = v //this is a workaround.  Template library does not seem to like emtpy strings
	app.render(w, r, "log.page.html", td)
}

func (app *application) addlog(w http.ResponseWriter, r *http.Request) {
	var c *Ctype
	var err error
	td := initTemplateData()
	td.Logger = true
	err = r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
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
			return
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

	call := f.Get("call")
	t, err := app.stationModel.getLogsByCall(call)
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Show = false
	td.Edit = false

	c, err = app.stationModel.getQRZ(call)
	if err != nil {
		if errors.Is(err, errNoRecord) {
			q, err := app.getHamInfo(call)
			if err != nil {
				app.serverError(w, err)
				return
			}
			c = &q.Callsign

			c.QSOCount = len(t)
			err = app.stationModel.insertQRZ(c)
			if err != nil {
				app.serverError(w, err)
				return
			}
			app.render(w, r, "log.page.html", td)
			return
		} else {
			app.serverError(w, err)
			return
		}
	}
	err = app.stationModel.updateQSOCount(call, len(t))
	if err != nil {
		app.serverError(w, err)
		return
	}
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
		return
	}

	td.Table, err = app.stationModel.getLatestLogs(app.displayLines)
	if err != nil {
		app.serverError(w, err)
		return
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
		return
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
			return
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
	v, err := app.stationModel.getDefault("band")
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.FormData.Set("band", v)
	v, err = app.stationModel.getDefault("mode")
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Mode = v //this is a workaround.  Template library does not seem to like emtpy strings
	td.Show = false
	td.Edit = false
	app.render(w, r, "log.page.html", td)
}

func (app *application) getConn(w http.ResponseWriter, r *http.Request) {
	callSign := r.URL.Query().Get("call")
	if callSign == "" {
		app.infoLog.Printf("Got an empty call sign")
		return
	}
	c, err := app.stationModel.getQRZ(callSign)
	if err != nil {
		if errors.Is(err, errNoRecord) {
			q, err := app.getHamInfo(callSign)
			if err != nil {
				app.serverError(w, err)
				return
			}
			c = &q.Callsign
		} else {
			app.serverError(w, err)
			return
		}
	}
	update := &LogType{
		Name:    fmt.Sprintf("%s %s", c.Fname, c.Lname),
		Country: c.Country,
	}
	b, err := json.Marshal(update)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (app *application) callSearch(w http.ResponseWriter, r *http.Request) {
	var c *Ctype
	msg := `<p>This record is from the local database.</p>`
	callSign := r.URL.Query().Get("call")
	if callSign == "" {
		app.infoLog.Printf("Got an empty call sign\n") //this is for testing
		return
	}
	c, err := app.stationModel.getQRZ(callSign)
	if err != nil {
		if errors.Is(err, errNoRecord) {
			q, err := app.getHamInfo(callSign)
			if err != nil {
				app.serverError(w, err)
				return
			}
			c = &q.Callsign
			msg = `<p>This record is not in the local database, want to add it?</p>
		  <form action="/updateQRZ"><div class="row"><button type="submit" class="btn btn-primary">Update</button></div></form>`
		} else {
			app.serverError(w, err)
			return
		}
	}
	err = app.stationModel.stashQRZdata(c)
	if err != nil {
		app.serverError(w, err)
		return
	}

	update := &QRZType{
		Call:     c.Call,
		Born:     fmt.Sprintf("Born in: %s", c.Born),
		Addr1:    fmt.Sprintf("%s   %s   %s   %s", c.Addr1, c.Addr2, c.State, c.Country),
		GeoLoc:   fmt.Sprintf("Lat: %s,   Long: %s,   Grid: %s", c.Lat, c.Long, c.Grid),
		Class:    fmt.Sprintf("Class: %s", c.Class),
		TimeZone: fmt.Sprintf("Time Zone: %s", c.TimeZone),
		QSLCount: fmt.Sprintf("QSO Count: %d", c.QSOCount),
	}

	nn := c.NickName
	if nn == "" {
		update.Name = fmt.Sprintf("%s %s", c.Fname, c.Lname)
	} else {
		update.Name = fmt.Sprintf("%s %s (%s)", c.Fname, c.Lname, c.NickName)
	}
	update.QRZMsg = msg

	b, err := json.Marshal(update)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (app *application) updateQRZ(w http.ResponseWriter, r *http.Request) {
	c, err := app.stationModel.unstashQRZdata()
	if err != nil {
		app.serverError(w, err)
		return
	}
	logs, err := app.stationModel.getLogsByCall(c.Call)
	c.QSOCount = len(logs)
	err = app.stationModel.insertQRZ(c)
	if err != nil {
		app.serverError(w, err)
		return
	}

	td := initTemplateData()
	td.Logger = true
	td.Table, err = app.stationModel.getLatestLogs(app.displayLines)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.render(w, r, "log.page.html", td)
}

func (app *application) defaults(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	td.Logger = true
	v, err := app.stationModel.getDefault("mode")
	if err != nil {
		if errors.Is(err, errNoRecord) {
			v = "USB"
		} else {
			app.serverError(w, err)
			return
		}
	}
	td.LogEdit.Mode = v
	v, err = app.stationModel.getDefault("band")
	if err != nil {
		if errors.Is(err, errNoRecord) {
			v = "20m"
		} else {
			app.serverError(w, err)
			return
		}
	}
	td.LogEdit.Band = v
	app.render(w, r, "defaults.page.html", td)
}

func (app *application) storeDefaults(w http.ResponseWriter, r *http.Request) {
	var v string
	td := initTemplateData()
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	m := r.PostForm.Get("mode")
	switch m {
	case "1":
		v = "USB"
	case "2":
		v = "LSB"
	case "3":
		v = "CW"
	default:
		v = fmt.Sprintf("A Bad Mode Choice. ")
	}
	err = app.stationModel.updateDefault("mode", v)
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.LogEdit.Mode = v
	b := r.PostForm.Get("band")
	switch b {
	case "1":
		v = "10m"
	case "2":
		v = "20m"
	case "3":
		v = "40m"
	case "4":
		v = "80m"
	case "5":
		v = "160m"
	default:
		v = fmt.Sprintf("A Bad Band Choice. ")
	}
	err = app.stationModel.updateDefault("band", v)
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.LogEdit.Band = v
	app.render(w, r, "defaults.page.html", td)
}

func (app *application) contacts(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	call := r.URL.Query().Get("contact-call")
	if call == "" {
		app.infoLog.Printf("Got an empty call sign\n") //this is for testing
		return
	}
	c, err := app.stationModel.getQRZ(call)
	if err != nil {
		if errors.Is(err, errNoRecord) {
			c.Call = "call sign not in the database"
		} else {
			app.serverError(w, err)
			return
		}
	}
	td.LookUp = c
	app.render(w, r, "contacts.page.html", td)
}
