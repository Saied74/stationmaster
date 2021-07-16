package main

import (
	"database/sql"
	"errors"
	"time"
)

type dbModel interface {
	getDefault(string) (string, error)
	updateDefault(string, string) error
	insertQRZ(*Ctype) error
	getQRZ(string) (*Ctype, error)
	updateQSOCount(string, int) error
	stashQRZdata(*Ctype) error
	unstashQRZdata() (*Ctype, error)
	insertLog(*LogsRow) (int, error)
	getLogByID(int) (*LogsRow, error)
	getLogsByCall(string) ([]*LogsRow, error)
	getLatestLogs(int) ([]LogsRow, error)
	updateLog(*LogsRow, int) error
}

type stationModel struct {
	DB *sql.DB
}

var errNoRecord = errors.New("no matching record found")

//TableRow type is the table header and content for the log table
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
