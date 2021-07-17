package main

import "errors"

var errTest = errors.New("error for use in testing")

type mockLogsModel struct {
	row         LogsRow
	rows        []LogsRow
	lastLogsErr error
	defaultErr  error
	band        string
	mode        string
}

var testToggle int

func (f *mockLogsModel) getLatestLogs(n int) ([]LogsRow, error) {
	row := LogsRow{
		Id:   10,
		Call: "AD2CC",
		Band: "40m",
		Mode: "FOO",
		Sent: "59",
		Rcvd: "37",
	}
	switch testToggle {
	case 1:
		row.Band = "20m"
	case 2:
		row.Band = "40m"
	default:
		row.Band = "80m"
	}
	if f.lastLogsErr == nil {
		return []LogsRow{row, row}, nil
	}
	return []LogsRow{}, f.lastLogsErr
}

func (f *mockLogsModel) insertLog(l *LogsRow) (int, error) {
	return 0, nil
}

func (f *mockLogsModel) getLogByID(id int) (*LogsRow, error) {
	return &LogsRow{}, nil
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
	return &Ctype{}, nil
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
