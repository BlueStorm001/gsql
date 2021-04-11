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

package gsql

import (
	"database/sql"
	"errors"
	"gsql/datatable"
	"gsql/mssqls"
	"gsql/mysqls"
	"gsql/util"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

type DatabaseType string

const (
	MSSql DatabaseType = "MSSql"
	MySql DatabaseType = "MySql"
)

type Serve struct {
	*datatable.Serve
	mu  sync.Mutex
	chs chan *ORM
}

func NewServer(host string, port int) *Serve {
	serve := &datatable.Serve{Host: host, Port: port, ConnectMax: runtime.NumCPU() * 2, Timeout: 60}
	server := &Serve{Serve: serve}
	return server
}

func NewDrive(baseType DatabaseType, drive func() (db *sql.DB, err error)) *Serve {
	serve := &datatable.Serve{ConnectMax: runtime.NumCPU() * 2, Timeout: 60, Auth: new(datatable.Auth)}
	s := &Serve{Serve: serve}
	s.Database(baseType, "")
	s.Drive = drive
	s.DriveMode = 2
	return s
}

func (s *Serve) Database(baseType DatabaseType, database string) *Serve {
	s.Serve.Database = database
	switch baseType {
	case MySql:
		s.ISQL = &mysqls.Serve{Serve: s.Serve}
	case MSSql:
		s.ISQL = &mssqls.Serve{Serve: s.Serve}
	}
	return s
}

func (s *Serve) Config(connectMax, timeout int) *Serve {
	if connectMax > 0 {
		s.ConnectMax = connectMax
	}
	if timeout > 0 {
		s.Timeout = timeout
	}
	return s
}

//
func (s *Serve) Login(user, pass string) *Serve {
	if s.Auth == nil {
		s.Auth = &datatable.Auth{User: user, Pass: pass}
	} else {
		s.Auth.User = user
		s.Auth.Pass = pass
	}
	return s
}

func (s *Serve) NewDrive(drive func(s *datatable.Serve) (db *sql.DB, err error)) *Serve {
	s.DriveServe = drive
	s.DriveMode = 1
	return s
}

type useMode int

const (
	not useMode = iota
	get
	add
	set
	del
)

type ORM struct {
	*datatable.ORM
	Error error
	Id    int
	ST    time.Time     //execution start time
	TC    time.Duration //time consuming
	mode  useMode
	s     *Serve
	mu    sync.Mutex
}

func (s *Serve) NewStruct(table string, inStruct interface{}) *ORM {
	if s.chs == nil {
		s.mu.Lock()
		if s.chs == nil {
			s.chs = make(chan *ORM, s.ConnectMax)
			for i := 0; i < s.ConnectMax; i++ {
				s.chs <- &ORM{ORM: &datatable.ORM{SqlCommand: util.NewBuilder()}, s: s, Id: i + 1}
			}
		}
		s.mu.Unlock()
	}
	orm := s.GetORM()
	if orm.Error == nil {
		orm.TableName = table
		orm.SqlStructMap = reflects(inStruct)
	}
	return orm
}

func (s *Serve) GetORM() *ORM {
	if s.err() {
		return &ORM{Error: s.Error}
	}
	for i := 0; i < 10*s.Timeout; i++ {
		select {
		case c := <-s.chs:
			return c
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}
	return &ORM{Error: errors.New("time out! max connections of reached")}
}

func (o *ORM) SetStruct(inStruct interface{}) *ORM {
	if o.s.err() {
		return o
	}
	o.SqlStructMap = reflects(inStruct)
	return o
}

func (o *ORM) ColumnUse(columns ...string) *ORM {
	if len(o.SqlStructMap) > 0 && len(columns) > 0 {
		o.mu.Lock()
		for k := range o.SqlStructMap {
			var use bool
			for _, column := range columns {
				if k == column {
					use = true
					break
				}
			}
			if !use {
				delete(o.SqlStructMap, k)
			}
		}
		o.mu.Unlock()
	}
	return o
}

func (o *ORM) ColumnExclude(columns ...string) *ORM {
	if len(o.SqlStructMap) > 0 && len(columns) > 0 {
		o.mu.Lock()
		for _, column := range columns {
			if _, ok := o.SqlStructMap[column]; ok {
				delete(o.SqlStructMap, column)
			}
		}
		o.mu.Unlock()
	}
	return o
}

func (o *ORM) Select(columns ...string) *ORM {
	if o.s.err() {
		return o
	}
	o.mu.Lock()
	o.ST = time.Now()
	o.mode = get
	o.Error = o.s.ISQL.Select(o.ORM, columns...)
	return o
}

func (o *ORM) Insert(columns ...string) *ORM {
	if o.s.err() {
		return o
	}
	if len(columns) > 0 {
		o.ColumnUse(columns...)
	}
	o.mu.Lock()
	o.ST = time.Now()
	o.mode = add
	o.Error = o.s.ISQL.Insert(o.ORM)
	return o
}

func (o *ORM) InsertExclude(columns ...string) *ORM {
	if len(columns) > 0 {
		o.ColumnExclude(columns...)
	}
	return o.Insert()
}

func (o *ORM) Update(columns ...string) *ORM {
	if o.s.err() {
		return o
	}
	if len(columns) > 0 {
		o.ColumnUse(columns...)
	}
	o.mu.Lock()
	o.ST = time.Now()
	o.mode = set
	o.Error = o.s.ISQL.Update(o.ORM)
	return o
}

func (o *ORM) UpdateExclude(columns ...string) *ORM {
	if len(columns) > 0 {
		o.ColumnExclude(columns...)
	}
	return o.Update()
}

func (o *ORM) Delete() *ORM {
	if o.s.err() {
		return o
	}
	o.mu.Lock()
	o.ST = time.Now()
	o.mode = del
	o.Error = o.s.ISQL.Delete(o.ORM)
	return o
}

func (o *ORM) Where(wheres ...string) *ORM {
	if o.s.err() {
		return o
	}
	o.Error = o.s.ISQL.Where(o.ORM, wheres...)
	return o
}

func (o *ORM) OrderBy(field string) *ORM {
	if o.s.err() {
		return o
	}
	o.Error = o.s.ISQL.OrderBy(o.ORM, field)
	return o
}

func (o *ORM) GroupBy(field string) *ORM {
	if o.s.err() {
		return o
	}
	o.Error = o.s.ISQL.GroupBy(o.ORM, field)
	return o
}

func (o *ORM) Limit(limit int, offset ...int) *ORM {
	if o.s.err() {
		return o
	}
	o.Error = o.s.ISQL.Limit(o.ORM, limit, offset...)
	return o
}

func (o *ORM) Pagination(size int, page int) *ORM {
	limit := size
	if page <= 1 {
		page = 0
	} else {
		limit = size * page
		page = limit - size
	}
	return o.Limit(limit, page)
}

func (o *ORM) Execute() *ORM {
	defer o.s.reset(o)
	if o.s.err() {
		return nil
	}
	switch o.mode {
	case get:
		dt, err := o.s.ISQL.DataTable(o.ORM)
		if err == nil {
			o.Result = &datatable.SqlResult{DataTable: dt}
			if dt != nil {
				o.Result.RowsAffected = int64(dt.Count)
				o.Result.DataTable.Name = o.TableName
			}
		} else {
			o.Error = err
		}
	case add, set, del:
		res, err := o.s.ISQL.Execute(o.ORM)
		if err == nil {
			rowsAffected, _ := res.RowsAffected()
			lastInsertId, _ := res.LastInsertId()
			o.Result = &datatable.SqlResult{LastInsertId: lastInsertId, RowsAffected: rowsAffected}
		} else {
			o.Error = err
		}
	}
	return o
}

func (o *ORM) GetStruct(inStruct interface{}) error {
	if o.s.err() {
		return o.s.Error
	}
	if o.Error != nil {
		return o.Error
	}
	if o.Result.RowsAffected == 0 {
		return errors.New("datatable rows count 0")
	}

	return nil
}
func (s *Serve) reset(orm *ORM) {
	orm.mode = not
	orm.SqlCommand.Reset()
	orm.SqlValues = nil
	orm.TC = time.Since(orm.ST)
	orm.mu.Unlock()
	select {
	case s.chs <- orm:
		break
	default:
		break
	}
}

func (s *Serve) err() bool {
	if s == nil {
		s.Error = errors.New(msg(504))
		return true
	}
	if s.Auth == nil {
		s.Error = errors.New(msg(501))
		return true
	}
	if s.ISQL == nil {
		s.Error = errors.New(msg(502))
		return true
	}
	if s.chs == nil {
		s.Error = errors.New(msg(505))
		return true
	}
	if s.Error != nil {
		return true
	}
	return false
}

func msg(code int) string {
	switch code {
	case 504:
		return "[504]Serve must be created first!"
	case 502:
		return "[502]ISQL is null"
	case 505:
		return "[505]ORM must be created first!"
	case 501:
		return "[501]Auth must be created first!"
	default:
		return ""
	}
}

func reflects(in interface{}) map[string]*datatable.Field {
	if in == nil {
		return nil
	}
	maps := make(map[string]*datatable.Field)
	switch mp := in.(type) {
	case map[string]interface{}:
		for k, v := range mp {
			maps[k] = &datatable.Field{Tag: "", Val: v}
		}
		return maps
	}
	refValue := reflect.ValueOf(in) // value
	refType := reflect.TypeOf(in)   // type
	var fieldCount int              // field count
	//指针类型
	if refValue.Kind() == reflect.Ptr {
		fieldCount = refValue.Elem().NumField()
		refType = refType.Elem()
		refValue = refValue.Elem()
	} else {
		fieldCount = refValue.NumField()
	}
	for i := 0; i < fieldCount; i++ {
		key := refType.Field(i).Name // field type
		tag := refType.Field(i).Tag.Get("sql")
		val := refValue.Field(i).Interface()
		maps[key] = &datatable.Field{Tag: strings.ToLower(tag), Val: val}
	}
	return maps
}
