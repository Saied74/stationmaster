package main

import (
	"fmt"
	"net/http"
)

const (
	radio  = "radio"
	yaesu  = "Yaesu"
	tentec = "Ten Tec"
)

func (app *application) setYaesu(w http.ResponseWriter, r *http.Request) {
	app.otherModel.updateDefault(radio, yaesu)
	err := app.clearPorts()
	if err != nil {
		app.errorLog.Println("failed to clear ports in Yaesu %v", err)
	}
	err = app.classifyRemotes()
	if err != nil {
		app.errorLog.Println("failed to start remote radio in USB %v", err)
	}
	err = app.initRadio()
	if err != nil {
		app.errorLog.Println("failed to initialize radio in USB %v", err)
	}

	//	fmt.Println(yaesu)
	app.defaults(w, r)
}

func (app *application) setTenTec(w http.ResponseWriter, r *http.Request) {
	app.otherModel.updateDefault(radio, tentec)
	fmt.Println(tentec)
	err := app.clearPorts()
	if err != nil {
		app.errorLog.Println("failed to clear ports in Ten Tec %v", err)
	}

	err = app.classifyRemotes()
	if err != nil {
		app.errorLog.Println("failed to start Ten Tec CW in USB %v", err)
	}

	app.defaults(w, r)
}
