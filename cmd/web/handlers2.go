package main

import (
	"net/http"
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

func (app *application) analysis(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
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

func (app *application) repeat(w http.ResponseWriter, r *http.Request) {
	td := initTemplateData()
	app.render(w, r, "repeat.page.html", td)
}

func (app *application) countrySelect(w http.ResponseWriter, r *http.Request) {
	app.infoLog.Printf("countrySelect")
	td := initTemplateData()
	country := r.URL.Query().Get("sel")
	app.infoLog.Printf("Country: %s\n", country)
	t, err := app.logsModel.getLogsByCountry(country)
	if err != nil {
		app.serverError(w, err)
		return
	}
	td.Table = t
	app.render(w, r, "log.page.html", td)
}
