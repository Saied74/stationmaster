package main

import (
	"database/sql"
	"errors"
	"time"
)

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
	"Time",
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
