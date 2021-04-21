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

import (
	"github.com/BlueStorm001/gsql/util"
	"sort"
)

func (dt *DataTable) OrderBy(query string) *DataTable {
	if dt.Count <= 1 {
		return dt
	}
	exp := OrderBy([]byte(query))
	count := len(exp.OrderExpr)
	if count == 0 {
		return dt
	}
	var less = make([]lessFunc, count)
	for i, item := range exp.OrderExpr {
		less[i] = func(c1, c2 map[string]interface{}) bool {
			v1, ok1 := util.FormatFloat(c1[item.Name])
			v2, ok2 := util.FormatFloat(c2[item.Name])
			if ok1 && ok2 {
				if item.Op == "desc" {
					return v1 > v2
				} else {
					return v1 < v2
				}
			} else {
				if item.Op == "desc" {
					return util.ToString(c1[item.Name]) > util.ToString(c2[item.Name])
				} else {
					return util.ToString(c1[item.Name]) < util.ToString(c2[item.Name])
				}
			}
		}
	}
	fn(less...).sorts(dt.Rows)
	return dt
}

type lessFunc func(p1, p2 map[string]interface{}) bool

type multiSorter struct {
	changes []map[string]interface{}
	less    []lessFunc
}

func (ms *multiSorter) sorts(changes []map[string]interface{}) {
	ms.changes = changes
	sort.Sort(ms)
}

func fn(less ...lessFunc) *multiSorter {
	return &multiSorter{
		less: less,
	}
}

func (ms *multiSorter) Len() int {
	return len(ms.changes)
}

func (ms *multiSorter) Swap(i, j int) {
	ms.changes[i], ms.changes[j] = ms.changes[j], ms.changes[i]
}

func (ms *multiSorter) Less(i, j int) bool {
	p, q := ms.changes[i], ms.changes[j]
	// Try all but the last comparison.
	var k int
	for k = 0; k < len(ms.less)-1; k++ {
		less := ms.less[k]
		switch {
		case less(p, q):
			// p < q, so we have a decision.
			return true
		case less(q, p):
			// p > q, so we have a decision.
			return false
		}
		// p == q; try the next comparison.
	}
	return ms.less[k](p, q)
}
