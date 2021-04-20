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
		"a=1 and b=2",
		"(a=1 and b=2) or (c=2 and d=3)",
		"a='wen' and ( b=1 or c=2 ) and d=3",
		"a='wen' and (b=1 or c=2 or ( d=1 and e=2 ) )",
		"(a='wen' or b='wu' or(c='dong' and d=1 or(e=3 and f=4 and (g=6 or h=7) ) ) ) and i=1",
		"( ( a=1 and b=2 ) or c=1 or d=2)and e='wen' and (f=1 or g=2 or ( h=1 and i=2 ) )",
		"",
	}
	for i, s := range valid {
		exp, _ := Where([]byte(s)) //(col=123 or name='wu') and (col1=321 or col2=111) and(status=1 or status=2)
		t.Log(i, exp.WhereExpr)
	}
}
