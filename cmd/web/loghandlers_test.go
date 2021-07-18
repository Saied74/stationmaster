package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockModal struct {
	call     string
	band     string
	mode     string
	sent     string
	rcvd     string
	name     string
	country  string
	comment  string
	lotwsent string
	lotwrcvd string
}

func TestQSOLog(t *testing.T) {
	app := newTestApp()

	tests := []struct {
		name      string
		dataInput *mockModal
		errInput1 error
		errInput2 error
		wantCode  int
	}{
		//mockModal is used here to inject test data, not in its functional role
		{"good20", &mockModal{band: "20m", mode: "BARFOO"}, nil, nil, 200},
		{"good40", &mockModal{band: "40m", mode: "FOOBAR"}, nil, nil, 200},
		{"badLastLogs", &mockModal{band: "20m", mode: "FOO"}, errTest, nil, 500},
		{"badDefault", &mockModal{band: "20m", mode: "FOO"}, nil, errTest, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := LogsRow{
				Band: tt.dataInput.band,
				Mode: tt.dataInput.mode,
			}
			app.logsModel = &mockLogsModel{
				rows:        []LogsRow{row, row},
				lastLogsErr: tt.errInput1,
				mode:        tt.dataInput.mode,
				band:        tt.dataInput.band,
			}
			app.otherModel = &mockOtherModel{
				defaultErr: tt.errInput2,
				mode:       tt.dataInput.mode,
				band:       tt.dataInput.band,
			}
			rr := httptest.NewRecorder()
			r, err := http.NewRequest(http.MethodGet, "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			app.qsolog(rr, r)
			rs := rr.Result()
			defer rs.Body.Close()
			if rs.StatusCode != tt.wantCode {
				t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
			}
			body, err := io.ReadAll(rs.Body)
			if err != nil {
				t.Fatal(err)
			}
			if rs.StatusCode == http.StatusOK {
				if !bytes.Contains(body, []byte(tt.dataInput.band)) {
					t.Errorf("expected body to contain %s, did not get", tt.dataInput.band)
				}
				if !bytes.Contains(body, []byte(tt.dataInput.mode)) {
					t.Errorf("expected body to contain %s, did not get", tt.dataInput.mode)
				}
			}
		})
	}
}

func TestAddLog(t *testing.T) {
	app := newTestApp()
	r := ioutil.NopCloser(bytes.NewReader([]byte(xmlTestData)))
	client = newMockClient(200, r)
	tests := []struct {
		name      string
		dataInput *mockModal
		errInput1 error
		errInput2 error
		wantCode  int
	}{
		{"bad long", &mockModal{"AD2CC", "99m", "FOO", "797690", "880", "Dr. Who", "UK", "Tardis?", "No", "No"}, nil, nil, 200},
		{"bad missing", &mockModal{"", "99m", "FOO", "79", "880", "Dr. Who", "UK", "Tardis?", "No", "No"}, nil, nil, 200},
		{"bad short", &mockModal{"AD2CC", "99m", "FOO", "7", "880", "Dr. Who", "UK", "Tardis?", "No", "No"}, nil, nil, 200},
		{"bad int", &mockModal{"AD2CC", "99m", "FOO", "abc", "880", "Dr. Who", "UK", "Tardis?", "No", "No"}, nil, nil, 200},
		{"good", &mockModal{"AD2CC", "99m", "FOO", "79", "880", "Dr. Who", "UK", "Tardis?", "No", "No"}, nil, nil, 200},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(fmt.Sprintf(
				"call=%s&band=%s&mode=%s&sent=%s&rcvd=%s&name=%s&country=%s&comment=%s&lotwsent=%s&lotwrcvd=%s",
				tt.dataInput.call, tt.dataInput.band, tt.dataInput.mode, tt.dataInput.sent,
				tt.dataInput.rcvd, tt.dataInput.name, tt.dataInput.country,
				tt.dataInput.comment, tt.dataInput.lotwsent, tt.dataInput.lotwrcvd))
			r, err := http.NewRequest(http.MethodPost, "/", body)
			if err != nil {
				t.Fatal(err)
			}
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			row := LogsRow{
				Band: tt.dataInput.band,
				Mode: tt.dataInput.mode,
			}
			app.logsModel = &mockLogsModel{
				rows: []LogsRow{row, row},
				band: tt.dataInput.band,
				mode: tt.dataInput.mode}
			rr := httptest.NewRecorder()
			app.addlog(rr, r)
			rs := rr.Result()
			defer rs.Body.Close()
			bod, err := io.ReadAll(rs.Body)
			if err != nil {
				t.Fatal(err)
			}
			switch i {
			case 0:
				if rs.StatusCode != tt.wantCode {
					t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
				}
				if !bytes.Contains(bod, []byte("this field is too long")) {
					t.Errorf("expected body to contain this field is too long, did not get")
				}
			case 1:
				if rs.StatusCode != tt.wantCode {
					t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
				}
				if !bytes.Contains(bod, []byte("this field cannoot be blank")) {
					t.Errorf("expected body to contain this field cannoot be blank, did not get")
				}
			case 2:
				if rs.StatusCode != tt.wantCode {
					t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
				}
				if !bytes.Contains(bod, []byte("this field is too short")) {
					t.Errorf("expected body to contain this field is too short, did not get")
				}
			case 3:
				if rs.StatusCode != tt.wantCode {
					t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
				}
				if !bytes.Contains(bod, []byte("this field must be integers")) {
					t.Errorf("expected body to contain this field must be integers, did not get")
				}
			case 4:
				if rs.StatusCode != tt.wantCode {
					t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
				}
				if !bytes.Contains(bod, []byte("FOO")) {
					t.Errorf("expected body to contain FOO, did not get")
				}
			}
		})
	}
}

func TestEditLog(t *testing.T) {
	app := newTestApp()
	putId, getId := saveId()
	app.putId = putId
	app.getId = getId
	tests := []struct {
		name      string
		dataInput string
		errInput1 error
		errInput2 error
		wantCode  int
	}{
		{"bad int", "abc", nil, nil, 400},
		{"bad logById", "10", nil, nil, 500},
		{"good", "1", errTest, nil, 200},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/foo?id=%s", tt.dataInput)
			r, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				t.Fatal(err)
			}
			app.editlog(rr, r)
			rs := rr.Result()
			defer rs.Body.Close()
			bod, err := io.ReadAll(rs.Body)
			if err != nil {
				t.Fatal(err)
			}
			switch i {
			case 0:
				if rs.StatusCode != tt.wantCode {
					t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
				}
			case 1:
				if rs.StatusCode != tt.wantCode {
					t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
				}
			case 2:
				if rs.StatusCode != tt.wantCode {
					t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
				}
				if !bytes.Contains(bod, []byte("FOOBAR")) {
					t.Errorf("expected body to FOOBAR, did not get")
				}
			}
		})
	}
}

func TestUpdateDB(t *testing.T) {
	app := newTestApp()
	putId, getId := saveId()
	app.putId = putId
	app.getId = getId
	r := ioutil.NopCloser(bytes.NewReader([]byte(xmlTestData)))
	client = newMockClient(200, r)
	tests := []struct {
		name      string
		dataInput *mockModal
		errInput1 error
		errInput2 error
		wantCode  int
	}{
		{"bad long", &mockModal{"AD2CC", "99m", "FOO", "797690", "880", "Dr. Who", "UK", "Tardis?", "No", "No"}, nil, nil, 200},
		{"bad missing", &mockModal{"", "99m", "FOO", "79", "880", "Dr. Who", "UK", "Tardis?", "No", "No"}, nil, nil, 200},
		{"bad short", &mockModal{"AD2CC", "99m", "FOO", "7", "880", "Dr. Who", "UK", "Tardis?", "No", "No"}, nil, nil, 200},
		{"bad int", &mockModal{"AD2CC", "99m", "FOO", "abc", "880", "Dr. Who", "UK", "Tardis?", "No", "No"}, nil, nil, 200},
		{"good", &mockModal{"AD2CC", "99m", "FOO", "79", "880", "Dr. Who", "UK", "Tardis?", "No", "No"}, nil, nil, 200},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(fmt.Sprintf(
				"call=%s&band=%s&mode=%s&sent=%s&rcvd=%s&name=%s&country=%s&comment=%s&lotwsent=%s&lotwrcvd=%s",
				tt.dataInput.call, tt.dataInput.band, tt.dataInput.mode, tt.dataInput.sent,
				tt.dataInput.rcvd, tt.dataInput.name, tt.dataInput.country,
				tt.dataInput.comment, tt.dataInput.lotwsent, tt.dataInput.lotwrcvd))
			r, err := http.NewRequest(http.MethodPost, "/", body)
			if err != nil {
				t.Fatal(err)
			}
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			row := LogsRow{
				Band: tt.dataInput.band,
				Mode: tt.dataInput.mode,
			}
			app.logsModel = &mockLogsModel{
				rows: []LogsRow{row, row},
				band: tt.dataInput.band,
				mode: tt.dataInput.mode}
			rr := httptest.NewRecorder()
			app.updatedb(rr, r)
			rs := rr.Result()
			defer rs.Body.Close()
			bod, err := io.ReadAll(rs.Body)
			if err != nil {
				t.Fatal(err)
			}
			switch i {
			case 0:
				if rs.StatusCode != tt.wantCode {
					t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
				}
				if !bytes.Contains(bod, []byte("this field is too long")) {
					t.Errorf("expected body to contain this field is too long, did not get")
				}
			case 1:
				if rs.StatusCode != tt.wantCode {
					t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
				}
				if !bytes.Contains(bod, []byte("this field cannoot be blank")) {
					t.Errorf("expected body to contain this field cannoot be blank, did not get")
				}
			case 2:
				if rs.StatusCode != tt.wantCode {
					t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
				}
				if !bytes.Contains(bod, []byte("this field is too short")) {
					t.Errorf("expected body to contain this field is too short, did not get")
				}
			case 3:
				if rs.StatusCode != tt.wantCode {
					t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
				}
				if !bytes.Contains(bod, []byte("this field must be integers")) {
					t.Errorf("expected body to contain this field must be integers, did not get")
				}
			case 4:
				if rs.StatusCode != tt.wantCode {
					t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
				}
				if !bytes.Contains(bod, []byte("FOO")) {
					t.Errorf("expected body to contain FOO, did not get")
				}
			}
		})
	}
}

func TestGetConn(t *testing.T) {
	app := newTestApp()
	putId, getId := saveId()
	app.putId = putId
	app.getId = getId
	r := ioutil.NopCloser(bytes.NewReader([]byte(xmlTestData)))
	client = newMockClient(200, r)
	tests := []struct {
		name      string
		dataInput string
		errInput1 error
		errInput2 error
		wantCode  int
	}{
		{"empty call", "", nil, nil, 200},
		{"good call", "AD2CC", nil, nil, 200},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/foo?call=%s", tt.dataInput)
			r, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				t.Fatal(err)
			}
			app.getConn(rr, r)
			rs := rr.Result()
			defer rs.Body.Close()
			bod, err := io.ReadAll(rs.Body)
			if err != nil {
				t.Fatal(err)
			}
			switch i {
			case 0:
				if rs.StatusCode != tt.wantCode {
					t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
				}
			case 1:
				if rs.StatusCode != tt.wantCode {
					t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
				}
				if !bytes.Contains(bod, []byte("FOO BAR")) {
					t.Errorf("expected body to contain FOO BAR, did not get")
				}
			}
		})
	}
}

func TestCallSearch(t *testing.T) {
	app := newTestApp()
	putId, getId := saveId()
	app.putId = putId
	app.getId = getId
	r := ioutil.NopCloser(bytes.NewReader([]byte(xmlTestData)))
	client = newMockClient(200, r)
	tests := []struct {
		name      string
		dataInput string
		errInput1 error
		errInput2 error
		wantCode  int
	}{
		{"empty call", "", nil, nil, 200},
		{"good call", "AD2CC", nil, nil, 200},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/foo?call=%s", tt.dataInput)
			r, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				t.Fatal(err)
			}
			app.callSearch(rr, r)
			rs := rr.Result()
			defer rs.Body.Close()
			bod, err := io.ReadAll(rs.Body)
			if err != nil {
				t.Fatal(err)
			}
			switch i {
			case 0:
				if rs.StatusCode != tt.wantCode {
					t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
				}
			case 1:
				if rs.StatusCode != tt.wantCode {
					t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
				}
				if !bytes.Contains(bod, []byte("FOO BAR")) {
					t.Errorf("expected body to contain FOO BAR, did not get")
				}
				if !bytes.Contains(bod, []byte("FOOFOO")) {
					t.Errorf("expected body to contain FOOFOO, did not get")
				}
				if !bytes.Contains(bod, []byte("1952")) {
					t.Errorf("expected body to contain 1952 BAR, did not get")
				}
				if !bytes.Contains(bod, []byte("EST")) {
					t.Errorf("expected body to contain EST, did not get")
				}
			}
		})
	}
}

func TestUpdateQRZ(t *testing.T) {
	app := newTestApp()
	putId, getId := saveId()
	app.putId = putId
	app.getId = getId
	r := ioutil.NopCloser(bytes.NewReader([]byte(xmlTestData)))
	client = newMockClient(200, r)
	tests := []struct {
		name       string
		dataInput1 string
		dataInput2 string
		dataInput3 string
		dataInput4 string
		errInput1  error
		errInput2  error
		wantCode   int
	}{
		{"empty call", "", "", "", "", nil, nil, 200},
		{"good call", "FOO BAR", "FOOFOO", "1952", "EST", nil, nil, 200},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := LogsRow{
				Call:    tt.dataInput1,
				Band:    tt.dataInput2,
				Mode:    tt.dataInput3,
				Country: tt.dataInput4,
			}
			app.logsModel = &mockLogsModel{rows: []LogsRow{row, row}}
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/foo?call=%s", tt.dataInput1)
			r, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				t.Fatal(err)
			}
			app.updateQRZ(rr, r)
			rs := rr.Result()
			defer rs.Body.Close()
			bod, err := io.ReadAll(rs.Body)
			if err != nil {
				t.Fatal(err)
			}
			switch i {
			case 0:
				if rs.StatusCode != tt.wantCode {
					t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
				}
			case 1:
				if rs.StatusCode != tt.wantCode {
					t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
				}
				if !bytes.Contains(bod, []byte("FOO BAR")) {
					t.Errorf("expected body to contain FOO BAR, did not get")
				}
				if !bytes.Contains(bod, []byte("FOOFOO")) {
					t.Errorf("expected body to contain FOOFOO, did not get")
				}
				if !bytes.Contains(bod, []byte("1952")) {
					t.Errorf("expected body to contain 1952 BAR, did not get")
				}
				if !bytes.Contains(bod, []byte("EST")) {
					t.Errorf("expected body to contain EST, did not get")
				}
			}
		})
	}
}

func TestDefaults(t *testing.T) {
	app := newTestApp()
	putId, getId := saveId()
	app.putId = putId
	app.getId = getId
	r := ioutil.NopCloser(bytes.NewReader([]byte(xmlTestData)))
	client = newMockClient(200, r)
	tests := []struct {
		name       string
		dataInput1 string
		dataInput2 string
		errInput1  error
		errInput2  error
		wantCode   int
	}{
		{"200 meter", "200m", "FT99", nil, nil, 200},
		{"1000 meter", "1000m", "FTnothing", nil, nil, 200},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app.otherModel = &mockOtherModel{band: tt.dataInput1, mode: tt.dataInput2}
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/")
			r, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				t.Fatal(err)
			}
			app.defaults(rr, r)
			rs := rr.Result()
			defer rs.Body.Close()
			bod, err := io.ReadAll(rs.Body)
			if err != nil {
				t.Fatal(err)
			}
			switch i {
			case 0:
				if rs.StatusCode != tt.wantCode {
					t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
				}
				if !bytes.Contains(bod, []byte("200m")) {
					t.Errorf("expected body to contain 200m, did not get")
				}
				if !bytes.Contains(bod, []byte("FT99")) {
					t.Errorf("expected body to contain FT99, did not get")
				}
			case 1:
				if rs.StatusCode != tt.wantCode {
					t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
				}
				if !bytes.Contains(bod, []byte("1000m")) {
					t.Errorf("expected body to contain 1000m, did not get")
				}
				if !bytes.Contains(bod, []byte("FTnothing")) {
					t.Errorf("expected body to contain FTnothing, did not get")
				}
			}
		})
	}
}

func TestStoreDefaults(t *testing.T) {
	app := newTestApp()
	tests := []struct {
		name      string
		dataInput []string
		errInput1 error
		errInput2 error
		wantCode  int
		wantData1 string
		wantData2 string
	}{
		{"USB 10m", []string{"1", "1"}, nil, nil, 200, "USB", "10m"},
		{"USB 20m", []string{"1", "2"}, nil, nil, 200, "USB", "20m"},
		{"LSB 40M", []string{"2", "3"}, nil, nil, 200, "LSB", "40m"},
		{"LSB 80m", []string{"2", "4"}, nil, nil, 200, "LSB", "80m"},
		{"bad CW 160m", []string{"3", "5"}, nil, nil, 200, "CW", ""},
		{"None None", []string{"", "9"}, nil, nil, 200, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(fmt.Sprintf("band=%s&mode=%s", tt.dataInput[0], tt.dataInput[1]))
			r, err := http.NewRequest(http.MethodPost, "/", body)
			if err != nil {
				t.Fatal(err)
			}
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			row := LogsRow{
				Band: tt.dataInput[0],
				Mode: tt.dataInput[1],
			}
			app.logsModel = &mockLogsModel{
				rows: []LogsRow{row, row},
				band: tt.dataInput[0],
				mode: tt.dataInput[1],
			}
			rr := httptest.NewRecorder()
			app.storeDefaults(rr, r)
			rs := rr.Result()
			defer rs.Body.Close()
			bod, err := io.ReadAll(rs.Body)
			if err != nil {
				t.Fatal(err)
			}
			if rs.StatusCode != tt.wantCode {
				t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
			} else {
				if !bytes.Contains(bod, []byte(tt.wantData1)) {
					t.Errorf("expected body to contain %s it did not", tt.wantData1)
				}
				if !bytes.Contains(bod, []byte(tt.wantData2)) {
					t.Errorf("expected body to contain %s it did not", tt.wantData2)
				}
			}
		})
	}
}

func TestContacts(t *testing.T) {
	app := newTestApp()
	putId, getId := saveId()
	app.putId = putId
	app.getId = getId
	app.qrzModel = &mockQRZModel{}
	r := ioutil.NopCloser(bytes.NewReader([]byte(xmlTestData)))
	client = newMockClient(200, r)
	tests := []struct {
		name      string
		dataInput string
		errInput1 error
		errInput2 error
		wantCode  int
	}{
		{"empty call", "", nil, nil, 200},
		{"good call", "AD2CC", nil, nil, 200},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/foo?contact-call=%s", tt.dataInput)
			r, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				t.Fatal(err)
			}
			app.contacts(rr, r)
			rs := rr.Result()
			defer rs.Body.Close()
			bod, err := io.ReadAll(rs.Body)
			if err != nil {
				t.Fatal(err)
			}
			switch i {
			case 0:
				if rs.StatusCode != tt.wantCode {
					t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
				}
			case 1:
				if rs.StatusCode != tt.wantCode {
					t.Errorf("expected %d got %d", tt.wantCode, rs.StatusCode)
				}
				if !bytes.Contains(bod, []byte("FOO BAR")) {
					t.Errorf("expected body to contain FOO BAR, did not get")
				}
				if !bytes.Contains(bod, []byte("FOOFOO")) {
					t.Errorf("expected body to contain FOOFOO, did not get")
				}
				if !bytes.Contains(bod, []byte("1952")) {
					t.Errorf("expected body to contain 1952, did not get")
				}
				if !bytes.Contains(bod, []byte("EST")) {
					t.Errorf("expected body to contain EST, did not get")
				}
			}
		})
	}
}
