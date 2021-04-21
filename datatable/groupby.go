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

package datatable

import "github.com/BlueStorm001/gsql/util"

func (dt *DataTable) GroupBy(query string) *DataTable {
	if dt.Count <= 1 {
		return dt
	}
	exp := GroupBy([]byte(query))
	count := len(exp.GroupExpr)
	if count == 0 {
		return dt
	}
	dataTable := &DataTable{Name: dt.Name}
	for _, item := range exp.GroupExpr {
		for _, column := range dt.Columns {
			if column.Name == item.Name {
				dataTable.Columns = append(dataTable.Columns, column)
				break
			}
		}
	}
	if len(dataTable.Columns) == 0 {
		return dt
	}
	var group = make(map[string]int)
	for _, dr := range dt.Rows {
		row := make(map[string]interface{})
		var key string
		for _, column := range dataTable.Columns {
			value := dr[column.Name]
			row[column.Name] = value
			key += "$" + util.ToString(value) + "$"
		}
		if _, ok := group[key]; ok {
			group[key]++
		} else {
			group[key] = 1
			row["$GroupKey$"] = key
			dataTable.Rows = append(dataTable.Rows, row)
		}
	}

	for _, dr := range dataTable.Rows {
		dr["$GroupCount$"] = group[util.ToString(dr["$GroupKey$"])]
		dataTable.Count++
	}
	return dataTable
}
