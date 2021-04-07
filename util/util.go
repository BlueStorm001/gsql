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

package util

import (
	"reflect"
	"regexp"
	"strings"
	"unsafe"
)

//nocopy 转换string
func BytToStr(src []byte) (dst string) {
	s := (*reflect.SliceHeader)(unsafe.Pointer(&src))
	d := (*reflect.StringHeader)(unsafe.Pointer(&dst))
	d.Len = s.Len
	d.Data = s.Data
	return
}

func Verify(sql string) bool {
	str := `(?:')|(?:--)|(/\\*(?:.|[\\n\\r])*?\\*/)|(\b(select|update|delete|insert|trancate|char|chr|into|substr|ascii|declare|exec|count|master|into|drop|execute)\b)`
	re, err := regexp.Compile(str)
	if err != nil {
		return true
	}
	return re.MatchString(sql)
}

func GetFieldName(w string) string {
	bytes := []byte(w)
	builder := Builder{}
	for _, b := range bytes {
		switch b {
		case ' ':
			if builder.Len() > 0 {
				d := strings.ToLower(builder.String())
				if d == "and" || d == "or" {
					builder.Reset()
				}
			}
		case
			'=',
			'?',
			'>',
			'<',
			'!',
			'%':
			break
		default:
			builder.AppendByte(b)
		}
	}
	return builder.String()
}
