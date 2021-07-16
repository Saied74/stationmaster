package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestQSOLog(t *testing.T) {
	app := newTestApp()
	row := LogsRow{
		Id:   10,
		Call: "AD2CC",
		Mode: "USB",
		Sent: "59",
		Rcvd: "48",
	}

	tests := []struct {
		name     string
		model    *fakeModel
		wantCode int
	}{
		{"good20", &fakeModel{row, []LogsRow{row, row}, nil, nil, "20m", "USB"}, 200},
		{"good40", &fakeModel{row, []LogsRow{row, row}, nil, nil, "40m", "LSB"}, 200},
		{"badLastLogs", &fakeModel{row, []LogsRow{row, row}, errTest, nil, "20m", "USB"}, 500},
		{"badDefault", &fakeModel{row, []LogsRow{row, row}, nil, errTest, "20m", "USB"}, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app.stationModel = tt.model
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
				if !bytes.Contains(body, []byte(tt.model.band)) {
					t.Errorf("expected body to contain %s, did not get", tt.model.band)
				}
				if !bytes.Contains(body, []byte(tt.model.mode)) {
					t.Errorf("expected body to contain %s, did not get", tt.model.mode)
				}
			}
		})
	}
}

func TestAddLog(t *testing.T) {
	app := newTestApp()
	row := LogsRow{
		Id:   10,
		Call: "AD2CC",
		Mode: "USB",
		Sent: "59",
		Rcvd: "488888",
	}

	tests := []struct {
		name     string
		model    *fakeModel
		wantCode int
	}{
		{"good20", &fakeModel{row, []LogsRow{row, row}, nil, nil, "20m", "USB"}, 200},
		{"good40", &fakeModel{row, []LogsRow{row, row}, nil, nil, "40m", "LSB"}, 200},
		{"badLastLogs", &fakeModel{row, []LogsRow{row, row}, errTest, nil, "20m", "USB"}, 500},
		{"badDefault", &fakeModel{row, []LogsRow{row, row}, nil, errTest, "20m", "USB"}, 500},
	}
	body := strings.NewReader(fmt.Sprintf("call=AD2CC&mode=%s&band=%s", tests[0].model.mode, tests[0].model.band))
	app.stationModel = tests[0].model
	r := httptest.NewRequest(http.MethodPost, "/", body)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	t.Errorf("Request %v\n", r)
	r.ParseForm()
	t.Errorf("parseform %v\n", r.PostForm)
	app.addlog(rr, r)
	rs := rr.Result()
	defer rs.Body.Close()
	// bb, err := io.ReadAll(rs.Body)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// t.Errorf("body %v\n", string(bb))
}
