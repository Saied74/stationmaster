package main

import (
	"net/http"
	"path/filepath"
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
	f.required("contestname", "starttime", "startdate", "enddate", "endtime", "contestfile")
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
