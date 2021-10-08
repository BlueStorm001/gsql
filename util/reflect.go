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
	"errors"
	"reflect"
)

func SetStruct(in interface{}, rows []map[string]interface{}) error {
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
