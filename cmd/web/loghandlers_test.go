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
		{"good20", &mockModal{band: "20m", mode: "FOO"}, nil, nil, 200},
		{"good40", &mockModal{band: "40m", mode: "FOO"}, nil, nil, 200},
		{"badLastLogs", &mockModal{band: "20m", mode: "FOO"}, errTest, nil, 500},
		{"badDefault", &mockModal{band: "20m", mode: "FOO"}, nil, errTest, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app.logsModel = &mockLogsModel{lastLogsErr: tt.errInput1}
			app.otherModel = &mockOtherModel{defaultErr: tt.errInput2}
			switch tt.name {
			case "good20":
				testToggle = 1
			case "good40":
				testToggle = 2
			default:
				testToggle = 1
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
