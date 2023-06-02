package main

import (
	"database/sql"
	"errors"
)

type qrzType interface {
	insertQRZ(*Ctype) error
	getQRZ(string) (*Ctype, error)
	updateQSOCount(string, int) error
	stashQRZdata(*Ctype) error
	unstashQRZdata() (*Ctype, error)
	getUniqueCounties() ([]LogsRow, error)
	getRepeatContacts() ([]LogsRow, error)
	getUniqueStates() ([]LogsRow, error)
}

type qrzModel struct {
	DB *sql.DB
}

func (m *qrzModel) insertQRZ(c *Ctype) error {
	if c.Call == "" {
		return nil
	}
	if len(c.Call) > 20 {
		c.Call = c.Call[0:20]
	}
	if len(c.Aliases) > 50 {
		c.Aliases = c.Aliases[0:50]
	}
	if len(c.Dxcc) > 5 {
		c.Dxcc = c.Dxcc[0:5]
	}
	if len(c.Fname) > 100 {
		c.Fname = c.Fname[0:100]
	}
	if len(c.Lname) > 100 {
		c.Lname = c.Lname[0:100]
	}
	if len(c.NickName) > 50 {
		c.NickName = c.NickName[0:50]
	}
	if len(c.Born) > 5 {
		c.Born = c.Born[0:5]
	}
	if len(c.Addr1) > 50 {
		c.Addr1 = c.Addr1[0:50]
	}
	if len(c.Addr2) > 50 {
		c.Addr2 = c.Addr2[0:50]
	}
	if len(c.State) > 20 {
		c.State = c.State[0:20]
	}
	if len(c.Zip) > 10 {
		c.Zip = c.Zip[0:10]
	}
	if len(c.Country) > 50 {
		c.Country = c.Country[0:50]
	}
	if len(c.CountryCode) > 5 {
		c.CountryCode = c.CountryCode[0:5]
	}
	if len(c.Lat) > 15 {
		c.Lat = c.Lat[0:15]
	}
	if len(c.Long) > 15 {
		c.Long = c.Long[0:15]
	}
	if len(c.Grid) > 10 {
		c.Grid = c.Grid[0:10]
	}
	if len(c.County) > 50 {
		c.County = c.County[0:50]
	}
	if len(c.FIPS) > 10 {
		c.FIPS = c.FIPS[0:10]
	}
	if len(c.Land) > 50 {
		c.Land = c.Land[0:50]
	}
	if len(c.CQzone) > 5 {
		c.CQzone = c.CQzone[0:5]
	}
	if len(c.ITUzone) > 5 {
		c.ITUzone = c.ITUzone[0:5]
	}
	if len(c.GeoLocation) > 10 {
		c.GeoLocation = c.GeoLocation[0:10]
	}
	if len(c.EffDate) > 10 {
		c.EffDate = c.EffDate[0:10]
	}
	if len(c.ExpDate) > 10 {
		c.ExpDate = c.ExpDate[0:10]
	}
	if len(c.PrevCall) > 10 {
		c.PrevCall = c.PrevCall[0:10]
	}
	if len(c.Class) > 5 {
		c.Class = c.Class[0:5]
	}
	if len(c.Codes) > 5 {
		c.Codes = c.Codes[0:5]
	}
	if len(c.QSLMgr) > 100 {
		c.QSLMgr = c.QSLMgr[0:100]
	}
	if len(c.Email) > 50 {
		c.Email = c.Email[0:50]
	}
	if len(c.URL) > 50 {
		c.URL = c.URL[0:50]
	}
	if len(c.Views) > 20 {
		c.Views = c.Views[0:20]
	}
	if len(c.Bio) > 50 {
		c.Bio = c.Bio[0:50]
	}
	if len(c.Image) > 150 {
		c.Image = c.Image[0:150]
	}
	if len(c.ModDate) > 30 {
		c.ModDate = c.ModDate[0:30]
	}
	if len(c.MSA) > 5 {
		c.MSA = c.MSA[0:5]
	}
	if len(c.AreaCode) > 5 {
		c.AreaCode = c.AreaCode[0:5]
	}
	if len(c.TimeZone) > 20 {
		c.TimeZone = c.TimeZone[0:20]
	}
	if len(c.GMTOffset) > 5 {
		c.GMTOffset = c.GMTOffset[0:5]
	}
	if len(c.DST) > 3 {
		c.DST = c.DST[0:3]
	}
	if len(c.EQSL) > 3 {
		c.EQSL = c.EQSL[0:3]
	}
	if len(c.MQSL) > 3 {
		c.MQSL = c.MQSL[0:3]
	}
	if len(c.Attn) > 100 {
		c.Attn = c.Attn[0:100]
	}
	
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

func (m *qrzModel) getQRZ(call string) (*Ctype, error) {

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
		}
		return nil, err
	}
	return c, nil
}

func (m *qrzModel) updateQSOCount(callSign string, qsoCount int) error {
	stmt := `UPDATE qrztable SET qso_count = ? WHERE callsign = ?`
	_, err := m.DB.Exec(stmt, qsoCount, callSign)
	if err != nil {
		return err
	}
	return nil
}

func (m *qrzModel) stashQRZdata(c *Ctype) error {
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

func (m *qrzModel) unstashQRZdata() (*Ctype, error) {

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

func (m *qrzModel) getUniqueCounties() ([]LogsRow, error) {
	stmt := `select county, state from qrztable where state in
	(select distinct state from qrztable where country=?)
	and country=? and county <> ''`
	rows, err := m.DB.Query(stmt, "United States", "United States")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tr := []*LogsRow{}

	for rows.Next() {
		s := &LogsRow{}

		err = rows.Scan(&s.County, &s.State)

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

func (m *qrzModel) getRepeatContacts() ([]LogsRow, error) {

	stmt := `SELECT stationlogs.id, stationlogs.time, stationlogs.callsign,
	stationlogs.mode, stationlogs.sent, stationlogs.rcvd,	stationlogs.band,
	stationlogs.name, stationlogs.comment, stationlogs.lotwsent,
	stationlogs.lotwrcvd, qrztable.country, qrztable.state
	FROM stationlogs inner join qrztable on
	stationlogs.callsign=qrztable.callsign WHERE qrztable.qso_count > 1
	ORDER BY time DESC`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return []LogsRow{}, err
	}
	defer rows.Close()
	tr := []*LogsRow{}

	for rows.Next() {
		s := &LogsRow{}

		err = rows.Scan(&s.Id, &s.Time, &s.Call, &s.Mode, &s.Sent, &s.Rcvd,
			&s.Band, &s.Name, &s.Comment, &s.Lotwsent, &s.Lotwrcvd, &s.Country,
			&s.State)

		if err != nil {
			return []LogsRow{}, err
		}
		tr = append(tr, s)
	}
	if err = rows.Err(); err != nil {
		return []LogsRow{}, err
	}
	t := []LogsRow{}
	for _, item := range tr {
		t = append(t, *item)
	}

	return t, nil
}

func (m *qrzModel) getUniqueStates() ([]LogsRow, error) {
	stmt := `select distinct state, country from qrztable where country = ?
	and state <> '' order by state asc`
	rows, err := m.DB.Query(stmt, "United States")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tr := []*LogsRow{}

	for rows.Next() {
		s := &LogsRow{}

		err = rows.Scan(&s.State, &s.Country)

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
