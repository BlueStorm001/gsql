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

// Where 条件匹配 (a=1 and b=2) or (c=2 and d=3)
// Condition match
func (dt *DataTable) Where(query string) *DataTable {
	dt.mode = normal
	return dt.wheres(query)
}

// Like 模糊匹配 关键字(%) 例如 name='fred%'
// Like Fuzzy matching keyword (%) For example name='fred%''
func (dt *DataTable) Like(query string) *DataTable {
	dt.mode = likeMode
	return dt.wheres(query)
}

// Find 利用正则表达式查找 例如 name='^fred.{1,3}$'
// Find using regular expressions For example name='^fred.{1,3}$'
func (dt *DataTable) Find(query string) *DataTable {
	dt.mode = regXMode
	return dt.wheres(query)
}
