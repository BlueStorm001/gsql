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
	"strconv"
	"strings"
	"time"
	"unsafe"
)

// BytToStr nocopy 转换string
func BytToStr(src []byte) (dst string) {
	s := (*reflect.SliceHeader)(unsafe.Pointer(&src))
	d := (*reflect.StringHeader)(unsafe.Pointer(&dst))
	d.Len = s.Len
	d.Data = s.Data
	return
}

// ToString exp: float:2保留两位小数
func ToString(value interface{}) string {
	switch v := value.(type) {
	case float32, float64:
		f := ToFloat64(v)
		return strconv.FormatFloat(f, 'f', -1, 64)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		f := ToInt64(v)
		return strconv.FormatInt(f, 10)
	case []byte:
		return BytToStr(v)
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	case nil:
		return ""
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	default:
		return ""
	}
}

func ToFloat64(value interface{}) float64 {
	switch v := value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		f := ToInt(v)
		return float64(f)
	case float64:
		return v
	case float32:
		return float64(v)
	case bool:
		if v {
			return 1
		}
		return 0
	case []byte:
		str := BytToStr(v)
		f, _ := strconv.ParseFloat(str, 64)
		return f
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	default:
		return 0
	}
}

func ToInt(value interface{}) int {
	switch v := value.(type) {
	case float64, float32:
		f := ToFloat64(v)
		return int(f)
	case string:
		f, _ := strconv.Atoi(v)
		return f
	case int:
		return v
	case uint:
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case bool:
		if v {
			return 1
		}
		return 0
	case []byte:
		f, _ := strconv.Atoi(BytToStr(v))
		return f
	default:
		return 0
	}
}

func ToInt64(value interface{}) int64 {
	return int64(ToInt(value))
}

func FormatFloat(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float32, float64:
		return ToFloat64(v), true
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return float64(ToInt(v)), true
	default:
		return 0, false
	}
}

func ToNowStr() string {
	return ToDateTimeStr(time.Now())
}

func ToDateTimeStr(d time.Time) string {
	return d.Format("2006-01-02 15:04:05")
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

func WhetherToSkip(mode int, columns map[string]struct{}, column string) bool {
	if mode != 0 {
		_, matched := columns[column]
		switch mode {
		case 1:
			if !matched {
				return true
			}
		case -1:
			if matched {
				return true
			}
		}
	}
	return false
}
