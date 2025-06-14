package main

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
)

func (app *application) adif(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	t, err := app.logsModel.getADIFData()
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Table = t
	app.render(w, r, "adif.page.html", td)
}

func (app *application) genadif(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	t, err := app.logsModel.getADIFData()
	if err != nil {
		app.serverError(w, err)
		return
	}
	err = app.genADIFFile(t)
	if err != nil {
		app.serverError(w, err)
		return
	}
	for _, row := range t {
		app.logsModel.updateLOTWSent(row.Id)
		if err != nil {
			app.serverError(w, err)
		}
	}
	t, err = app.logsModel.getADIFData()
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Table = t
	app.render(w, r, "adif.page.html", td)
}

func (app *application) confirmQSLs(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		app.clientError(w, http.StatusBadRequest)
		return
	}
	qslFile := r.PostForm.Get("qslfile")

	output, err := app.getQSLData(qslFile)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	for _, row := range output {
		err = app.logsModel.updateQSO(row)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	td := initTemplateData()
	td.Logger = true
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
	app.render(w, r, "log.page.html", td)
}

//<---------------------------  Cabrillo ---------------------------------->

func (app *application) cabrillo(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	fc, err := app.otherModel.getDefault("fieldCount")
	if err != nil {
		app.serverError(w, err)
		return
	}
	c, err := strconv.Atoi(fc)
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.FieldCount = c
	for i := 0; i < c; i++ {
		q := fmt.Sprintf("field%dName", i+1)
		sfn, err := app.otherModel.getDefault(q)
		if err != nil {
			app.serverError(w, err)
		}
		td.FieldNames = append(td.FieldNames, sfn)
	}
	app.render(w, r, "cabrillo.page.html", td)
}

func (app *application) genCabrillo(w http.ResponseWriter, r *http.Request) {
	var err error
	td := initTemplateData()

	err = r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	f := newForm(r.PostForm)
	f.required("contestname", "starttime", "startdate", "enddate",
		"endtime", "contestfile")
	f.maxLength("contestname", 45)
	f.dateCheck("startdate")
	f.dateCheck("enddate")
	f.timeCheck("starttime")
	f.timeCheck("endtime")
	f.fileCheck("contestfile")

	start := f.datetimeFormat("startdate", "starttime")
	end := f.datetimeFormat("enddate", "endtime")
	if !f.valid() {
		td.FormData = f
		app.render(w, r, "cabrillo.page.html", td)
		return
	}

	cd := &contestData{
		filename:  filepath.Join(app.contestDir, f.Get("contestfile")),
		name:      f.Get("contestname"),
		startTime: start,
		endTime:   end,
		score:     "0",
	}
	rows, err := app.logsModel.getCabrilloData(cd)
	if err != nil {
		app.serverError(w, err)
		return
	}
	err = app.genCabrilloFile(rows, cd)
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Table = rows
	app.render(w, r, "cabrillo.page.html", td)
}

func (app *application) genCabrilloNew(w http.ResponseWriter, r *http.Request) {
	var err error
	td := initTemplateData()

	err = r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	f := newForm(r.PostForm)
	f.required("contestname", "contestfile", "callWidth")
	f.isInt("callWidth")
	f.maxLength("contestname", 45)
	f.fileCheck("contestfile")
	fc, err := app.otherModel.getDefault("fieldCount")
	if err != nil {
		app.serverError(w, err)
		return
	}
	cnt, err := strconv.Atoi(fc)
	if err != nil {
		app.serverError(w, err)
		return
	}
	for i := 0; i < cnt; i++ {
		q := fmt.Sprintf("field%dName", i+1)
		sfn, err := app.otherModel.getDefault(q)
		if err != nil {
			app.serverError(w, err)
		}
		td.FieldNames = append(td.FieldNames, sfn)
	}

	for i := 0; i < cnt; i++ {
		fW := fmt.Sprintf("field%dWidth", i)
		f.required(fW)
		f.isInt(fW)
	}

	if !f.valid() {
		td.FormData = f
		app.render(w, r, "cabrillo.page.html", td)
		return
	}
	cData, err := app.contestModel.getContest(f.Get("contestname"))
	if err != nil {
		if errors.Is(err, errNoRecord) {
			td.Message = fmt.Sprintf("contest name %s does not exist", f.Get("contestname"))
			app.render(w, r, "cabrillo.page.html", td)
			return
		}
		app.serverError(w, err)
		return
	}
	w0, _ := strconv.Atoi(f.Get("callWidth"))
	w1, _ := strconv.Atoi(f.Get("field0Width"))
	w2, _ := strconv.Atoi(f.Get("field1Width"))
	w3, _ := strconv.Atoi(f.Get("field2Width"))
	w4, _ := strconv.Atoi(f.Get("field3Width"))
	w5, _ := strconv.Atoi(f.Get("field4Width"))

	cd := &contestData{
		filename:    filepath.Join(app.contestDir, f.Get("contestfile")),
		name:        f.Get("contestname"),
		score:       "0",
		fieldCount:  cData.FieldCount,
		callWidth:   w0,
		field1Width: w1,
		field2Width: w2,
		field3Width: w3,
		field4Width: w4,
		field5Width: w5,
	}
	td.FieldCount = cData.FieldCount
	td.Top.Field1Name = cData.Field1Name
	td.Top.Field2Name = cData.Field2Name
	td.Top.Field3Name = cData.Field3Name
	td.Top.Field4Name = cData.Field4Name
	td.Top.Field5Name = cData.Field5Name

	rows, err := app.logsModel.getNewCabrilloData(cd)
	if err != nil {
		app.serverError(w, err)
		return
	}
	err = app.genNewCabrilloFile(rows, cd)
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Table = rows
	app.render(w, r, "cabrillo.page.html", td)
}

func (app *application) analysis(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	t, err := app.logsModel.getLatestLogs(1000000)
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Stats.Contacts = len(t)
	t, err = app.logsModel.getConfirmedContacts()
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Stats.ConfirmedContacts = len(t)
	t, err = app.qrzModel.getRepeatContacts()
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Stats.RepeatContacts = len(t)
	t, err = app.logsModel.getUniqueCountries()
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Stats.Country = len(t)
	t, err = app.logsModel.getConfirmedCountries()
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Stats.ConfirmedCountry = len(t)
	t, err = app.qrzModel.getUniqueStates()
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Stats.State = len(t)
	t, err = app.logsModel.getConfirmedStates()
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Stats.ConfirmedState = len(t)

	t, err = app.qrzModel.getUniqueCounties()
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Stats.County = len(t)
	t, err = app.logsModel.getConfirmedCounties()
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Stats.ConfirmedCounty = len(t)
	t, err = app.logsModel.getSimpleLogs("CW", "%", "%")
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Stats.CWContacts = len(t)
	t, err = app.logsModel.getSimpleLogs("CW", "YES", "%")
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Stats.ConfirmedCW = len(t)
	t, err = app.logsModel.getUniqueCountry("CW", "YES")
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Stats.ConfirmedCWCountry = len(t)
	t, err = app.logsModel.getUniqueState("CW", "YES")
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Stats.ConfirmedCWState = len(t)
	app.render(w, r, "analysis.page.html", td)
}

func (app *application) country(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	t, err := app.logsModel.getUniqueCountries()
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Table = t
	app.render(w, r, "country.page.html", td)
}

func (app *application) countryConfirmed(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	t, err := app.logsModel.getConfirmedCountries()
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Table = t
	app.render(w, r, "country.page.html", td)
}

func (app *application) state(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	t, err := app.qrzModel.getUniqueStates()
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Table = t
	app.render(w, r, "state.page.html", td)
}

func (app *application) stateConfirmed(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	t, err := app.logsModel.getConfirmedStates()
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Table = t
	app.render(w, r, "state.page.html", td)
}

func (app *application) contactsConfirmed(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	t, err := app.logsModel.getConfirmedContacts()
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Table = t
	app.render(w, r, "log.page.html", td)
}

func (app *application) repeat(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	t, err := app.qrzModel.getRepeatContacts()
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Table = t
	app.render(w, r, "log.page.html", td)
}

func (app *application) countrySelect(w http.ResponseWriter, r *http.Request) {

	td := initTemplateData()
	country := r.URL.Query().Get("sel")
	t, err := app.logsModel.getLogsByCountry(country)
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Table = t
	app.render(w, r, "log.page.html", td)
}

func (app *application) county(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	t, err := app.qrzModel.getUniqueCounties()
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Table = t
	app.render(w, r, "county.page.html", td)
}

func (app *application) countyConfirmed(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	t, err := app.logsModel.getConfirmedCounties()
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Top.Cnty = true
	td.Table = t
	app.render(w, r, "county.page.html", td)
}

func (app *application) countySelect(w http.ResponseWriter, r *http.Request) {

	td := initTemplateData()
	county := r.URL.Query().Get("sel")

	t, err := app.logsModel.getLogsByCounty(county)
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Table = t
	td.Top.Cnty = true
	app.render(w, r, "log.page.html", td)
}

func (app *application) stateSelect(w http.ResponseWriter, r *http.Request) {

	td := initTemplateData()
	state := r.URL.Query().Get("sel")

	t, err := app.logsModel.getLogsByState(state)
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Table = t
	td.Top.Cnty = true
	app.render(w, r, "log.page.html", td)
}

func (app *application) cwContacts(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	t, err := app.logsModel.getSimpleLogs("CW", "%", "%")
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Table = t
	app.render(w, r, "log.page.html", td)

}

func (app *application) cwConfirmed(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	t, err := app.logsModel.getSimpleLogs("CW", "YES", "%")
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Table = t
	app.render(w, r, "log.page.html", td)

}

func (app *application) cwConfirmedState(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	t, err := app.logsModel.getUniqueState("CW", "YES")
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Table = t
	//	for _, item := range td.Table {
	//		fmt.Println(item.State)
	//	}
	app.render(w, r, "state.page.html", td)

}

func (app *application) cwConfirmedCountry(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	t, err := app.logsModel.getUniqueCountry("CW", "YES")
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Table = t
	app.render(w, r, "country.page.html", td)

}
