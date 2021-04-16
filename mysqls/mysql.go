// Copyright (c) 2021 BlueStorm
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFINGEMENT IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package mysqls

import (
	"database/sql"
	"errors"
	"gsql/datatable"
	"gsql/util"
	"strings"
)

type Serve struct {
	*datatable.Serve
	conn *sql.DB
	//drive func(s *Serve) (db *sql.DB, err error)
}

func (s *Serve) connect() error {
	switch s.DriveMode {
	case 1:
		s.conn, s.Error = s.DriveServe(s.Serve)
	case 2:
		s.conn, s.Error = s.Drive()
	default:
		s.Error = errors.New("drive mode error")
	}
	if s.Error != nil {
		return s.Error
	}
	return s.conn.Ping()
}

func (s *Serve) close() error {
	var err error
	if s.conn != nil {
		if err = s.conn.Close(); err == nil {
			s.conn = nil
		}
	}
	return err
}

func (s *Serve) query(command string, args ...interface{}) (*sql.Rows, error) {
	if s.conn == nil {
		if err := s.connect(); err != nil {
			return nil, err
		}
	}
	return s.conn.Query(command, args...)
}

func (s *Serve) exec(command string, args ...interface{}) (sql.Result, error) {
	if s.conn == nil {
		if err := s.connect(); err != nil {
			return nil, err
		}
	}
	return s.conn.Exec(command, args...)
}

func (s *Serve) dataTable(command string, params ...interface{}) (*datatable.DataTable, error) {
	rows, err := s.query(command, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sr := datatable.SqlRows{Rows: rows}
	return sr.GetDataTable()
}

func (s *Serve) dataSet(command string, params ...interface{}) (*datatable.DataSet, error) {
	rows, err := s.query(command, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sr := datatable.SqlRows{Rows: rows}
	return sr.GetDataSet()
}

func (s *Serve) Select(orm *datatable.ORM, columns ...string) error {
	orm.SqlCommand.Reset()
	orm.SqlCommand.Append("SELECT ")
	var use bool
	if len(columns) > 0 {
		for _, column := range columns {
			if util.Verify(column) {
				return errors.New("verification failed")
			}
			if use {
				orm.SqlCommand.Append(",")
			}
			orm.SqlCommand.Append(column)
			use = true
		}
	} else {
		for k := range orm.SqlStructMap {
			if use {
				orm.SqlCommand.Append(",")
			}
			orm.SqlCommand.Append(k)
			use = true
		}
	}
	orm.SqlCommand.Append(" FROM ").Append(orm.TableName)
	return nil
}

func (s *Serve) Count(orm *datatable.ORM) error {
	orm.SqlCommand.Reset()
	orm.SqlCommand.Append(" SELECT count(1) as count FROM ").Append(orm.TableName)
	return nil
}

func (s *Serve) Insert(orm *datatable.ORM) error {
	orm.SqlCommand.Reset()
	orm.SqlCommand.Append(" INSERT INTO ").Append(orm.TableName)
	fieldStr := "("
	valueStr := "("
	for k, v := range orm.SqlStructMap {
		if v.Tag != "" && strings.Contains(v.Tag, "auto_increment") {
			continue
		}
		if len(fieldStr) > 1 {
			fieldStr += ","
			valueStr += ","
		}
		fieldStr += k
		valueStr += "?"
		orm.SqlValues = append(orm.SqlValues, v.Val)
	}
	fieldStr += ")"
	valueStr += ")"
	orm.SqlCommand.Append(fieldStr).Append("VALUES").Append(valueStr)
	return nil
}

func (s *Serve) Update(orm *datatable.ORM) error {
	orm.SqlCommand.Reset()
	orm.SqlCommand.Append(" UPDATE ").Append(orm.TableName).Append(" SET ")
	var use bool
	for k, v := range orm.SqlStructMap {
		if v.Tag != "" && strings.Contains(v.Tag, "auto_increment") {
			continue
		}
		if use {
			orm.SqlCommand.Append(",")
		}
		orm.SqlCommand.Append(k).Append("=?")
		orm.SqlValues = append(orm.SqlValues, v.Val)
		use = true
	}
	return nil
}

func (s *Serve) Delete(orm *datatable.ORM) error {
	orm.SqlCommand.Reset()
	orm.SqlCommand.Append(" DELETE FROM ").Append(orm.TableName).Append(" ")
	return nil
}

func (s *Serve) Where(orm *datatable.ORM, wheres ...string) error {
	if len(wheres) == 0 {
		return nil
	}
	orm.SqlCommand.Append(" WHERE")
	for _, w := range wheres {
		if util.Verify(w) {
			return errors.New("verification failed")
		}
		f := util.GetFieldName(w)
		if v, ok := orm.SqlStructMap[f]; ok {
			orm.SqlValues = append(orm.SqlValues, v.Val)
		} else {
			return errors.New("the query condition does not exist")
		}
		orm.SqlCommand.Append(" ").Append(w)
	}
	return nil
}

func (s *Serve) OrderBy(orm *datatable.ORM, field string) error {
	orm.SqlCommand.Append(" ORDER BY ").Append(field)
	return nil
}

func (s *Serve) GroupBy(orm *datatable.ORM, field string) error {
	orm.SqlCommand.Append(" GROUP BY ").Append(field)
	return nil
}

func (s *Serve) Limit(orm *datatable.ORM, limit int, offset ...int) error {
	if len(offset) > 0 {
		orm.SqlCommand.AppendInt(offset[0]).Append(",").AppendInt(limit)
	} else {
		orm.SqlCommand.AppendInt(limit)
	}
	return nil
}

func (s *Serve) DataSet(orm *datatable.ORM) (*datatable.DataSet, error) {
	return s.dataSet(orm.SqlCommand.String(), orm.SqlValues...)
}

func (s *Serve) DataTable(orm *datatable.ORM) (*datatable.DataTable, error) {
	return s.dataTable(orm.SqlCommand.String(), orm.SqlValues...)
}

func (s *Serve) Execute(orm *datatable.ORM) (sql.Result, error) {
	return s.exec(orm.SqlCommand.String(), orm.SqlValues...)
}
