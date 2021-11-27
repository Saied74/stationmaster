package main

import (
	"errors"
)

var errTest = errors.New("error for use in testing")

type mockLogsModel struct {
	row         LogsRow
	rows        []LogsRow
	lastLogsErr error
	defaultErr  error
	band        string
	mode        string
}

func (f *mockLogsModel) getUniqueCountries() ([]LogsRow, error) {
	return []LogsRow{}, nil
}

func (f *mockLogsModel) getLogsByCountry(country string) ([]LogsRow, error) {
	return []LogsRow{}, nil
}

func (f *mockLogsModel) getLatestLogs(n int) ([]LogsRow, error) {
	if f.lastLogsErr == nil {
		return f.rows, nil
	}
	return []LogsRow{}, f.lastLogsErr
}

func (f *mockLogsModel) getContestLogs(n int) ([]LogsRow, error) {
	return []LogsRow{}, f.lastLogsErr
}

func (f *mockLogsModel) getCabrilloData(cd *contestData) ([]LogsRow, error) {
	return []LogsRow{}, nil
}

func (f *mockLogsModel) insertLog(l *LogsRow) (int, error) {
	return 0, nil
}

func (f *mockLogsModel) getLogByID(id int) (*LogsRow, error) {
	row := LogsRow{
		Id:   10,
		Call: "FOOBAR",
		Band: "40m",
		Mode: "FOO",
		Sent: "59",
		Rcvd: "37",
	}
	switch id {
	case 10:
		return &LogsRow{}, errTest
	case 1:
		return &row, nil
	}
	return &LogsRow{}, errTest
}

func (f *mockLogsModel) getLogsByCall(call string) ([]*LogsRow, error) {
	r := []*LogsRow{}
	for _, rr := range f.rows {
		r = append(r, &rr)
	}
	return r, nil
}

func (f *mockLogsModel) updateLog(l *LogsRow, id int) error {
	return nil
}

func (f *mockLogsModel) getADIFData() ([]LogsRow, error) {
	return []LogsRow{}, nil
}

func (f *mockLogsModel) updateQSO(map[itemType]string) error {
	return nil
}

func (f *mockLogsModel) getConfirmedCountries() ([]LogsRow, error) {
	return []LogsRow{}, nil
}

func (f *mockLogsModel) getLogsByCounty(county string) ([]LogsRow, error) {
	return []LogsRow{}, nil
}

func (f *mockLogsModel) getConfirmedCounties() ([]LogsRow, error) {
	return []LogsRow{}, nil
}

func (f *mockLogsModel) calcContestScore(*contestData) (int, error) {
	return 0, nil
}

func (f *mockLogsModel) getConfirmedContacts() ([]LogsRow, error) {
	return []LogsRow{}, nil
}

func (f *mockLogsModel) getConfirmedStates() ([]LogsRow, error) {
	return []LogsRow{}, nil
}

func (f *mockLogsModel) getLogsByState(state string) ([]LogsRow, error) {
	return []LogsRow, nil
}

func (f *mockLogsModel) findNeed(dx []DXClusters) ([]DXClusters, error) {
	return []DXClusters{}, nil
}

type mockQRZModel struct {
	row         LogsRow
	rows        []LogsRow
	lastLogsErr error
	defaultErr  error
	band        string
	mode        string
}

func (f *mockQRZModel) insertQRZ(c *Ctype) error {
	return nil
}

func (f *mockQRZModel) getQRZ(call string) (*Ctype, error) {
	return &Ctype{
		Fname:    "FOO",
		Lname:    "BAR",
		Born:     "1952",
		TimeZone: "EST",
		NickName: "FOOFOO",
		CQzone:   "EST",
		ITUzone:  "FOO BAR",
		Lat:      "1952",
		Long:     "EST",
	}, nil
}

func (f *mockQRZModel) updateQSOCount(call string, id int) error {
	return nil
}

func (f *mockQRZModel) stashQRZdata(*Ctype) error {
	return nil
}

func (f *mockQRZModel) unstashQRZdata() (*Ctype, error) {
	return &Ctype{}, nil
}

func (m *mockQRZModel) getUniqueCounties() ([]LogsRow, error) {
	return []LogsRow{}, nil
}

func (m *mockQRZModel) getRepeatContacts() ([]LogsRow, error) {
	return []LogsRow{}, nil
}

func (m *mockQRZModel) getUniqueStates() ([]LogsRow, error) {
	return []LogsRow{}, nil
}

type mockOtherModel struct {
	row         LogsRow
	rows        []LogsRow
	lastLogsErr error
	defaultErr  error
	band        string
	mode        string
}

func (f *mockOtherModel) getDefault(d string) (string, error) {
	if f.defaultErr != nil {
		return "", f.defaultErr
	}
	switch d {
	case "band":
		return f.band, nil
	case "mode":
		return f.mode, nil
	default:
		return "", nil
	}
}

func (f *mockOtherModel) updateDefault(k, v string) error {
	return nil
}

func (m *mockLogsModel) updateLOTWSent(id int) error {
	return nil
}
