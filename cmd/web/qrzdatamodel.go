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
}

type qrzModel struct {
	DB *sql.DB
}

func (m *qrzModel) insertQRZ(c *Ctype) error {

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
