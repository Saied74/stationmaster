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

func TestHome(t *testing.T) {
	app := newTestApp()
	rr := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "/home", nil)
	if err != nil {
		t.Fatal(err)
	}
	app.home(rr, r)
	rs := rr.Result()
	defer rs.Body.Close()
	if rs.StatusCode != http.StatusOK {
		t.Errorf("expected %d got %d", http.StatusOK, rs.StatusCode)
	}
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(body, []byte("Station Master Project")) {
		t.Errorf("expected body to contain Station Master Project, did not get it")
	}
}

func TestAnt(t *testing.T) {
	app := newTestApp()
	rr := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "/ant", nil)
	if err != nil {
		t.Fatal(err)
	}
	app.ant(rr, r)
	rs := rr.Result()
	defer rs.Body.Close()
	if rs.StatusCode != http.StatusOK {
		t.Errorf("expected %d got %d", http.StatusOK, rs.StatusCode)
	}
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(body, []byte("to improve on the antenna tunning")) {
		t.Errorf("expected body to contain to improve on the antenna tunning, did not get it")
	}
}

func TestKtutor(t *testing.T) {
	app := newTestApp()
	rr := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "/ant", nil)
	if err != nil {
		t.Fatal(err)
	}
	app.ktutor(rr, r)
	rs := rr.Result()
	defer rs.Body.Close()
	if rs.StatusCode != http.StatusOK {
		t.Errorf("expected %d got %d", http.StatusOK, rs.StatusCode)
	}
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(body, []byte("Code Tutor and Keyer")) {
		t.Errorf("expected body to contain Code Tutor and Keyer, did not get it")
	}
}

type keyerPattern struct {
	name      string
	speed     string
	farnspeed string
	wsm       string
	lsm       string
	mode      string
	contains  []string
}

func TestStart(t *testing.T) {
	ktps := []keyerPattern{
		{
			name:      "normal keyer data",
			speed:     "12",
			farnspeed: "18",
			wsm:       "1.2",
			lsm:       "1.5",
			mode:      "1",
			contains:  []string{"12", "18", "1.2", "1.5", "1"},
		},
		{
			name:      "bad data",
			speed:     "abc",
			farnspeed: "18",
			wsm:       "1.g",
			lsm:       "f.3",
			mode:      "90",
			contains:  []string{"Sending speed must be a number"},
		},
	}
	for _, tp := range ktps {
		app := newTestApp()
		t.Run(tp.name, func(t *testing.T) {
			body := strings.NewReader(fmt.Sprintf(
				"speed=%s&farnspeed=%s&wsm=%s&lsm=%s&mode=%s",
				tp.speed, tp.farnspeed, tp.wsm, tp.lsm, tp.mode))
			r, err := http.NewRequest(http.MethodPost, "/", body)
			if err != nil {
				t.Fatal(err)
			}
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()
			app.start(rr, r)
			rs := rr.Result()
			defer rs.Body.Close()
			if rs.StatusCode != http.StatusOK {
				t.Errorf("expected %d got %d", http.StatusOK, rs.StatusCode)
			}
			b, err := io.ReadAll(rs.Body)
			if err != nil {
				t.Fatal(err)
			}
			for _, item := range tp.contains {
				if !bytes.Contains(b, []byte(item)) {
					t.Errorf("expected body to contain %s, did not get it", item)
				}
			}
		})
	}
}
