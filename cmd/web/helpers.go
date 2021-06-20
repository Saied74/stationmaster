package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"unicode/utf8"
)

type formErrors map[string][]string

type formData struct {
	url.Values
	Errors formErrors
}

//template caching and rendering are right out of "Let's Go"

func newTemplateCache(dir string) (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob(filepath.Join(dir, "*.page.html"))
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.html"))
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.html"))
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}
	return cache, nil
}

func (app *application) render(w http.ResponseWriter, r *http.Request,
	name string, td *templateData) {
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist",
			name))
		return
	}
	buf := new(bytes.Buffer)
	err := ts.Execute(buf, td)
	if err != nil {
		app.serverError(w, err)
		return
	}
	buf.WriteTo(w)
}

//to do: I will have to find a better way to hold this data and also
//the context of the session - perhaps after I install mySQL
func getBaseTemp() *templateData {
	return &templateData{
		Speed:     speed,
		FarnSpeed: farnspeed,
		Lsm:       lsm,
		Wsm:       wsm,
	}
}

//<+++++++++++++++++++++  Form Error Handling  ++++++++++++++++++++++++>

func (e formErrors) add(field, message string) {
	e[field] = append(e[field], message)
}

func (e formErrors) Get(field string) string {
	es, ok := e[field]
	if !ok || len(es) == 0 {
		return ""
	}
	return es[0]
}

func newForm(data url.Values) *formData {
	return &formData{
		data,
		formErrors(map[string][]string{}),
	}
}

func (f *formData) required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.add(field, "this field cannoot be blank")
		}
	}
}

func (f *formData) maxLength(field string, d int) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if utf8.RuneCountInString(value) > d {
		f.Errors.add(field, fmt.Sprintf(`this field is too long 
		(maximum is %d characters)`, d))
	}
}

func (f *formData) minLength(field string, d int) {
	value := f.Get(field)
	if utf8.RuneCountInString(value) < d {
		f.Errors.add(field, fmt.Sprintf(`this field is too short 
		(minimum is %d characters)`, d))
	}
}

func (f *formData) isInt(field string) {
	value := f.Get(field)
	_, err := strconv.Atoi(value)
	if err != nil {
		f.Errors.add(field, "this field must be integers")
	}
}

func (f *formData) mustFloat(field string) float64 {
	value := f.Get(field)
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		f.Errors.add(field, "this field must be numbers")
		return 0
	}
	return num
}

func (f *formData) valid() bool {
	return len(f.Errors) == 0
}

//extracts the floating value of the keyer fields
func (f *formData) extractFloat(field, c string, fc float64) (float64, string) {
	value := f.Get(field)

	//if the field is blank, use the example numbers
	if strings.TrimSpace(value) == "" {
		return fc, c
	} else {
		//if it is not, convert it to float64 and process any error
		s, err := strconv.ParseFloat(value, 64)
		if err != nil {
			f.Errors.add(field, "Sending speed must be a number")
			return 0, value
		}
		return s, value
	}
}

//<++++++++++++++++   centralized error handling   +++++++++++++++++++>

func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace) //to not get the helper file...

	http.Error(w, http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}
