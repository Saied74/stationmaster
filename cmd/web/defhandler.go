package main

import (
	"fmt"
	"net/http"
)

func (app *application) setYaesu(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Yaesu")
	app.defaults(w, r)
}

func (app *application) setTenTec(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Ten Tec")
	app.defaults(w, r)
}
