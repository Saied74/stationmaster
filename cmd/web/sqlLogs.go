package main

import (
	"database/sql"
	"errors"
	"fmt"
)

//will insert a new record into the stationlogs table
func (m *stationModel) insertLog(l *LogsRow) (int, error) {

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
func (m *stationModel) getLogByID(id int) (*LogsRow, error) {
	stmt := `SELECT time, callsign, mode, sent, rcvd,
	band, name, country, comment, lotwsent, lotwrcvd
	FROM stationlogs WHERE id = ?`

	row := m.DB.QueryRow(stmt, id)
	s := &LogsRow{}

	err := row.Scan(&s.Time, &s.Call, &s.Mode,
		&s.Sent, &s.Rcvd, &s.Band, &s.Name, &s.Country,
		&s.Comment, &s.Lotwrcvd, &s.Lotwsent)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errNoRecord
		} else {
			return nil, err
		}

	}
	return s, nil
}

//will return the n most recently created logs
func (m *stationModel) getLatestLogs(n int) ([]LogsRow, error) {
	stmt := fmt.Sprintf(`SELECT time, callsign, mode, sent, rcvd,
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

		err = rows.Scan(&s.Time, &s.Call, &s.Mode,
			&s.Sent, &s.Rcvd, &s.Band, &s.Name, &s.Country,
			&s.Comment, &s.Lotwrcvd, &s.Lotwsent)

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
