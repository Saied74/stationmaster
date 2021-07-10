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
	FormData  *formData
	LookUp    *Ctype
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

func (app *application) getConn(w http.ResponseWriter, r *http.Request) {
	callSign := r.URL.Query().Get("call")
	if callSign == "" {
		app.infoLog.Printf("Got an empty call sign")
		return
	}
	mode := r.URL.Query().Get("mode")
	var m string
	switch mode {
	case "1":
		m = "LSB"
	case "2":
		m = "USB"
	case "3":
		m = "CW"
	default:
		m = ""
		app.errorLog.Printf("bad mode value %s was recieved", mode)
		return
	}
	band := r.URL.Query().Get("band")
	var bnd string
	switch band {
	case "1":
		bnd = "160m"
	case "2":
		bnd = "80m"
	case "3":
		bnd = "40m"
	case "4":
		bnd = "20m"
	case "5":
		bnd = "10m"
	default:
		bnd = ""
		app.errorLog.Printf("bad band value %s was recieved", band)
		return
	}
	q, err := app.getHamInfo(callSign)
	if err != nil {
		app.errorLog.Printf("API call to QRZ returned error %v", err)
		return
	}
	update := &LogType{
		Name:    fmt.Sprintf("%s %s", q.Callsign.Fname, q.Callsign.Lname),
		Country: q.Callsign.Country,
		Mode:    m,
		Band:    bnd,
	}
	b, err := json.Marshal(update)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
	//app.infoLog.Printf("q is %v", *q)
}

func (app *application) callSearch(w http.ResponseWriter, r *http.Request) {

	callSign := r.URL.Query().Get("call")
	if callSign == "" {
		app.infoLog.Printf("Got an empty call sign\n") //this is for testing
		return
	}
	c, err := app.stationModel.getQRZ(callSign)
	if errors.Is(err, errNoRecord) {
		q, err := app.getHamInfo(callSign)
		if err != nil {
			app.serverError(w, err)
			return
		}
		err = app.stationModel.stashQRZdata(&q.Callsign)
		if err != nil {
			app.serverError(w, err)
			return
		}

		update := &QRZType{
			Call:     q.Callsign.Call,
			Born:     fmt.Sprintf("Born in: %s", q.Callsign.Born),
			Addr1:    q.Callsign.Addr1,
			Addr2:    fmt.Sprintf("%s %s", q.Callsign.Addr2, q.Callsign.State),
			Country:  q.Callsign.Country,
			Class:    fmt.Sprintf("Class: %s", q.Callsign.Class),
			TimeZone: fmt.Sprintf("Time Zone: %s", q.Callsign.TimeZone),
			QSLCount: fmt.Sprintf("QSO Count: %d", q.Callsign.QSOCount),
		}

		nn := q.Callsign.NickName
		if nn == "" {
			update.Name = fmt.Sprintf("%s %s", q.Callsign.Fname, q.Callsign.Lname)
		} else {
			update.Name = fmt.Sprintf("%s %s (%s)", q.Callsign.Fname, q.Callsign.Lname, q.Callsign.NickName)
		}
		update.QRZMsg = `<p>This record is not in the local database, want to add it?</p>
  <form action="/updateQRZ"><div class="row"><button type="submit" class="btn btn-primary">Update</button></div></form>`

		b, err := json.Marshal(update)
		if err != nil {
			app.serverError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
		return
	} else {
		if err != nil {
			app.serverError(w, err)
			return
		}
	}
	update := &QRZType{
		Call:     c.Call,
		Born:     fmt.Sprintf("Born in: %s", c.Born),
		Addr1:    c.Addr1,
		Addr2:    fmt.Sprintf("%s %s", c.Addr2, c.State),
		Country:  c.Country,
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

	update.QRZMsg = `<p>This record is from the local database.</p>`

	b, err := json.Marshal(update)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (app *application) updateQRZ(w http.ResponseWriter, r *http.Request) {
	app.infoLog.Printf("updateQRZ was called")
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
	}
	app.render(w, r, "log.page.html", td)
}
