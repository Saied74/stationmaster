package main

import "errors"

type fakeModel struct {
	row         LogsRow
	rows        []LogsRow
	lastLogsErr error
	defaultErr  error
	band        string
	mode        string
}

var errTest = errors.New("error for use in testing")

func (f *fakeModel) getLatestLogs(n int) ([]LogsRow, error) {
	if f.lastLogsErr == nil {
		return f.rows, nil
	}
	return []LogsRow{}, f.lastLogsErr
}

func (f *fakeModel) getDefault(d string) (string, error) {
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

func (f *fakeModel) updateDefault(k, v string) error {
	return nil
}

func (f *fakeModel) insertQRZ(c *Ctype) error {
	return nil
}

func (f *fakeModel) getQRZ(call string) (*Ctype, error) {
	return &Ctype{}, nil
}

func (f *fakeModel) updateQSOCount(call string, id int) error {
	return nil
}

func (f *fakeModel) stashQRZdata(*Ctype) error {
	return nil
}

func (f *fakeModel) unstashQRZdata() (*Ctype, error) {
	return &Ctype{}, nil
}
func (f *fakeModel) insertLog(l *LogsRow) (int, error) {
	return 0, nil
}

func (f *fakeModel) getLogByID(id int) (*LogsRow, error) {
	return &LogsRow{}, nil
}

func (f *fakeModel) getLogsByCall(call string) ([]*LogsRow, error) {
	r := []*LogsRow{}
	for _, rr := range f.rows {
		r = append(r, &rr)
	}
	return r, nil
}

func (f *fakeModel) updateLog(l *LogsRow, id int) error {
	return nil
}
