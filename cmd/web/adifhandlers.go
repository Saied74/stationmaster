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
