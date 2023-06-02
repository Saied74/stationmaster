package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

//Qtype is the top level structure for the QRZ API
type Qtype struct {
	XMLName  xml.Name `xml:"QRZDatabase"`
	Callsign Ctype
	Session  Stype
}

//Ctype is the main payload of the QRZ API
type Ctype struct {
	XMLName     xml.Name `xml:"Callsign"`
	Id          int32
	Time        time.Time
	Call        string `xml:"call"`
	Aliases     string `xml:"aliases"`
	Dxcc        string `xml:"dxcc"`
	Fname       string `xml:"fname"`
	Lname       string `xml:"name"`
	Addr1       string `xml:"addr1"`
	Addr2       string `xml:"addr2"`
	State       string `xml:"state"`
	Zip         string `xml:"zip"`
	Country     string `xml:"country"` //Country name for QSL mailing address
	CountryCode string `xml:"ccode"`
	Lat         string `xml:"lat"`
	Long        string `xml:"lon"`
	Grid        string `xml:"grid"`
	County      string `xml:"county"`
	FIPS        string `xml:"fips"`
	Land        string `xml:"land"` //DXCC country name of the call sign
	EffDate     string `xml:"efdate"`
	ExpDate     string `xml:"expdate"`
	PrevCall    string `xml:"p_call"` //previous calls
	Class       string `xml:"class"`
	Codes       string `xml:"codes"` //FCC code for the license
	QSLMgr      string `xml:"qslmgr"`
	Email       string `xml:"email"`
	URL         string `xml:"url"`     //typically QRZ webpage.
	Views       string `xml:"u_views"` //Number of QRZ webpage views
	Bio         string `xml:"bio"`     //Number of bytes and updated date of QRZ Bio.
	Image       string `xml:"image"`   //url of the image on QRZ
	ModDate     string `xml:"moddate"`
	MSA         string `xml:"MSA"` //USPS Metropolitan Serving Area
	AreaCode    string `xml:"AreaCode"`
	TimeZone    string `xml:"TimeZone"`
	GMTOffset   string `xml:"GMTOffset"`
	DST         string `xml:"DST"` //DST observed (or not)
	EQSL        string `xml:"eqsl"`
	MQSL        string `xml:"mqsl"`
	CQzone      string `xml:"cqzone"`
	ITUzone     string `xml:"ituzone"`
	GeoLocation string `xml:"geoloc"`
	Attn        string `xml:"attn"`
	NickName    string `xml:"nickname"`
	WholeName   string `xml:"name_fmt"`
	Born        string `xml:"born"`
	QSOCount    int
}

//Stype is the validation part of the QRZ API
type Stype struct {
	Key   string `xml:"Key"`
	Count string `xml:"Count"`
	Time  string `xml:"GMTime"`
}

var noKey = errors.New("no session id")

func (app *application) getHamInfo(callSign string) (*Qtype, error) {
	key, err := app.sKey("")
	if err != nil {
		if errors.Is(err, noKey) {
			key, err = loginQRZ(app.qrzuser, app.qrzpw)
			if err != nil {
				return &Qtype{}, err
			}
			key, err = app.sKey(key)
			if err != nil {
				return &Qtype{}, err
			}
		} else {
			return &Qtype{}, err
		}
	}
	url := fmt.Sprintf("https://xmldata.qrz.com/xml/current/?s=%s;callsign=%s", key, callSign)
	result, err := getXML(url)
	if err != nil {
		return &Qtype{}, err
	}
	v := Qtype{Callsign: Ctype{}, Session: Stype{}}
	err = xml.Unmarshal(result, &v)
	if err != nil {
		return &Qtype{}, err
	}
	//QRZ.COM does not return the key if it has expired and it must be renewed.
	if v.Session.Key == "" {
		key, err = loginQRZ(app.qrzuser, app.qrzpw)
		if err != nil {
			return &Qtype{}, err
		}
		key, err = app.sKey(key)
		if err != nil {
			return &Qtype{}, err
		}
		url := fmt.Sprintf("https://xmldata.qrz.com/xml/current/?s=%s;callsign=%s", key, callSign)
		result, err := getXML(url)
		if err != nil {
			return &Qtype{}, err
		}
		v := Qtype{Callsign: Ctype{}, Session: Stype{}}
		err = xml.Unmarshal(result, &v)
		if err != nil {
			return &Qtype{}, err
		}
	}
	return &v, nil
}

func loginQRZ(user, pwd string) (string, error) {
	url := fmt.Sprintf("https://xmldata.qrz.com/xml/current/?username=%s;password=%s;agent=ad2ccSM", user, pwd)
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
	v := Qtype{Callsign: Ctype{}, Session: Stype{}}
	err := xml.Unmarshal(data, &v)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(v.Session.Key) == "" {
		return "", fmt.Errorf("No key returned")
	}
	return v.Session.Key, nil
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

func (m *otherModel) sKey(key string) (string, error) {
	sessionKey, err := m.getDefault("qrzkey")
	if err != nil {
		return "", err
	}
	if key == "" {
		if sessionKey == "" {
			return "", noKey
		}
		return sessionKey, nil
	}
	err = m.updateDefault("qrzkey", key)
	if err != nil {
		return "", noKey
	}
	return sessionKey, nil
}
