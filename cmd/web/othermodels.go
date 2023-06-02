package main

import (
	"database/sql"
	"errors"
)

type otherType interface {
	getDefault(string) (string, error)
	updateDefault(string, string) error
}

type otherModel struct {
	DB *sql.DB
}

type DefType struct {
	Id  int
	Kee string
	Val string
}

func (m *otherModel) getDefault(k string) (string, error) {
	stmt := `SELECT id, kee, val FROM defaults WHERE kee = ?`

	row := m.DB.QueryRow(stmt, k)
	d := &DefType{}

	err := row.Scan(&d.Id, &d.Kee, &d.Val)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errNoRecord
		}
		return "", err
	}
	return d.Val, nil
}

func (m *otherModel) updateDefault(k, v string) error {
	_, err := m.getDefault(k)
	if errors.Is(err, errNoRecord) {
		stmt := `INSERT INTO defaults (kee, val) VALUES (?, ?)`
		_, err := m.DB.Exec(stmt, k, v)
		if err != nil {
			return err
		}
		return nil
	}
	stmt := `UPDATE defaults SET val = ? WHERE kee = ?`
	_, err = m.DB.Exec(stmt, v, k)
	if err != nil {
		return err
	}
	return nil
}
