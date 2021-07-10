package main

import (
	"database/sql"
	"errors"
	"fmt"
)

func (m *stationModel) insertQRZ(c *Ctype) error {

	stmt := `INSERT INTO qrztable (time, callsign, aliases, dxcc, first_name,
		last_name, nickname, born, addr1, addr2, state, zip, country, country_code,
		lat, lon, grid, county, fips, land, cqzone, ituzone, geolocation, effdate,
		expdate, prevcall, class, codes, qslmgr, email, url, views, bio, image,
		moddate, msa, areacode, timezone, gmtoffset, dst, eqsl, mqsl, attn, qso_count)
	VALUES (UTC_TIMESTAMP(), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
	?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := m.DB.Exec(stmt, c.Call, c.Aliases, c.Dxcc, c.Fname, c.Lname,
		c.NickName, c.Born, c.Addr1, c.Addr2, c.State, c.Zip, c.Country, c.CountryCode,
		c.Lat, c.Long, c.Grid, c.County, c.FIPS, c.Land, c.CQzone, c.ITUzone,
		c.GeoLocation, c.EffDate, c.ExpDate, c.PrevCall, c.Class, c.Codes, c.QSLMgr,
		c.Email, c.URL, c.Views, c.Bio, c.Image, c.ModDate, c.MSA, c.AreaCode, c.TimeZone,
		c.GMTOffset, c.DST, c.EQSL, c.MQSL, c.Attn, c.QSOCount)
	if err != nil {
		return err
	}
	return nil
}

func (m *stationModel) getQRZ(call string) (*Ctype, error) {

	stmt := `SELECT id, time, callsign, aliases, dxcc, first_name,
		last_name, nickname, born, addr1, addr2, state, zip, country, country_code,
		lat, lon, grid, county, fips, land, cqzone, ituzone, geolocation, effdate,
		expdate, prevcall, class, codes, qslmgr, email, url, views, bio, image,
		moddate, msa, areacode, timezone, gmtoffset, dst, eqsl, mqsl, attn,
		qso_count FROM qrztable WHERE callsign = ?`

	row := m.DB.QueryRow(stmt, call)
	c := &Ctype{}

	err := row.Scan(&c.Id, &c.Time, &c.Call, &c.Aliases, &c.Dxcc, &c.Fname, &c.Lname,
		&c.NickName, &c.Born, &c.Addr1, &c.Addr2, &c.State, &c.Zip, &c.Country, &c.CountryCode,
		&c.Lat, &c.Long, &c.Grid, &c.County, &c.FIPS, &c.Land, &c.CQzone, &c.ITUzone,
		&c.GeoLocation, &c.EffDate, &c.ExpDate, &c.PrevCall, &c.Class, &c.Codes, &c.QSLMgr,
		&c.Email, &c.URL, &c.Views, &c.Bio, &c.Image, &c.ModDate, &c.MSA, &c.AreaCode, &c.TimeZone,
		&c.GMTOffset, &c.DST, &c.EQSL, &c.MQSL, &c.Attn, &c.QSOCount)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errNoRecord
		} else {
			return nil, err
		}

	}
	return c, nil
}

func (m *stationModel) stashQRZdata(c *Ctype) error {
	stmt := `UPDATE stashtable SET callsign = ?, aliases = ?, dxcc = ?, first_name = ?,
		last_name = ?, nickname = ?, born = ?, addr1 = ?, addr2 = ?, state = ?, zip = ?, country = ?, country_code = ?,
		lat = ?, lon = ?, grid = ?, county = ?, fips = ?, land = ?, cqzone = ?, ituzone = ?, geolocation = ?, effdate = ?,
		expdate = ?, prevcall = ?, class = ?, codes = ?, qslmgr = ?, email = ?, url = ?, views = ?, bio = ?, image = ?,
		moddate = ?, msa = ?, areacode = ?, timezone = ?, gmtoffset = ?, dst = ?, eqsl = ?, mqsl = ?, attn = ?,
		qso_count = ? WHERE id = ?`

	_, err := m.DB.Exec(stmt, c.Call, c.Aliases, c.Dxcc, c.Fname, c.Lname,
		c.NickName, c.Born, c.Addr1, c.Addr2, c.State, c.Zip, c.Country, c.CountryCode,
		c.Lat, c.Long, c.Grid, c.County, c.FIPS, c.Land, c.CQzone, c.ITUzone,
		c.GeoLocation, c.EffDate, c.ExpDate, c.PrevCall, c.Class, c.Codes, c.QSLMgr,
		c.Email, c.URL, c.Views, c.Bio, c.Image, c.ModDate, c.MSA, c.AreaCode, c.TimeZone,
		c.GMTOffset, c.DST, c.EQSL, c.MQSL, c.Attn, c.QSOCount, 1)

	if err != nil {
		return err
	}
	return nil
}

func (m *stationModel) unstashQRZdata() (*Ctype, error) {

	stmt := `SELECT id, time, callsign, aliases, dxcc, first_name,
		last_name, nickname, born, addr1, addr2, state, zip, country, country_code,
		lat, lon, grid, county, fips, land, cqzone, ituzone, geolocation, effdate,
		expdate, prevcall, class, codes, qslmgr, email, url, views, bio, image,
		moddate, msa, areacode, timezone, gmtoffset, dst, eqsl, mqsl, attn,
		qso_count FROM stashtable WHERE id = ?`

	row := m.DB.QueryRow(stmt, 1)
	c := &Ctype{}

	err := row.Scan(&c.Id, &c.Time, &c.Call, &c.Aliases, &c.Dxcc, &c.Fname, &c.Lname,
		&c.NickName, &c.Born, &c.Addr1, &c.Addr2, &c.State, &c.Zip, &c.Country, &c.CountryCode,
		&c.Lat, &c.Long, &c.Grid, &c.County, &c.FIPS, &c.Land, &c.CQzone, &c.ITUzone,
		&c.GeoLocation, &c.EffDate, &c.ExpDate, &c.PrevCall, &c.Class, &c.Codes, &c.QSLMgr,
		&c.Email, &c.URL, &c.Views, &c.Bio, &c.Image, &c.ModDate, &c.MSA, &c.AreaCode, &c.TimeZone,
		&c.GMTOffset, &c.DST, &c.EQSL, &c.MQSL, &c.Attn, &c.QSOCount)

	if err != nil {
		return nil, err
	}
	return c, nil
}

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
	stmt := `SELECT id, time, callsign, mode, sent, rcvd,
	band, name, country, comment, lotwsent, lotwrcvd
	FROM stationlogs WHERE id = ?`

	row := m.DB.QueryRow(stmt, id)
	s := &LogsRow{}

	err := row.Scan(&s.Id, &s.Time, &s.Call, &s.Mode,
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

func (m *stationModel) getLogsByCall(call string) ([]*LogsRow, error) {
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
			&s.Comment, &s.Lotwrcvd, &s.Lotwsent)

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
func (m *stationModel) getLatestLogs(n int) ([]LogsRow, error) {
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

func (m *stationModel) updateLog(l *LogsRow, id int) error {
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
