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
	getContestLogs(int) ([]LogsRow, error)
	getADIFData() ([]LogsRow, error)
	getCabrilloData(*contestData) ([]LogsRow, error)
	updateLOTWSent(int) error
	updateLog(*LogsRow, int) error
	calcContestScore(*contestData) (int, error)
	getUniqueCountries() ([]LogsRow, error)
	getConfirmedCountries() ([]LogsRow, error)
	getLogsByCountry(string) ([]LogsRow, error)
	getLogsByCounty(string) ([]LogsRow, error)
	getConfirmedCounties() ([]LogsRow, error)
	getConfirmedContacts() ([]LogsRow, error)
	updateQSO(map[itemType]string) error
	getConfirmedStates() ([]LogsRow, error)
	getLogsByState(string) ([]LogsRow, error)
	findNeed([]DXClusters) ([]DXClusters, error)
}

type logsModel struct {
	DB *sql.DB
}

var errNoRecord = errors.New("no matching record found")

//LogsRow is the data for the logs table rows
type LogsRow struct {
	Id          int
	Time        time.Time
	Call        string
	Mode        string
	Sent        string
	Rcvd        string
	Contest     string
	ContestName string
	ExchSent    string
	ExchRcvd    string
	Band        string
	Name        string
	County      string
	Cnty        bool
	State       string
	Country     string
	Comment     string
	Lotwsent    string
	Lotwrcvd    string
	LotwQSOdate time.Time
	LotwQSLdate time.Time
}

type headRow struct {
	Id          string
	Time        string
	Call        string
	Mode        string
	Sent        string
	Rcvd        string
	Contest     string
	ContestName string
	ExchSent    string
	ExchRcvd    string
	Band        string
	Name        string
	County      string
	Cnty        bool
	Country     string
	Comment     string
	Lotwsent    string
	Lotwrcvd    string
}

var tableHead = headRow{
	"ID",
	"Time (UTC)",
	"Call",
	"Mode",
	"Sent",
	"Rcvd",
	"",
	"",
	"Exch sent",
	"Exch rcvd",
	"Band",
	"Name",
	"County",
	false,
	"Country",
	"Comment",
	"LOTW Sent",
	"LOTW Rcvd",
}

//will insert a new record into the stationlogs table
func (m *logsModel) insertLog(l *LogsRow) (int, error) {
	if len(l.Name) > 100 {
		l.Name = l.Name[0:100]
	}
	if len(l.Country) > 100 {
		l.Country = l.Country[0:100]
	}
	if len(l.Comment) > 100 {
		l.Comment = l.Comment[0:100]
	}

	stmt := `INSERT INTO stationlogs (time, callsign, mode, sent, rcvd,
	band, name, country, comment, lotwsent, lotwrcvd, contest, exchsent,
exchrcvd, contestname)
	VALUES (UTC_TIMESTAMP(), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := m.DB.Exec(stmt,
		l.Call, l.Mode, l.Sent, l.Rcvd,
		l.Band, l.Name, l.Country, l.Comment, l.Lotwsent, l.Lotwrcvd,
		l.Contest, l.ExchSent, l.ExchRcvd, l.ContestName)
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
	band, name, country, comment, lotwsent, lotwrcvd, contest, exchsent,
	exchrcvd, contestname	FROM stationlogs ORDER BY time DESC LIMIT %d`, n)

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
			&s.Comment, &s.Lotwsent, &s.Lotwrcvd, &s.Contest,
			&s.ExchSent, &s.ExchRcvd, &s.ContestName)

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

func (m *logsModel) getConfirmedContacts() ([]LogsRow, error) {
	stmt := `SELECT id, time, callsign, mode, sent, rcvd,
	band, name, country, comment, lotwsent, lotwrcvd, contest, exchsent,
	exchrcvd, contestname	FROM stationlogs WHERE lotwrcvd = ? AND lotwsent = ?
	ORDER BY time DESC`

	rows, err := m.DB.Query(stmt, "Yes", "Yes")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tr := []*LogsRow{}

	for rows.Next() {
		s := &LogsRow{}

		err = rows.Scan(&s.Id, &s.Time, &s.Call, &s.Mode,
			&s.Sent, &s.Rcvd, &s.Band, &s.Name, &s.Country,
			&s.Comment, &s.Lotwsent, &s.Lotwrcvd, &s.Contest,
			&s.ExchSent, &s.ExchRcvd, &s.ContestName)

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

//will return the n most recently created logs
func (m *logsModel) getContestLogs(n int) ([]LogsRow, error) {
	stmt := fmt.Sprintf(`SELECT id, time, callsign, mode, sent, rcvd,
	band, name, country, comment, lotwsent, lotwrcvd, contest, exchsent,
	exchrcvd, contestname	FROM stationlogs WHERE contest='Yes' ORDER BY time DESC LIMIT %d`, n)

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
			&s.Comment, &s.Lotwsent, &s.Lotwrcvd, &s.Contest,
			&s.ExchSent, &s.ExchRcvd, &s.ContestName)

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

func (m *logsModel) calcContestScore(cd *contestData) (int, error) {
	stmt := `SELECT exchrcvd FROM stationlogs WHERE contest = ? AND
	contestname = ? AND time >= ? AND time <= ?`

	rows, err := m.DB.Query(stmt, "Yes", cd.name, cd.startTime, cd.endTime)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	tr1 := []LogsRow{}
	for rows.Next() {
		s := LogsRow{}
		err := rows.Scan(&s.ExchRcvd)

		if err != nil {
			return 0, err
		}
		tr1 = append(tr1, s)

	}
	if err = rows.Err(); err != nil {
		return 0, err
	}

	stmt = `SELECT DISTINCT exchrcvd FROM stationlogs WHERE contest = ? AND
	contestname = ? AND time >= ? AND time <= ?`

	rows2, err := m.DB.Query(stmt, "Yes", cd.name, cd.startTime, cd.endTime)
	if err != nil {
		return 0, err
	}
	defer rows2.Close()

	tr2 := []LogsRow{}
	for rows2.Next() {
		s := LogsRow{}
		err := rows2.Scan(&s.ExchRcvd)

		if err != nil {
			return 0, err
		}
		tr2 = append(tr2, s)

	}
	if err = rows2.Err(); err != nil {
		return 0, err
	}

	return 2 * len(tr1) * len(tr2), nil
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
	FROM stationlogs WHERE lotwsent <> ? ORDER BY time DESC`

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

func (m *logsModel) getCabrilloData(cd *contestData) ([]LogsRow, error) {
	stmt := `SELECT id, time, callsign, mode, sent, rcvd,
	band, name, country, comment, lotwsent, lotwrcvd, contest, exchsent,
	exchrcvd, contestname FROM stationlogs
	WHERE contest = ? AND contestname = ? AND time >= ? AND time <= ?
	ORDER BY time DESC`

	rows, err := m.DB.Query(stmt, "Yes", cd.name, cd.startTime, cd.endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tr := []LogsRow{}
	for rows.Next() {
		s := LogsRow{}
		err := rows.Scan(&s.Id, &s.Time, &s.Call, &s.Mode,
			&s.Sent, &s.Rcvd, &s.Band, &s.Name, &s.Country,
			&s.Comment, &s.Lotwsent, &s.Lotwrcvd, &s.Contest,
			&s.ExchSent, &s.ExchRcvd, &s.ContestName)

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

func (m *logsModel) getUniqueCountries() ([]LogsRow, error) {

	stmt := `SELECT DISTINCT country FROM stationlogs ORDER BY country ASC`
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tr := []*LogsRow{}

	for rows.Next() {
		s := &LogsRow{}

		err = rows.Scan(&s.Country)

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

func (m *logsModel) getConfirmedCountries() ([]LogsRow, error) {

	stmt := `SELECT DISTINCT country FROM stationlogs where lotwrcvd = ? ORDER BY country ASC`
	rows, err := m.DB.Query(stmt, "YES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tr := []*LogsRow{}

	for rows.Next() {
		s := &LogsRow{}

		err = rows.Scan(&s.Country)

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

func (m *logsModel) getLogsByCountry(country string) ([]LogsRow, error) {
	stmt := `SELECT id, time, callsign, mode, sent, rcvd,
	band, name, country, comment, lotwsent, lotwrcvd
	FROM stationlogs WHERE country = ? ORDER BY time DESC`

	rows, err := m.DB.Query(stmt, country)
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

func (m *logsModel) getLogsByCounty(county string) ([]LogsRow, error) {
	stmt := `SELECT stationlogs.id, stationlogs.time, stationlogs.callsign,
	stationlogs.mode, stationlogs.sent, stationlogs.rcvd,	stationlogs.band,
	stationlogs.name, stationlogs.comment, stationlogs.lotwsent,
	stationlogs.lotwrcvd, qrztable.county, qrztable.state
	FROM stationlogs inner join qrztable on
	stationlogs.callsign=qrztable.callsign WHERE qrztable.county = ? and
	stationlogs.country = ? ORDER BY time DESC`

	rows, err := m.DB.Query(stmt, county, "United States")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tr := []*LogsRow{}

	for rows.Next() {
		s := &LogsRow{}

		err = rows.Scan(&s.Id, &s.Time, &s.Call, &s.Mode, &s.Sent, &s.Rcvd,
			&s.Band, &s.Name, &s.Comment, &s.Lotwsent, &s.Lotwrcvd, &s.County,
			&s.State)

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
		item.Cnty = true
		t = append(t, *item)
	}

	return t, nil
}

func (m *logsModel) getLogsByState(state string) ([]LogsRow, error) {
	stmt := `SELECT stationlogs.id, stationlogs.time, stationlogs.callsign,
	stationlogs.mode, stationlogs.sent, stationlogs.rcvd,	stationlogs.band,
	stationlogs.name, stationlogs.comment, stationlogs.lotwsent,
	stationlogs.lotwrcvd, qrztable.county, qrztable.state
	FROM stationlogs inner join qrztable on
	stationlogs.callsign=qrztable.callsign WHERE qrztable.state = ? and
	stationlogs.country = ? ORDER BY time DESC`

	rows, err := m.DB.Query(stmt, state, "United States")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tr := []*LogsRow{}

	for rows.Next() {
		s := &LogsRow{}

		err = rows.Scan(&s.Id, &s.Time, &s.Call, &s.Mode, &s.Sent, &s.Rcvd,
			&s.Band, &s.Name, &s.Comment, &s.Lotwsent, &s.Lotwrcvd, &s.County,
			&s.State)

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
		item.Cnty = true
		t = append(t, *item)
	}

	return t, nil
}

func (m *logsModel) getConfirmedCounties() ([]LogsRow, error) {
	stmt := `SELECT stationlogs.id, stationlogs.time, stationlogs.callsign,
	stationlogs.mode, stationlogs.sent, stationlogs.rcvd,	stationlogs.band,
	stationlogs.name, stationlogs.comment, stationlogs.lotwsent,
	stationlogs.lotwrcvd, qrztable.county, qrztable.state
	FROM stationlogs inner join qrztable on
	stationlogs.callsign=qrztable.callsign WHERE stationlogs.lotwrcvd = ? and
	stationlogs.country = ? and qrztable.county <> '' ORDER BY time DESC`

	rows, err := m.DB.Query(stmt, "YES", "United States")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tr := []*LogsRow{}

	for rows.Next() {
		s := &LogsRow{}

		err = rows.Scan(&s.Id, &s.Time, &s.Call, &s.Mode, &s.Sent, &s.Rcvd,
			&s.Band, &s.Name, &s.Comment, &s.Lotwsent, &s.Lotwrcvd, &s.County,
			&s.State)

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
		item.Cnty = true
		t = append(t, *item)
	}

	return t, nil
}

func (m *logsModel) getConfirmedStates() ([]LogsRow, error) {
	stmt := `SELECT DISTINCT qrztable.state
	FROM stationlogs inner join qrztable on
	stationlogs.callsign=qrztable.callsign WHERE stationlogs.lotwrcvd = ? and
	stationlogs.country = ? and qrztable.state <> '' ORDER BY state ASC`

	rows, err := m.DB.Query(stmt, "YES", "United States")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tr := []*LogsRow{}

	for rows.Next() {
		s := &LogsRow{}

		err = rows.Scan(&s.State)

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
		item.Cnty = true
		t = append(t, *item)
	}

	return t, nil
}

func (m *logsModel) updateQSO(row map[itemType]string) error {

	stmt := `SELECT id, time, callsign, mode, band FROM stationlogs
	WHERE callSign = ?` // AND time = ?`
	qsoTime := row[itemQSOTimeStamp]
	t, err := time.Parse(time.RFC3339, qsoTime)
	if err != nil {
		return err
	}
	rows, err := m.DB.Query(stmt, row[itemCall]) //, t)
	if err != nil {
		return err
	}
	defer rows.Close()
	var s = &LogsRow{}
	for rows.Next() {
		s = &LogsRow{}
		err := rows.Scan(&s.Id, &s.Time, &s.Call, &s.Mode, &s.Band)
		if err != nil {
			return err
		}

		if int(t.Sub(s.Time).Round(time.Minute)) == 0 {
			break
		}
	}
	if err = rows.Err(); err != nil {
		return err
	}

	tQSO, err := timeIt(row[itemRxQSO])
	if err != nil {
		return err
	}
	tQSL, err := timeIt(row[itemRxQSL])
	if err != nil {
		return err
	}
	stmt = `UPDATE stationlogs SET lotwrcvd = ?, lotwqsodate=?, lotwqsldate= ?
		WHERE id = ?`
	_, err = m.DB.Exec(stmt, "YES", tQSO, tQSL, s.Id)
	if err != nil {
		return err
	}
	return nil
}

func (m *logsModel) findNeed(dx []DXClusters) ([]DXClusters, error) {
	newDX := []DXClusters{}
	stmt := `SELECT DISTINCT country FROM stationlogs where lotwrcvd = ? and country = ?`
	for _, d := range dx {
		row := m.DB.QueryRow(stmt, "YES", d.Country)
		s := &LogsRow{}
		err := row.Scan(&s.Call)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				d.Need = "Yes"
				newDX = append(newDX, d)
				continue
			} else {
				return nil, err
			}
		}
		d.Need = "No"
		newDX = append(newDX, d)
	}

	return newDX, nil
}
