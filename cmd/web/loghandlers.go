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
	Contest   string
	Stats     *Stats
	VFO       *VFO
}

type Stats struct {
	Contacts          int
	ConfirmedContacts int
	RepeatContacts    int
	Country           int
	ConfirmedCountry  int
	State             int
	ConfirmedState    int
	County            int
	ConfirmedCounty   int
}

//LogType is for passing data to the add button of the logger
type LogType struct {
	Name    string `json:"Name"`
	Country string `json:"Country"`
	Band    string `json:"Band"` //todo, I don't think this is used anymore
	Mode    string `json:"Mode"` //todo, I don't think this is used anymore
}

//QRZType is for passing data to the call sign search botton of the logger.
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
	v, err := app.otherModel.getDefault("contest") //Yes or No
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Contest = v
	td.FormData.Set("contest", v)
	td.LogEdit.Contest = v

	if v == "No" {
		td.Table, err = app.logsModel.getLatestLogs(app.displayLines)
		if err != nil {
			app.serverError(w, err)
			return
		}
	} else {
		td.Table, err = app.logsModel.getContestLogs(app.displayLines)
		if err != nil {
			app.serverError(w, err)
			return
		}
		v, err = app.otherModel.getDefault("contestname")
		if err != nil {
			app.serverError(w, err)
			return
		}
		td.LogEdit.ContestName = v
		v, err = app.otherModel.getDefault("sent")
		if err != nil {
			app.serverError(w, err)
			return
		}
		td.FormData.Set("sent", v)
		td.LogEdit.ExchSent = v
		v, err = app.otherModel.getDefault("exch")
		if err != nil {
			app.serverError(w, err)
			return
		}
		td.FormData.Set("exchsent", v)
		td.LogEdit.ExchSent = v
	}
	v, err = app.otherModel.getDefault("band")
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.FormData.Set("band", v)
	v, err = app.otherModel.getDefault("mode")
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Mode = v

	app.render(w, r, "log.page.html", td)
}

func (app *application) addlog(w http.ResponseWriter, r *http.Request) {
	var c *Ctype
	var err error
	td := initTemplateData()
	td.Logger = true
	contestOn, err := app.otherModel.getDefault("contest") //Yes or No
	if err != nil {
		app.serverError(w, err)
		return
	}
	//   td.Contest = contestOn
	//   tr.Contest = contestOn
	//   td.FormData.Set("contest", contestOn)
	// tr.Contest = contestOn
	err = r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	f := newForm(r.PostForm)
	f.required("call", "sent", "rcvd", "band")
	if contestOn == "Yes" {
		f.required("exchsent", "exchrcvd")
		//// TODO: Add max length
	}
	f.checkAllLogMax()
	f.minLength("sent", 2)
	f.minLength("rcvd", 2)
	f.isInt("sent")
	f.isInt("rcvd")

	//<+++++++++++++++  Start of invalid forrm handling
	if !f.valid() {
		var err error
		if contestOn == "No" {
			td.Table, err = app.logsModel.getLatestLogs(app.displayLines)
			if err != nil {
				app.serverError(w, err)
				return
			}
		} else {
			td.Table, err = app.logsModel.getContestLogs(app.displayLines)
			if err != nil {
				app.serverError(w, err)
				return
			}
		}
		td.FormData = f
		td.Show = true
		td.Edit = false
		app.render(w, r, "log.page.html", td)
		return
	}
	//<++++++++++++++++ end of invalid form handling
	tr := copyPostForm(r)
	//<++++++++++++++++  get defaults
	//Band
	band, err := app.otherModel.getDefault("band")
	if err != nil {
		app.serverError(w, err)
		return
	}
	//Mode
	mode, err := app.otherModel.getDefault("mode")
	if err != nil {
		app.serverError(w, err)
		return
	}
	//Sent RRST
	sent, err := app.otherModel.getDefault("sent")
	if err != nil {
		app.serverError(w, err)
		return
	}
	//Exchange message
	exchange, err := app.otherModel.getDefault("exch")
	if err != nil {
		app.serverError(w, err)
		return
	}
	name, err := app.otherModel.getDefault("contestname")
	if err != nil {
		app.serverError(w, err)
		return
	}
	//<++++++++++++++++ end of get defaults
	// tr.ContestName = v
	// td.FormData.Set("sent", v)

	//<++++++++++++++  Save the new log
	if contestOn == "Yes" {
		tr.Contest = contestOn
		tr.ContestName = name
	}
	_, err = app.logsModel.insertLog(&tr)
	if err != nil {
		app.serverError(w, err)
		return
	}
	//<+++++++++++++  New log saved

	//<+++++++++++++  Set up the new display

	td.FormData.Set("band", band)
	td.Mode = mode

	if contestOn == "No" {
		td.Table, err = app.logsModel.getLatestLogs(app.displayLines)
		if err != nil {
			app.serverError(w, err)
			return
		}
	} else {
		td.Table, err = app.logsModel.getContestLogs(app.displayLines)
		if err != nil {
			app.serverError(w, err)
			return
		}
		// td.FormData.Set("contestname", name)
		tr.ExchSent = sent
		td.FormData.Set("sent", sent)
		td.FormData.Set("exchsent", exchange)
	}
	//<+++++++++++++++++  Calculate and store the number of logs with that call
	call := f.Get("call")
	t, err := app.logsModel.getLogsByCall(call)
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Show = false
	td.Edit = false
	c, err = app.qrzModel.getQRZ(call)
	if err != nil {
		if errors.Is(err, errNoRecord) {
			q, err := app.getHamInfo(call)
			if err != nil {
				app.serverError(w, err)
				return
			}
			c = &q.Callsign
			c.QSOCount = len(t)
			err = app.qrzModel.insertQRZ(c)
			if err != nil {
				app.serverError(w, err)
				return
			}
			//This is the case that this is the first contact
			app.render(w, r, "log.page.html", td)
			return
		}
		app.serverError(w, err)
		return
	}
	//this is the case that this the second of more contacts
	err = app.qrzModel.updateQSOCount(call, len(t))
	if err != nil {
		app.serverError(w, err)
		return
	}
	//<+++++++++++++++  Numberr of calls updated
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
	tr, err := app.logsModel.getLogByID(id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	td.Table, err = app.logsModel.getLatestLogs(app.displayLines)
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
		td.Table, err = app.logsModel.getLatestLogs(app.displayLines)
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
	err = app.logsModel.updateLog(&tr, id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	td.Table, err = app.logsModel.getLatestLogs(app.displayLines)
	if err != nil {
		app.serverError(w, err)
		return
	}
	v, err := app.otherModel.getDefault("band")
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.FormData.Set("band", v)
	v, err = app.otherModel.getDefault("mode")
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
	c, err := app.qrzModel.getQRZ(callSign)
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
	c, err := app.qrzModel.getQRZ(callSign)
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
	err = app.qrzModel.stashQRZdata(c)
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
	c, err := app.qrzModel.unstashQRZdata()
	if err != nil {
		app.serverError(w, err)
		return
	}
	logs, err := app.logsModel.getLogsByCall(c.Call)
	c.QSOCount = len(logs)
	err = app.qrzModel.insertQRZ(c)
	if err != nil {
		app.serverError(w, err)
		return
	}

	td := initTemplateData()
	td.Logger = true
	td.Table, err = app.logsModel.getLatestLogs(app.displayLines)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.render(w, r, "log.page.html", td)
}

func (app *application) defaults(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	td.Logger = true
	v, err := app.lookupDefault("mode")
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.LogEdit.Mode = v
	v, err = app.lookupDefault("band")
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.LogEdit.Band = v
	v, err = app.lookupDefault("contest")
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.LogEdit.Contest = v
	v, err = app.lookupDefault("contestname")
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.LogEdit.ContestName = v
	v, err = app.lookupDefault("sent")
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.LogEdit.Sent = v
	v, err = app.lookupDefault("exchange")
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.LogEdit.ExchSent = v
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
		v, err = app.lookupDefault("mode")
		if err != nil {
			app.serverError(w, err)
			return
		}
	}
	err = app.otherModel.updateDefault("mode", v)
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
		v = "15m"
	case "3":
		v = "20m"
	case "4":
		v = "40m"
	case "5":
		v = "80m"
	case "6":
		v = "160m"
	default:
		v, err = app.lookupDefault("band")
		if err != nil {
			app.serverError(w, err)
			return
		}
	}
	err = app.otherModel.updateDefault("band", v)
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.LogEdit.Band = v
	c := r.PostForm.Get("contest")
	switch c {
	case "1":
		v = "Yes"
	case "2":
		v = "No"
	default:
		v, err = app.lookupDefault("contest")
		if err != nil {
			app.serverError(w, err)
			return
		}
	}
	err = app.otherModel.updateDefault("contest", v)
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.LogEdit.Contest = v

	f := newForm(r.PostForm)
	if td.LogEdit.Contest == "Yes" {
		f.required("contestname", "rst", "exch")
		f.maxLength("contestname", 45)
		f.maxLength("rst", 3)
		f.maxLength("exch", 6)
		if !f.valid() {
			td.FormData = f
			app.render(w, r, "defaults.page.html", td)
			return
		}

		cn := r.PostForm.Get("contestname")
		err = app.otherModel.updateDefault("contestname", cn)
		if err != nil {
			app.serverError(w, err)
			return
		}
		td.LogEdit.ContestName = cn

		rst := r.PostForm.Get("rst")
		err = app.otherModel.updateDefault("sent", rst)
		if err != nil {
			app.serverError(w, err)
			return
		}
		td.LogEdit.Sent = rst

		e := r.PostForm.Get("exch")
		err = app.otherModel.updateDefault("exch", e)
		if err != nil {
			app.serverError(w, err)
			return
		}
		td.LogEdit.ExchSent = e
	}
	app.render(w, r, "defaults.page.html", td)
}

func (app *application) contacts(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	call := r.URL.Query().Get("contact-call")
	if call == "" {
		app.infoLog.Printf("Got an empty call sign\n") //this is for testing
		return
	}
	c, err := app.qrzModel.getQRZ(call)
	if err != nil {
		if errors.Is(err, errNoRecord) {
			c = &Ctype{}
			c.Call = "call sign not in the database"
		} else {
			app.serverError(w, err)
			return
		}
	}
	td.LookUp = c
	app.render(w, r, "contacts.page.html", td)
}
