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
		orm.SqlStructMap = reflects(inStruct)
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
	if err := o.s.error(); err != nil {
		o.Error = err
		return o
	}
	o.SqlStructMap = reflects(inStruct)
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

func (orm *ORM) Select(columns ...string) *ORM {
	o := orm.get()
	if err := o.s.error(); err != nil {
		o.Error = err
		return o
	}
	o.processLock.Lock()
	o.ST = time.Now()
	o.Mode = datatable.Get
	o.Error = o.s.ISQL.Select(o.ORM, columns...)
	return o
}

func (orm *ORM) Count() *ORM {
	o := orm.get()
	if err := o.s.error(); err != nil {
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
	if err := o.s.error(); err != nil {
		o.Error = err
		return o
	}
	if len(columns) > 0 {
		o.ColumnUse(columns...)
	}
	o.processLock.Lock()
	o.ST = time.Now()
	o.Mode = datatable.Add
	o.Error = o.s.ISQL.Insert(o.ORM)
	return o
}

func (o *ORM) InsertExclude(columns ...string) *ORM {
	if len(columns) > 0 {
		o.ColumnExclude(columns...)
	}
	return o.Insert()
}

func (orm *ORM) Update(columns ...string) *ORM {
	o := orm.get()
	if err := o.s.error(); err != nil {
		o.Error = err
		return o
	}
	if len(columns) > 0 {
		o.ColumnUse(columns...)
	}
	o.processLock.Lock()
	o.ST = time.Now()
	o.Mode = datatable.Set
	o.Error = o.s.ISQL.Update(o.ORM)
	return o
}

func (o *ORM) UpdateExclude(columns ...string) *ORM {
	if len(columns) > 0 {
		o.ColumnExclude(columns...)
	}
	return o.Update()
}

func (orm *ORM) Delete() *ORM {
	o := orm.get()
	if err := o.s.error(); err != nil {
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
	if err := o.s.error(); err != nil {
		o.Error = err
		return o
	}
	o.Error = o.s.ISQL.Where(o.ORM, wheres...)
	return o
}

func (o *ORM) OrderBy(field string) *ORM {
	if err := o.s.error(); err != nil {
		o.Error = err
		return o
	}
	if field != "" {
		o.Error = o.s.ISQL.OrderBy(o.ORM, field)
	}
	return o
}

func (o *ORM) GroupBy(field string) *ORM {
	if err := o.s.error(); err != nil {
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
	if err := o.s.error(); err != nil {
		o.Error = err
		return o
	}
	o.Error = o.s.ISQL.Limit(o.ORM, limit, offset...)
	return o
}

func (o *ORM) Page(size int, page int) *ORM {
	limit := size
	if page <= 1 {
		page = 0
	} else {
		limit = size * page
		page = limit - size
	}
	return o.Limit(limit, page)
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

func (o *ORM) Execute() *datatable.SqlResult {
	if o.chanState {
		o.chanComplete <- struct{}{}
	}
	defer o.s.reset(o)
	result := &datatable.SqlResult{}
	if err := o.s.error(); err != nil {
		result.Error = err
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

func (o *ORM) GetSQL() (string, map[string]*datatable.Field) {
	o.s.reset(o)
	return o.SqlCommand.ToString(), o.SqlStructMap
}

func (o *ORM) GetStruct(inStruct interface{}) error {
	if err := o.s.error(); err != nil {
		return err
	}
	if o.Error != nil {
		return o.Error
	}
	//if o.Result.RowsAffected == 0 {
	//	return errors.New("datatable rows count 0")
	//}

	return nil
}

func (o *ORM) get() *ORM {
	if o.chanState {
		return o
	}
	orm := o.s.GetORM()
	orm.SqlStructMap = o.SqlStructMap
	orm.TableName = o.TableName
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

func reflects(in interface{}) map[string]*datatable.Field {
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
