package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"  
)

type Qtype struct {
	XMLName  xml.Name `xml:"QRZDatabase"`
	Callsign []Ctype
	Session  []Stype
}

type Ctype struct {
	XMLName xml.Name `xml:"Callsign"`
	Call    string   `xml:"call"`
	Aliases string   `xml:"aliases"`
	Dxcc    string   `xml:"dxcc"`
	Fname   string   `xml:"fname"`
	Lname   string   `xml:"name"`
	Addr1   string   `xml:"addr1"`
	Addr2   string   `xml:"addr2"`
	State   string   `xml:"state"`
	Zip     string   `xml:"zip"`
	Country string   `xml:"country"`
}

type Stype struct {
	Key   string `xml:"Key"`
	Count string `xml:"Count"`
	Time string `xml:"GMTime"`
}

type httpClient interface{
	Get(url string) (*http.Response, error)
}

var noKey = errors.New("no session id")
var client httpClient

func init(){ 
	client = &http.Client{}
}


func (app *application) getHamInfo(callSign string) (*Qtype, error) {
	key, err := app.sKey("")
	if errors.Is(err, noKey) {
		key, err = loginQRZ(app.qrzuser, app.qrzpw)
		if err != nil {
			return nil, err
		}
		key, err = app.sKey(key)
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("//xmldata.qrz.com/xml/current/?s=%s;callsign=%s", key, callSign)
	result, err := getXML(url)
	if err != nil {
		return nil, err
	}
	v := Qtype{Callsign: []Ctype{}, Session: []Stype{}}
	err = xml.Unmarshal(result, &v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func loginQRZ(user, pwd string) (string, error) {
	url := fmt.Sprintf("//xmldata.qrz.com/xml/current/?username=%s;password=%s;agent=ad2ccSM", user, pwd)
	data, err := getXML(url)
	if err != nil {
		return "", err
	}
	id, err := sessionId(data)
	if err != nil {
		return "", err
	}
	return id, nil
}

func sessionId(data []byte) (string, error) {
	v := Qtype{Callsign: []Ctype{}, Session: []Stype{}}
	err := xml.Unmarshal(data, &v)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(v.Session[0].Key) == "" {
		return "", fmt.Errorf("No key returned")
	}
	return v.Session[0].Key, nil
}

func getXML(url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return []byte{}, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("Status error: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("Read body: %v", err)
	}
	return data, nil
}

type sessionMgr func(string) (string, error)

func sessionCache() sessionMgr {
	var sessionKey string
	mgr := func(s string) (string, error) {
		if s == "" {
			if sessionKey == "" {
				return "", noKey
			}
			return sessionKey, nil
		}
		sessionKey = s
		return sessionKey, nil
	}
	return mgr
}
