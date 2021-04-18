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
	"testing"
)

func TestExpr(t *testing.T) {
	valid := []string{
		"id=1 and code=2",
		"(id=1 and code=2) or (id=2 and code=3)",
		"name='wen' and ( code1=1 or code2=2 ) and title=3",
		"name='wen' and (status1=1 or status2=2 or ( code1=1 and code2=2 ) )",
		"(name='wen' or name='wu' or(name='dong' and status=1 or(code=3 and code=4 and (ok=6 or ok=7) ) ) ) and status=1",
		"( ( code1=1 and code2=2 ) or status1=1 or status2=2)and name='wen' and (status1=1 or status2=2 or ( code1=1 and code2=2 ) )",
		"",
	}
	for i, s := range valid {
		exp := Parse([]byte(s)) //(col=123 or name='wu') and (col1=321 or col2=111) and(status=1 or status=2)
		t.Log(i, exp.Tree)
	}
}
