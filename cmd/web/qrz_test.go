package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"
)

const (
	xmlTestData = `<?xml version="1.0" ?>
<QRZDatabase version="1.34">
  <Callsign>
      <call>AA7BQ</call>
      <aliases>N6UFT,KJ6RK,DL/AA7BQ</aliases>
      <dxcc>291</dxcc>
      <fname>FRED L</fname>
      <name>LLOYD</name>
      <addr1>8711 E PINNACLE PEAK RD 193</addr1>
      <addr2>SCOTTSDALE</addr2>
      <state>AZ</state>
      <zip>85255</zip>
      <country>United States</country>
      <ccode>291</ccode>
      <lat>34.23456</lat>
      <lon>-112.34356</lon>
      <grid>DM32af</grid>
      <county>Maricopa</county>
      <fips>04013</fips>
      <land>USA</land>
      <efdate>2000-01-20</efdate>
      <expdate>2010-01-20</expdate>
      <p_call>KJ6RK</p_call>
      <class>E</class>
      <codes>HAI</codes>
      <qslmgr>NONE</qslmgr>
      <email>flloyd@qrz.com</email>
      <url>https://www.qrz.com/db/aa7bq</url>
      <u_views>115336</u_views>
      <bio>3937/2003-11-04</bio>
      <image>https://files.qrz.com/q/aa7bq/aa7bq.jpg</image>
      <serial>3626</serial>
      <moddate>2003-11-04 19:37:02</moddate>
      <MSA>6200</MSA>
      <AreaCode>602</AreaCode>
      <TimeZone>Mountain</TimeZone>
      <GMTOffset>-7</GMTOffset>
      <DST>N</DST>
      <eqsl>Y</eqsl>
      <mqsl>Y</mqsl>
      <cqzone>3</cqzone>
      <ituzone>2</ituzone>
      <geoloc>user</geoloc>
      <attn>c/o QRZ LLC</attn>
      <nickname>The Boss</nickname>
      <name_fmt>FRED "The Boss" LLOYD</name_fmt>
      <born>1953</born>
  </Callsign>
  <Session>
      <Key>2331uf894c4bd29f3923f3bacf02c532d7bd9</Key>
      <Count>123</Count>
      <SubExp>Wed Jan 1 12:34:03 2013</SubExp>
      <GMTime>Sun Nov 16 04:13:46 2012</GMTime>
  </Session>
</QRZDatabase>`

	idTestData = `<QRZDatabase xmlns="http://xmldata.qrz.com" version="1.24">
<Session>
<Key>0b1df943c413fd6d7a60a7ca8af868fd</Key>
<Count>2</Count>
<SubExp>non-subscriber</SubExp>
<GMTime>Wed Jun 30 03:27:41 2021</GMTime>
<Remark>cpu: 0.160s</Remark>
</Session>
</QRZDatabase>`

	sId = `0b1df943c413fd6d7a60a7ca8af868fd`
)

func TestSessionId(t *testing.T) {
	id, err := sessionId([]byte(idTestData))
	if err != nil {
		t.Errorf("error is getting session id %v", err)
	}
	if id != sId {
		t.Errorf("in getting session id, wanted %s got %s", sId, id)
	}
}

type mockGetType func(url string) (*http.Response, error)

type mockClient struct {
	mockGet mockGetType
}

func (m *mockClient) Get(url string) (*http.Response, error) {
	return m.mockGet(url)
}

func TestGetXML(t *testing.T) {

	r := ioutil.NopCloser(bytes.NewReader([]byte(xmlTestData)))
	client = &mockClient{
		mockGet: func(url string) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       r,
			}, nil
		},
	}

	result, err := getXML("abc")
	if err != nil {
		t.Errorf("testing getXML call failed with %v ", err)
	}
	if len(result) == 0 {
		t.Errorf("empty result was returned")
	}
	v := Qtype{Callsign: []Ctype{}, Session: []Stype{}}
	err = xml.Unmarshal(result, &v)
	if err != nil {
		t.Errorf("returned xml did not decode %v", err)
	}
	if v.Callsign[0].Call != "AA7BQ" && v.Callsign[0].Fname != "LLOYD" {
		t.Errorf("wanted call sign AA7BQ got %s wanted firs name LLOYD got %s",
			v.Callsign[0].Call, v.Callsign[0].Fname)
	}
	client = &mockClient{
		mockGet: func(url string) (*http.Response, error) {
			return &http.Response{
				StatusCode: 400,
				Body:       r,
			}, nil
		},
	}
	result, err = getXML("abc")
	if err == nil {
		t.Errorf("testing getXML statusCode 400 failed with %v ", err)
	}

}

func TestLoginQRZ(t *testing.T) {
	r := ioutil.NopCloser(bytes.NewReader([]byte(idTestData)))
	client = &mockClient{
		mockGet: func(url string) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       r,
			}, nil
		},
	}
	id, err := loginQRZ("abc", "def")
	if err != nil {
		t.Errorf("loginQRZ returned error %v", err)
	}
	if id != sId {
		t.Errorf("expecting id %s got %s", sId, id)
	}
}

func TestGetHamInfo(t *testing.T) {
	r := ioutil.NopCloser(bytes.NewReader([]byte(xmlTestData)))
	client = &mockClient{
		mockGet: func(url string) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       r,
			}, nil
		},
	}

	app := &application{
		qrzuser: "abc",
		qrzpw:   "xyz",
		sKey:    sessionCache(),
	}
	app.sKey(sId)
	v, err := app.getHamInfo("abc")
	if err != nil {
		t.Errorf("getHamInfo returned error %v", err)
	}
	if v == nil {
		t.Errorf("getHamInfo did not return anything")
		return
	}
	if len(v.Callsign) < 1 {
		t.Errorf("Callsign field of returned hamInfo is empty")
		return
	}
	if v.Callsign[0].Call != "AA7BQ" && v.Callsign[0].Fname != "LLOYD" {
		t.Errorf("wanted call sign AA7BQ got %s wanted firs name LLOYD got %s",
			v.Callsign[0].Call, v.Callsign[0].Fname)
	}
}

func TestSessionCache(t *testing.T) {
	testKey := "some key"
	altKey := "another key"
	sm := sessionCache()

	key, err := sm("")
	if !errors.Is(err, noKey) {
		t.Errorf("empty session cache did not return noKey")
	}
	if key != "" {
		t.Errorf("empty session cache returned %s, not empty", key)
	}
	key, err = sm(testKey)
	if err != nil {
		t.Errorf("key insertion into cache failed %v", err)
	}
	if key != testKey {
		t.Errorf("key insertion did not return the inserted key %s", key)
	}
	if key == "" {
		t.Errorf("key insertion returned no key")
	}

	key, err = sm("")
	if errors.Is(err, noKey) {
		t.Errorf("key request returned noKey error")
	}
	if err != nil {
		t.Errorf("key request returned error %v", err)
	}
	if key == "" {
		t.Errorf("key request returned empty key instead of %s", testKey)
	}
	if key != testKey {
		t.Errorf("key request returned %s instead of %s", key, testKey)
	}

	key, err = sm(altKey)
	if errors.Is(err, noKey) {
		t.Errorf("key update returned noKey error")
	}
	if err != nil {
		t.Errorf("key update returned error %v", err)
	}
	if key == "" {
		t.Errorf("key update returned empty key instead of %s", altKey)
	}
	if key != altKey {
		t.Errorf("key request returned %s instead of %s", key, altKey)
	}
}
