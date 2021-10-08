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
	"github.com/BlueStorm001/gsql/clickhouse"
	"github.com/BlueStorm001/gsql/datatable"
	"github.com/BlueStorm001/gsql/mssqls"
	"github.com/BlueStorm001/gsql/mysqls"
	"github.com/BlueStorm001/gsql/util"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

type DatabaseType string

const (
	MSSql      DatabaseType = "MSSql"
	MySql      DatabaseType = "MySql"
	Clickhouse DatabaseType = "Clickhouse"
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
	serve := &datatable.Serve{ConnectMax: runtime.NumCPU() * 2, Timeout: 60}
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
	case Clickhouse:
		s.ISQL = &clickhouse.Serve{Serve: s.Serve}
	default:
		s.Error = errors.New("unknown database type")
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

type ORM struct {
	*datatable.ORM
	Error        error
	ErrorSQL     string
	Id           int
	ST           time.Time     //execution start time
	TC           time.Duration //time consuming
	s            *Serve
	processLock  *util.Mutex
	chanState    bool
	chanComplete chan struct{}
}

type SqlResult struct {
	*datatable.SqlResult
}

func (s *Serve) NewStruct(table string, inStruct interface{}) *ORM {
	if util.Verify(table) {
		return &ORM{Error: errors.New("verification failed")}
	}
	if s.chs == nil {
		s.mu.Lock()
		if s.chs == nil {
			s.chs = make(chan *ORM, s.ConnectMax)
			for i := 0; i < s.ConnectMax; i++ {
				s.chs <- &ORM{ORM: &datatable.ORM{SqlCommand: util.NewBuilder()}, s: s, Id: i + 1, processLock: new(util.Mutex), chanComplete: make(chan struct{})}
			}
		}
		s.mu.Unlock()
	}
	orm := s.GetORM()
	if orm.Error == nil {
		orm.TableName = table
		orm.SqlStructMap = getStruct(inStruct)
	}
	return orm
}

func (s *Serve) GetORM() *ORM {
	if err := s.error(); err != nil {
		return &ORM{Error: err}
	}
	for i := 0; i < 10*s.Timeout; i++ {
		select {
		case c := <-s.chs:
			c.chanState = true
			go func(orm *ORM) {
				select {
				case <-orm.chanComplete:
					return
				case <-time.After(time.Second * time.Duration(s.Timeout)):
					orm.Dispose()
				}
			}(c)
			return c
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}
	return &ORM{Error: errors.New("maximum number of connections exceeded")}
}

func (o *ORM) SetStruct(inStruct interface{}) *ORM {
	if err := o.error(); err != nil {
		o.Error = err
		return o
	}
	o.SqlStructMap = getStruct(inStruct)
	return o
}

func (o *ORM) ColumnUse(columns ...string) *ORM {
	if len(o.SqlStructMap) > 0 && len(columns) > 0 {
		o.processLock.Lock()
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
		o.processLock.Unlock()
	}
	return o
}

func (o *ORM) ColumnExclude(columns ...string) *ORM {
	if len(o.SqlStructMap) > 0 && len(columns) > 0 {
		o.processLock.Lock()
		for _, column := range columns {
			if _, ok := o.SqlStructMap[column]; ok {
				delete(o.SqlStructMap, column)
			}
		}
		o.processLock.Unlock()
	}
	return o
}

func (o *ORM) SelectExclude(columns ...string) *ORM {
	if len(columns) > 0 {
		o.Columns = make(map[string]struct{})
		for _, column := range columns {
			o.Columns[column] = struct{}{}
		}
		o.ColumnMode = -1
	}
	return o.Select()
}

func (orm *ORM) Select(columns ...string) *ORM {
	o := orm.get()
	if err := o.error(); err != nil {
		o.Error = err
		return o
	}
	if len(columns) > 0 {
		o.Columns = make(map[string]struct{})
		for _, column := range columns {
			o.Columns[column] = struct{}{}
		}
		o.ColumnMode = 1
	}
	o.processLock.Lock()
	o.ST = time.Now()
	o.Mode = datatable.Get
	o.Error = o.s.ISQL.Select(o.ORM)
	return o
}

func (orm *ORM) Count() *ORM {
	o := orm.get()
	if err := o.error(); err != nil {
		o.Error = err
		return o
	}
	o.processLock.Lock()
	o.ST = time.Now()
	o.Mode = datatable.Count
	o.Error = o.s.ISQL.Count(o.ORM)
	return o
}

func (orm *ORM) Insert(columns ...string) *ORM {
	o := orm.get()
	if err := o.error(); err != nil {
		o.Error = err
		return o
	}
	if len(columns) > 0 {
		o.Columns = make(map[string]struct{})
		for _, column := range columns {
			o.Columns[column] = struct{}{}
		}
		o.ColumnMode = 1
	}
	o.processLock.Lock()
	o.ST = time.Now()
	o.Mode = datatable.Add
	o.Error = o.s.ISQL.Insert(o.ORM)
	return o
}

func (o *ORM) InsertExclude(columns ...string) *ORM {
	if len(columns) > 0 {
		o.Columns = make(map[string]struct{})
		for _, column := range columns {
			o.Columns[column] = struct{}{}
		}
		o.ColumnMode = -1
	}
	return o.Insert()
}

func (orm *ORM) Update(columns ...string) *ORM {
	o := orm.get()
	if err := o.error(); err != nil {
		o.Error = err
		return o
	}
	if len(columns) > 0 {
		o.Columns = make(map[string]struct{})
		for _, column := range columns {
			o.Columns[column] = struct{}{}
		}
		o.ColumnMode = 1
	}
	o.processLock.Lock()
	o.ST = time.Now()
	o.Mode = datatable.Set
	o.Error = o.s.ISQL.Update(o.ORM)
	return o
}

func (o *ORM) UpdateExclude(columns ...string) *ORM {
	if len(columns) > 0 {
		o.Columns = make(map[string]struct{})
		for _, column := range columns {
			o.Columns[column] = struct{}{}
		}
		o.ColumnMode = -1
	}
	return o.Update()
}

func (orm *ORM) Delete() *ORM {
	o := orm.get()
	if err := o.error(); err != nil {
		o.Error = err
		return o
	}
	o.processLock.Lock()
	o.ST = time.Now()
	o.Mode = datatable.Del
	o.Error = o.s.ISQL.Delete(o.ORM)
	return o
}

func (o *ORM) Where(wheres ...string) *ORM {
	if err := o.error(); err != nil {
		o.Error = err
		return o
	}
	o.Error = o.s.ISQL.Where(o.ORM, wheres...)
	return o
}

func (o *ORM) OrderBy(field string) *ORM {
	if err := o.error(); err != nil {
		o.Error = err
		return o
	}
	if field != "" {
		o.Error = o.s.ISQL.OrderBy(o.ORM, field)
	}
	return o
}

func (o *ORM) GroupBy(field string) *ORM {
	if err := o.error(); err != nil {
		o.Error = err
		return o
	}
	if field == "" {
		o.Error = errors.New("field is nil")
		return o
	}
	o.Error = o.s.ISQL.GroupBy(o.ORM, field)
	return o
}

func (o *ORM) Limit(limit int, offset ...int) *ORM {
	if err := o.error(); err != nil {
		o.Error = err
		return o
	}
	o.Error = o.s.ISQL.Limit(o.ORM, limit, offset...)
	return o
}

func (o *ORM) Page(size int, page int) *ORM {
	page = page - 1
	if page < 0 {
		page = 0
	}
	return o.Limit(size, page*size)
}

func (o *ORM) AddSql(command string) *ORM {
	if util.Verify(command) {
		o.Error = errors.New("verification failed")
		return o
	}
	command = strings.Replace(command, "\"", "'", -1)
	o.SqlCommand.Append(command)
	return o
}

func (o *ORM) Execute() *SqlResult {
	if o.chanState {
		o.chanComplete <- struct{}{}
	}
	defer o.s.reset(o)
	result := &SqlResult{SqlResult: new(datatable.SqlResult)}
	if err := o.error(); err != nil {
		result.Error = err
		if o.ORM != nil && o.SqlCommand.Len() > 0 {
			o.ErrorSQL = o.SqlCommand.ToString()
		}
		return result
	}
	switch o.Mode {
	case datatable.Get, datatable.Count:
		dt, err := o.s.ISQL.DataTable(o.ORM)
		if err == nil {
			if dt != nil {
				result.DataTable = dt
				result.RowsAffected = int64(dt.Count)
				result.DataTable.Name = o.TableName
				if o.Mode == datatable.Count && result.RowsAffected > 0 {
					result.RowsAffected = util.ToInt64(dt.Rows[0]["count"])
				}
			}
		} else {
			result.Error = err
		}
	case datatable.Add, datatable.Set, datatable.Del:
		res, err := o.s.ISQL.Execute(o.ORM)
		if err == nil {
			result.RowsAffected, _ = res.RowsAffected()
			result.LastInsertId, _ = res.LastInsertId()
		} else {
			result.Error = err
		}
	}
	if result.Error != nil {
		o.ErrorSQL = o.SqlCommand.ToString()
	}
	return result
}

func (o *ORM) Dispose() {
	if o.chanState {
		o.s.reset(o)
	}
}

func (o *ORM) Close() error {
	return o.s.Close()
}

func (s *Serve) Close() error {
	return s.ISQL.Close()
}

func (o *ORM) GetSQL() (string, map[string]*datatable.Field) {
	sqlStr := o.SqlCommand.ToString()
	maps := o.SqlStructMap
	o.s.reset(o)
	return sqlStr, maps
}

func (r *SqlResult) GetStruct(inStruct interface{}) error {
	if r.RowsAffected == 0 {
		return errors.New("data line is empty")
	}
	return setStruct(inStruct, r.DataTable.Rows)
}

func (o *ORM) get() *ORM {
	if o.chanState {
		return o
	}
	orm := o.s.GetORM()
	if orm.Error == nil {
		orm.SqlStructMap = o.SqlStructMap
		orm.TableName = o.TableName
	} else {
		orm.Error = o.Error
	}
	return orm
}

func (s *Serve) reset(orm *ORM) {
	if s == nil || orm == nil {
		return
	}
	orm.Mode = datatable.Not
	orm.chanState = false
	orm.Error = nil
	orm.SqlCommand.Reset()
	orm.SqlValues = nil
	orm.Columns = nil
	orm.ColumnMode = 0
	orm.TC = time.Since(orm.ST)
	if orm.processLock.State {
		orm.processLock.Unlock()
	}
	select {
	case s.chs <- orm:
		break
	default:
		break
	}
}

func (o *ORM) error() error {
	if o == nil {
		return errors.New(msg(505))
	}
	if o.Error != nil {
		return o.Error
	}
	if o.s == nil {
		return errors.New(msg(504))
	}
	if o.s.ISQL == nil {
		return errors.New(msg(502))
	}
	if o.s.chs == nil {
		return errors.New(msg(505))
	}
	if o.s.Error != nil {
		return errors.New("[500]" + o.s.Error.Error())
	}
	return nil
}

func (s *Serve) error() error {
	if s == nil {
		return errors.New(msg(504))
	}
	if s.ISQL == nil {
		return errors.New(msg(502))
	}
	if s.chs == nil {
		return errors.New(msg(505))
	}
	if s.Error != nil {
		return errors.New("[500]" + s.Error.Error())
	}
	return nil
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

func getStruct(in interface{}) map[string]*datatable.Field {
	if in == nil {
		return nil
	}
	maps := make(map[string]*datatable.Field)
	switch mp := in.(type) {
	case map[string]interface{}:
		for k, v := range mp {
			switch r := v.(type) {
			case []string:
				f := &datatable.Field{}
				for i, s := range r {
					switch i {
					case 0:
						f.Val = s
					case 1:
						f.Tag = s
					default:
						break
					}
				}
				maps[k] = f
			case []interface{}:
				if len(r) == 2 {
					maps[k] = &datatable.Field{Tag: util.ToString(r[1]), Val: r[0]}
				} else {
					maps[k] = &datatable.Field{Tag: "", Val: r[0]}
				}
			case *datatable.Field:
				maps[k] = r
			default:
				maps[k] = &datatable.Field{Tag: "", Val: r}
			}

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

func setStruct(in interface{}, rows []map[string]interface{}) error {
	refValue := reflect.ValueOf(in) // value
	refType := reflect.TypeOf(in)   // type
	rowsCount := len(rows)
	var fieldCount int
	kind := refValue.Kind()
	switch kind {
	case reflect.Ptr:
		fieldCount = refValue.Elem().NumField()
		refType = refType.Elem()
		refValue = refValue.Elem()
	case reflect.Struct:
		fieldCount = refValue.NumField()
	case reflect.Slice:
		for i := 0; i < refValue.Len(); i++ {
			if i >= rowsCount {
				return nil
			}
			e := refValue.Index(i)
			switch e.Kind() {
			case reflect.Ptr:
				refType = e.Type().Elem()
				fieldCount = refType.NumField()
				var value reflect.Value
				if e.IsNil() {
					value = reflect.New(refType)
				} else {
					value = e
				}
				for y := 0; y < fieldCount; y++ {
					key := refType.Field(y).Name
					if v, ok := rows[i][key]; ok {
						if reflect.ValueOf(v).Type() == value.Elem().Field(y).Type() {
							value.Elem().Field(y).Set(reflect.ValueOf(v))
						}
					}
				}
				e.Set(value)
			default:
				return errors.New("does not support this type")
			}
		}
		return nil
	}
	for i := 0; i < fieldCount; i++ {
		if i >= rowsCount {
			return nil
		}
		key := refType.Field(i).Name // field type
		if value, ok := rows[i][key]; ok {
			if reflect.ValueOf(value).Type() == refValue.Field(i).Type() {
				refValue.Field(i).Set(reflect.ValueOf(value))
			}
		}
	}
	return nil
}
