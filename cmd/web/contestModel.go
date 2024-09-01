package main

import (
	"database/sql"
	"errors"
	"time"
)

type contestType interface {
	insertContest(*ContestRow) error
	getContest(string) (*ContestRow, error)
}

type ContestRow struct {
	Id          int
	Time        time.Time
	ContestName string
	FieldCount  int
	Field1Name  string
	Field2Name  string
	Field3Name  string
	Field4Name  string
	Field5Name  string
}

type contestModel struct {
	DB *sql.DB
}

func (m *contestModel) insertContest(l *ContestRow) error {

	stmt := `INSERT INTO contests (time, contestname, fieldCount,
		field1Name, field2Name, field3Name, field4Name, field5Name)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := m.DB.Exec(stmt, l.Time, l.ContestName, l.FieldCount,
		l.Field1Name, l.Field2Name, l.Field3Name, l.Field4Name, l.Field5Name)
	if err != nil {
		return err
	}

	_, err = result.LastInsertId()
	if err != nil {
		return err
	}
	return nil
}

func (m *contestModel) getContest(cn string) (*ContestRow, error) {
	stmt := `SELECT id, time, contestName, fieldCount,
	field1Name, field2Name, field3Name, field4Name, field5Name
	FROM contests WHERE contestname = ?`

	row := m.DB.QueryRow(stmt, cn)
	s := &ContestRow{}

	err := row.Scan(&s.Id, &s.Time, &s.ContestName, &s.FieldCount,
		&s.Field1Name, &s.Field2Name, &s.Field3Name, &s.Field4Name,
		&s.Field5Name)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errNoRecord
		} else {
			return nil, err
		}

	}
	return s, nil
}
