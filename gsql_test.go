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
	"testing"
)

var serve = NewServer("127.0.0.1", 3306, "database", MySql).
	NewAuth("username", "password").
	NewConfig(100, 60)

type options struct {
	Id    int    `json:"id,string" sql:"primary key,auto_increment 1000"`
	Text  string `json:",string" sql:"varchar(20) default null"`
	Value string
}

func TestSql(t *testing.T) {
	option := &options{Id: 1, Text: "test"}
	orm := serve.NewStruct("table_options", option)
	if orm.Select().Where("Id=?").OrderBy("id desc").Execute().Error == nil {
		if orm.Result.RowsAffected > 0 {
			t.Log("row:", orm.Id, option.Id, orm.Result.DataTable.Rows[0]["Id"], orm.TC)
		} else {
			t.Log("row:", orm.Id, option.Id, "no data", orm.TC)
		}
	} else {
		t.Log(orm.Error)

	}
}
