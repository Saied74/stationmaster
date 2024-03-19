package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	//	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
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

func initTemplateData() *templateData {
	return &templateData{
		FormData: newForm(url.Values{}),
		Speed:    speed,
		Tone:     tone,
		Volume:   volume,
		Top:      tableHead,
		Table:    []LogsRow{},
		LogEdit:  &LogsRow{},
		Show:     false,
		Edit:     false,
		StopCode: false,
		Logger:   false,
		Contest:  "No",
		Stats:    &Stats{},
		VFO:      &VFO{},
	}
}

func copyPostForm(r *http.Request) LogsRow {
	return LogsRow{
		Call:     strings.ToUpper(r.PostForm.Get("call")),
		Sent:     r.PostForm.Get("sent"),
		Rcvd:     r.PostForm.Get("rcvd"),
		Band:     strings.ToLower(r.PostForm.Get("band")),
		Mode:     r.PostForm.Get("mode"), //m,
		Name:     r.PostForm.Get("name"),
		Country:  r.PostForm.Get("country"),
		Comment:  r.PostForm.Get("comment"),
		Lotwsent: r.PostForm.Get("lotwsent"),
		Lotwrcvd: r.PostForm.Get("lotwrcvd"),
		ExchSent: r.PostForm.Get("exchsent"),
		ExchRcvd: r.PostForm.Get("exchrcvd"),
	}
}

//<+++++++++++++++++++++    Default Handling   ++++++++++++++++++++++++>

func (app *application) lookupDefault(def string) (string, error) {
	v, err := app.otherModel.getDefault(def)
	if err != nil {
		if errors.Is(err, errNoRecord) {
			switch def {
			case "mode":
				v = "USB"
			case "band":
				v = "20m"
			case "contest":
				v = "No"
			case "contestname":
				v = "NJQP"
			case "sent":
				v = "59"
			case "exchange":
				v = "MIDD"
			default:
				return "", fmt.Errorf("Bad string passed to lookupDefaults")
			}
		} else {
			return "", err
		}
	}
	return v, nil
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
			f.Errors.add(field, "this field cannot be blank")
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

func (f *formData) checkAllLogMax() {
	f.maxLength("call", 10)
	f.maxLength("sent", 3)
	f.maxLength("rcvd", 3)
	f.maxLength("band", 8)
	f.maxLength("name", 85)
	f.maxLength("country", 85)
	f.maxLength("comment", 85)
	f.maxLength("lotwrcvd", 10)
	f.maxLength("lotwsent", 10)
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

func (f *formData) dateCheck(field string) {
	value := f.Get(field)
	bits := strings.Split(value, "-")
	if len(bits) != 3 {
		f.Errors.add(field, "incorrect date format")
		return
	}
	if len(bits[0]) != 4 {
		f.Errors.add(field, "incorrect year format")
	}
	if len(bits[1]) != 2 {
		f.Errors.add(field, "incorrect month format")
	}
	if len(bits[2]) != 2 {
		f.Errors.add(field, "incorrect day forrmat")
	}
	return
}

func (f *formData) timeCheck(field string) {
	value := f.Get(field)
	bits := strings.Split(value, ":")
	if len(bits) != 2 {
		f.Errors.add(field, "incorrect time format")
		return
	}
	if len(bits[0]) != 2 {
		f.Errors.add(field, "incorrect hour format")
	}
	if len(bits[1]) != 2 {
		f.Errors.add(field, "incorrect minute format")
	}
	return
}

func (f *formData) datetimeFormat(d, t string) time.Time {
	vD := f.Get(d)
	vT := f.Get(t)
	tt, err := time.Parse(time.RFC3339, vD+"T"+vT+":00Z")
	if err != nil {
		f.Errors.add(d, "date and time values did not yield a valid date")
		return time.Time{}
	}
	return tt

}

func (f *formData) fileCheck(d string) {
	v := f.Get(d)
	p := strings.Split(v, ".")
	if len(p) != 2 {
		f.Errors.add(d, "filename must be two parts seperated by .")
		return
	}
	if p[1] != "txt" {
		f.Errors.add(d, "second part of the filename must be txt")
	}
}

func (f *formData) valid() bool {
	return len(f.Errors) == 0
}

// extracts the floating value of the keyer fields
func (f *formData) extractFloat(field, c string, fc float64) (float64, string) {
	value := f.Get(field)

	//if the field is blank, use the example numbers
	if strings.TrimSpace(value) == "" {
		return fc, c
	}
	//if it is not, convert it to float64 and process any error
	s, err := strconv.ParseFloat(value, 64)
	if err != nil {
		f.Errors.add(field, "Sending speed must be a number")
		return 0, value
	}
	return s, value
}

func (f *formData) extractCWParameter(field string) (int, error) {
	value := f.Get(field)
	if strings.TrimSpace(value) == "" {
		return 0, fmt.Errorf("Balnk %v field", field)
	}
	s, err := strconv.Atoi(value)
	if err != nil {
		f.Errors.add(field, fmt.Sprintf("%v field must be a number", field))
		return 0, fmt.Errorf("%v field must be a number", field)
	}
	return s, nil
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

//<+++++++++++++++++   Context and ID Mgmt.   ++++++++++++++++++++>

type putCancelFunc func(context.Context, context.CancelFunc, bool)
type getCancelFunc func() (context.Context, context.CancelFunc, bool)

func contextStore() (putCancelFunc, getCancelFunc) {
	var ktutor bool
	var canFunc context.CancelFunc
	var ctx context.Context
	putCF := func(cx context.Context, cf context.CancelFunc, kt bool) {
		ktutor = kt
		canFunc = cf
		ctx = cx
	}
	getCF := func() (context.Context, context.CancelFunc, bool) {
		return ctx, canFunc, ktutor
	}
	return putCF, getCF
}

type putIdFunc func(int)
type getIdFunc func() int

func saveId() (putId putIdFunc, getId getIdFunc) {
	var id int
	put := func(n int) {
		id = n
	}
	get := func() int {
		return id
	}
	return put, get
}

//<++++++++++++++++++++++++  VFO  ++++++++++++++++++++++++++++++>

func (app *application) getVFOUpdate() (*VFO, error) {
	band, err := app.otherModel.getDefault("band")
	if err != nil {
		fmt.Println("1")
		return &VFO{}, err
	}
	v := vfoMemory[band]
	v.Band = band
	mode, err := app.otherModel.getDefault("mode")
	if err != nil {
		fmt.Println("2")
		return &VFO{}, err
	}
	v.Mode = mode
	x := band + "xfreq"
	xfreq, err := app.otherModel.getDefault(x)
	if err != nil {
		fmt.Println(x, "3")
		return &VFO{}, err
	}
	v.XFreq = xfreq
	r := band + "rfreq"
	rfreq, err := app.otherModel.getDefault(r)
	if err != nil {
		fmt.Println("4")
		return &VFO{}, err
	}
	v.RFreq = rfreq
	split, err := app.otherModel.getDefault("split")
	if err != nil {
		fmt.Println("5")
		return &VFO{}, err
	}
	v.Split = split
	return v, nil
}

func (app *application) pickZone(zone string, dxData []DXClusters) ([]DXClusters, error) {
	newData := []DXClusters{}
	i := 0
	j := 0
	for _, dx := range dxData {

		inUS := strings.HasPrefix(dx.DE, "K") ||
			strings.HasPrefix(dx.DE, "W") ||
			strings.HasPrefix(dx.DE, "N") ||
			strings.HasPrefix(dx.DE, "A")

		inZone := strings.Contains(dx.DE, "1") ||
			strings.Contains(dx.DE, "2") ||
			strings.Contains(dx.DE, "3") ||
			strings.Contains(dx.DE, "4")
		if inUS && inZone {
			newData = append(newData, dx)
			i++
		}
		if i == maxDXLines-1 {
			break
		}
		j++
	}
	return newData, nil

}

var noBandUpdate = errors.New("no band update")

func (app *application) getUpdateBand() (*VFO, error) {
	var b int
	var dx = []DXClusters{}
	v, err := app.getVFOUpdate() //populate VFO from dB
	if err != nil {
		return &VFO{}, err
	}
	b, err = app.readBand()
	if err != nil {
		return v, err
	}
	//b = bandselect.BandRead(app.bandData)
	update, ok := switchTable[b]
	if !ok {
		return &VFO{}, fmt.Errorf("bad data from the switch %d", b)
	}
	if v.Band != update.Band {
		err = app.otherModel.updateDefault("band", update.Band)
		if err != nil {
			return v, err
		}

		err = app.changeBand(update.Band)
		if err != nil {
			err = app.spiderError(err)
			if err != nil {
				return &VFO{}, err
			}
		}
		v, err := app.getVFOUpdate() //populate VFO from dB
		if err != nil {
			return &VFO{}, err
		}
		dx, err = app.getSpider(v.Band, dxLines)
		if err != nil {
			err = app.spiderError(err)
			if err != nil {
				return &VFO{}, err
			}
			dx, err = app.getSpider(v.Band, dxLines)
			if err != nil {
				return &VFO{}, err
			}
		}
		v.DX = dx
		v.Band = update.Band
		return v, nil
	}
	return v, nil
}

func (app *application) getUpdateMode(p *VFO) error {

	xf := p.Band + "xfreq"
	xFreq, err := app.otherModel.getDefault(xf)
	if err != nil {
		return err
	}
	if xFreq <= vfoMemory[p.Band].CWBoundary {
		switch xFreq {
		case vfoMemory[p.Band].FT4Freq:
			p.Mode = "FT4"
		case vfoMemory[p.Band].FT8Freq:
			p.Mode = "FT8"
		default:
			p.Mode = "CW"
		}
	} else {
		switch p.Band {
		case "10m":
			p.Mode = "USB"
		case "15m":
			p.Mode = "USB"
		case "20m":
			p.Mode = "USB"
		case "40m":
			p.Mode = "LSB"
		case "80m":
			p.Mode = "LSB"
		case "160m":
			p.Mode = "LSB"
		default:
			p.Mode = "No transmission"
		}
	}
	err = app.otherModel.updateDefault("mode", p.Mode)
	if err != nil {
		return err
	}
	return nil
}
