package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
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
