package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type logsType interface {
	insertLog(*LogsRow) (int, error)
	getLogByID(int) (*LogsRow, error)
	getLogsByCall(string) ([]*LogsRow, error)
	getLatestLogs(int) ([]LogsRow, error)
	getADIFData() ([]LogsRow, error)
	updateLOTWSent(int) error
	updateLog(*LogsRow, int) error
}

type logsModel struct {
	DB *sql.DB
}

var errNoRecord = errors.New("no matching record found")

//LogsRow is the data for the logs table rows
type LogsRow struct {
	Id       int
	Time     time.Time
	Call     string
	Mode     string
	Sent     string
	Rcvd     string
	Band     string
	Name     string
	Country  string
	Comment  string
	Lotwsent string
	Lotwrcvd string
}

type headRow struct {
	Id       string
	Time     string
	Call     string
	Mode     string
	Sent     string
	Rcvd     string
	Band     string
	Name     string
	Country  string
	Comment  string
	Lotwsent string
	Lotwrcvd string
}

var tableHead = headRow{
	"ID",
	"Time (UTC)",
	"Call",
	"Mode",
	"Sent",
	"Rcvd",
	"Band",
	"Name",
	"Country",
	"Comment",
	"LOTW Sent",
	"LOTW Rcvd",
}

//will insert a new record into the stationlogs table
func (m *logsModel) insertLog(l *LogsRow) (int, error) {

	stmt := `INSERT INTO stationlogs (time, callsign, mode, sent, rcvd,
	band, name, country, comment, lotwsent, lotwrcvd)
	VALUES (UTC_TIMESTAMP(), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := m.DB.Exec(stmt,
		l.Call, l.Mode, l.Sent, l.Rcvd,
		l.Band, l.Name, l.Country, l.Comment, l.Lotwsent, l.Lotwrcvd)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

//will get a record given its id
func (m *logsModel) getLogByID(id int) (*LogsRow, error) {
	stmt := `SELECT id, time, callsign, mode, sent, rcvd,
	band, name, country, comment, lotwsent, lotwrcvd
	FROM stationlogs WHERE id = ?`

	row := m.DB.QueryRow(stmt, id)
	s := &LogsRow{}

	err := row.Scan(&s.Id, &s.Time, &s.Call, &s.Mode,
		&s.Sent, &s.Rcvd, &s.Band, &s.Name, &s.Country,
		&s.Comment, &s.Lotwsent, &s.Lotwrcvd)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errNoRecord
		} else {
			return nil, err
		}

	}
	return s, nil
}

func (m *logsModel) getLogsByCall(call string) ([]*LogsRow, error) {
	stmt := `SELECT id, time, callsign, mode, sent, rcvd,
	band, name, country, comment, lotwsent, lotwrcvd
	FROM stationlogs WHERE callsign = ?`

	rows, err := m.DB.Query(stmt, call)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tr := []*LogsRow{}
	for rows.Next() {
		s := &LogsRow{}
		err := rows.Scan(&s.Id, &s.Time, &s.Call, &s.Mode,
			&s.Sent, &s.Rcvd, &s.Band, &s.Name, &s.Country,
			&s.Comment, &s.Lotwsent, &s.Lotwrcvd)

		if err != nil {
			return nil, err
		}
		tr = append(tr, s)

	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tr, nil
}

//will return the n most recently created logs
func (m *logsModel) getLatestLogs(n int) ([]LogsRow, error) {
	stmt := fmt.Sprintf(`SELECT id, time, callsign, mode, sent, rcvd,
	band, name, country, comment, lotwsent, lotwrcvd
	FROM stationlogs ORDER BY time DESC LIMIT %d`, n)

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tr := []*LogsRow{}

	for rows.Next() {
		s := &LogsRow{}

		err = rows.Scan(&s.Id, &s.Time, &s.Call, &s.Mode,
			&s.Sent, &s.Rcvd, &s.Band, &s.Name, &s.Country,
			&s.Comment, &s.Lotwsent, &s.Lotwrcvd)

		if err != nil {
			return nil, err
		}
		tr = append(tr, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	t := []LogsRow{}
	for _, item := range tr {
		t = append(t, *item)
	}

	return t, nil
}

func (m *logsModel) updateLog(l *LogsRow, id int) error {
	stmt := `UPDATE stationlogs SET callsign = ?, mode = ?, sent = ?,
rcvd = ?, band = ?, name = ?, country = ?, comment = ?, lotwsent = ?,
lotwrcvd = ?  WHERE id = ?`
	_, err := m.DB.Exec(stmt,
		l.Call, l.Mode, l.Sent, l.Rcvd,
		l.Band, l.Name, l.Country, l.Comment, l.Lotwsent, l.Lotwrcvd, id)
	if err != nil {
		return err
	}
	return nil
}

func (m *logsModel) getADIFData() ([]LogsRow, error) {
	stmt := `SELECT id, time, callsign, mode, sent, rcvd,
	band, name, country, comment, lotwsent, lotwrcvd
	FROM stationlogs WHERE lotwsent <> ?`

	rows, err := m.DB.Query(stmt, "YES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tr := []*LogsRow{}
	for rows.Next() {
		s := &LogsRow{}
		err := rows.Scan(&s.Id, &s.Time, &s.Call, &s.Mode,
			&s.Sent, &s.Rcvd, &s.Band, &s.Name, &s.Country,
			&s.Comment, &s.Lotwsent, &s.Lotwrcvd)

		if err != nil {
			return nil, err
		}
		tr = append(tr, s)

	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	t := []LogsRow{}
	for _, item := range tr {
		t = append(t, *item)
	}
	return t, nil
}

func (m *logsModel) updateLOTWSent(id int) error {
	stmt := `UPDATE stationlogs SET lotwsent = ? WHERE id = ?`
	_, err := m.DB.Exec(stmt, "YES", id)
	if err != nil {
		return err
	}
	return nil
}
