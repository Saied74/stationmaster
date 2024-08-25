package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//seed data for the keyer - tutor

// for feeding dynamic data and error reports to templates
type templateData struct {
	FormData *formData //for form validation error handling
	LookUp   *Ctype    //Full suite of QRZ individual ham data
	Speed    int8      //code sending speed
	Tone     int16     //Practice tone
	Volume   int8      //Practice volume
	Mode     string    //keying mode, tutor or keyer
	Band     string
	Top      headRow   //Log table column titles
	Table    []LogsRow //full set of log table rows
	LogEdit  *LogsRow  //single row of the log table for editing
	Show     bool
	Edit     bool
	StopCode bool
	Logger   bool
	Contest  string
	Stats    *Stats
	VFO      *VFO
	Message  string
}

type Stats struct {
	Contacts           int
	ConfirmedContacts  int
	RepeatContacts     int
	Country            int
	ConfirmedCountry   int
	State              int
	ConfirmedState     int
	County             int
	ConfirmedCounty    int
	CWContacts         int
	ConfirmedCW        int
	ConfirmedCWState   int
	ConfirmedCWCountry int
}

// LogType is for passing data to the add button of the logger
type LogType struct {
	Name    string `json:"Name"`
	Country string `json:"Country"`
	Band    string `json:"Band"` //todo, I don't think this is used anymore
	Mode    string `json:"Mode"` //todo, I don't think this is used anymore
}

// QRZType is for passing data to the call sign search botton of the logger.
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

	//<+++++++++++++++  Start of invalid form handling
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
	//Sent RST
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

//<<=============================  Defaults =================================>>

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
	case "4":
		v = "FT8"
	case "5":
		v = "FT4"
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
		f.required("contestname", "contestdate", "contesttime", "fieldCount")
		f.maxLength("contestname", 45)
		f.maxLength("fieldCount", 1)
		for i, field := range fields {
			fieldName = "field" + strconv.Itoa(i)
			f.required(fieldName)
			f.maxLength(field, 10)
		}
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

		cd := r.PostForm.Get("contestdate")
		f.dateCheck("contestdate")
		if !f.valid() {
			td.FormData = f
			app.render(w, r, "defaults.page.html", td)
			return
		}
		err = app.otherModel.updateDefault("contestdate", cd)
		if err != nil {
			app.serverError(w, err)
			return
		}
		td.LogEdit.ContestDate = cd

		ct := r.PostForm.Get("contesttime")
		f.timeCheck("contesttime")
		if !f.valid() {
			td.FormData = f
			app.render(w, r, "defaults.page.html", td)
			return
		}
		err = app.otherModel.updateDefault("contesttime", ct)
		if err != nil {
			app.serverError(w, err)
			return
		}
		td.LogEdit.ContestTime = ct

		fc := r.PostForm.Get("fieldCount")
		fieldCount, err := strconv.Atoi(fc)
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
		fieldCount += 2
		err = app.otherModel.updateDefault("fieldCount", strconv.Itoa(fieldCount))
		if err != nil {
			app.serverError(w, err)
		}

		fieldNames := r.PostForm.Get("fieldNames")
		fields := strings.Split(fieldNames, ",")
		for i, _ := range fields {
			fields[i] = strings.TrimSpace(fields[i])
		}
		if len(fields) != fieldCount {
			err = log.Errorf("length of fields %v did not match field count %d", fields, fieldCount)
			app.clientError(w, http.StatusBadRequest)
			return
		}
		var fieldName string
		var fieldDataList []string
		for i, field := range fields {
			fieldName = "field" + strconv.Itoa(i)
			err = app.otherModel.updateDefault(fieldName, field)
			if err != nil {
				app.serverError(err)
			}
			fieldData := r.PostForm.Get(fieldName)
			err = app.otherModel.updateDefault(field, fieldData)
			if err != nil {
				app.serverError(err)
			}
			fieldDataList = append(fieldDataList, fieldData)
		}
		switch fieldCount {
		case 2:
			td.LogEdit.Field1Name = fields[0]
			td.LogEdit.Field2Name = fields[1]
			td.LogEdit.Field1Data = fieldDataList[0]
			td.LogEdit.Field2Data = fieldDataList[1]

		case 3:
			td.LogEdit.Field1Name = fields[0]
			td.LogEdit.Field2Name = fields[1]
			td.LogEdit.Field3Name = fields[2]
			td.LogEdit.Field1Data = fieldDataList[0]
			td.LogEdit.Field2Data = fieldDataList[1]
			td.LogEdit.Field3Data = fieldDataList[2]

		case 4:
			td.LogEdit.Field1Name = fields[0]
			td.LogEdit.Field2Name = fields[1]
			td.LogEdit.Field3Name = fields[2]
			td.LogEdit.Field4Name = fields[3]
			td.LogEdit.Field1Data = fieldDataList[0]
			td.LogEdit.Field2Data = fieldDataList[1]
			td.LogEdit.Field3Data = fieldDataList[2]
			td.LogEdit.Field4Data = fieldDataList[3]

		case 5:
			td.LogEdit.Field1Name = fields[0]
			td.LogEdit.Field2Name = fields[1]
			td.LogEdit.Field3Name = fields[2]
			td.LogEdit.Field4Name = fields[3]
			td.LogEdit.Field5Name = fields[4]
			td.LogEdit.Field1Data = fieldDataList[0]
			td.LogEdit.Field2Data = fieldDataList[1]
			td.LogEdit.Field3Data = fieldDataList[2]
			td.LogEdit.Field4Data = fieldDataList[3]
			td.LogEdit.Field5Data = fieldDataList[4]

		default:
			app.serverError(w, fmt.Errorf("fieldCount number error %d", fieldCount))
		}

		//rst := r.PostForm.Get("rst")
		//err = app.otherModel.updateDefault("sent", rst)
		//if err != nil {
		//	app.serverError(w, err)
		//	return
		//}
		//td.LogEdit.Sent = rst

		//e := r.PostForm.Get("exch")
		//err = app.otherModel.updateDefault("exch", e)
		//if err != nil {
		//	app.serverError(w, err)
		//	return
		//}
		//td.LogEdit.ExchSent = e
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

//<-------------------------- CONTESTING ------------------------------------>

func (app *application) contest(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	band, err := app.otherModel.getDefault("Band")
	if err != nil {
		app.serverError(w, err)
		return
	}
	mode, err := app.otherModel.getDefault("Mode")
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Band = band
	td.Mode = mode
	app.render(w, r, "contest.page.html", td) //data)
}

func (app *application) checkDupe(w http.ResponseWriter, r *http.Request) {
	callSign := r.URL.Query().Get("call")
	if callSign == "" {
		app.infoLog.Printf("Got an empty call sign")
		return
	}
	callSign = strings.ToUpper(callSign)
	cn, err := app.otherModel.getDefault("contestname")
	if err != nil {
		app.serverError(w, err)
		return
	}
	cd, err := app.otherModel.getDefault("contestdate")
	if err != nil {
		app.serverError(w, err)
		return
	}
	ct, err := app.otherModel.getDefault("contesttime")
	if err != nil {
		app.serverError(w, err)
		return
	}
	cdt := cd + "T" + ct + ":00Z"
	dateTime, err := time.Parse(time.RFC3339, cdt)
	if err != nil {
		app.serverError(w, err)
		return
	}
	band, err := app.otherModel.getDefault("band")
	if err != nil {
		app.serverError(w, err)
		return
	}
	mode, err := app.otherModel.getDefault("mode")
	if err != nil {
		app.serverError(w, err)
		return
	}
	//returns true if dupe
	dupe, err := app.logsModel.checkDupe(dateTime, cn, callSign, band, mode)
	if err != nil {
		app.serverError(w, err)
		return
	}
	var Duper struct {
		Isdupe string
	}
	Duper.Isdupe = "No"
	if dupe {
		Duper.Isdupe = "Yes"
	}
	b, err := json.Marshal(Duper)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (app *application) updateLog(w http.ResponseWriter, r *http.Request) {
	var c *Ctype
	var v struct {
		Call     string
		RST      string
		Exchange string
		Message  string
	}
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

	//fist, get band and mode
	band, err := app.otherModel.getDefault("band")
	if err != nil {
		app.serverError(w, err)
		return
	}
	mode, err := app.otherModel.getDefault("mode")
	if err != nil {
		app.serverError(w, err)
		return
	}

	//check to see if contest is on
	contestOn, err := app.otherModel.getDefault("contest") //Yes or No
	if err != nil {
		app.serverError(w, err)
		return
	}
	if contestOn != "Yes" {
		v.Message = "Contest is not on, you can't do this."
		b, err := json.Marshal(v)
		if err != nil {
			app.serverError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
		return
	}

	if v.Call == "" || v.RST == "" || v.Exchange == "" {
		v.Message = "There are one or more missing data fields."
		b, err := json.Marshal(v)
		if err != nil {
			app.serverError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
		return

	}

	//<++++++++++++++++  get more defaults

	//Sent RST
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

	//<+++++++++++++++++  Calculate and store the number of logs with that call

	t, err := app.logsModel.getLogsByCall(v.Call)
	if err != nil {
		app.serverError(w, err)
		return
	}
	repeat := true
	c, err = app.qrzModel.getQRZ(v.Call)
	if err != nil {
		if !errors.Is(err, errNoRecord) {
			app.serverError(w, err)
			return
		}
		q, err := app.getHamInfo(v.Call)
		if err != nil {
			app.serverError(w, err)
			return
		}
		c = &q.Callsign
		c.QSOCount = 1
		err = app.qrzModel.insertQRZ(c)
		if err != nil {
			app.serverError(w, err)
			return
		}
		//This is the case that this is the first contact
		repeat = false
	}
	if repeat {
		//this is the case that this the second of more contacts
		err = app.qrzModel.updateQSOCount(call, len(t)+1)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	//<++++++++++++++  Save the new log
	tr := LogsRow{
		Contest:     contestOn,
		ContestName: name,
		Call:        strings.ToUpper(v.Call),
		Sent:        sent,
		Rcvd:        v.RST,
		Band:        band,
		Mode:        mode,
		Name:        c.Fname + " " + c.Lname,
		Country:     c.Country,
		Comment:     "",
		ExchSent:    exchange,
		ExchRcvd:    v.Exchange,
	}

	_, err = app.logsModel.insertLog(&tr)
	if err != nil {
		app.serverError(w, err)
		return
	}
	//<+++++++++++++  New log saved

}
