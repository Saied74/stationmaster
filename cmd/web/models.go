package main

import (
	"errors"
	//"time"
	"database/sql"
)

type stationModel struct {
	DB *sql.DB
}

var errNoRecord = errors.New("no matching record found")

//TableRow type is the table header and content for the log table
type LogsRow struct {
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
