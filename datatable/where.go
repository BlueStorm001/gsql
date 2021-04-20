package datatable

import (
	"gsql/util"
	"regexp"
	"strconv"
	"strings"
)

func (dt *DataTable) wheres(query string) *DataTable {
	exp, err := Where([]byte(query))
	if err != nil {
		return dt
	}
	dataTable := dt.match(exp.WhereExpr)
	dataTable.Name = dt.Name
	dataTable.Columns = dt.Columns
	return dataTable
}

func (dt *DataTable) DeleteSymbolKey() *DataTable {
	for _, row := range dt.Rows {
		delete(row, "$key$")
	}
	return dt
}

func (dt *DataTable) match(w *Wheres) *DataTable {
	if wh := getWheres(w.Lh); wh != nil {
		dt.match(wh)
	}
	if wh := getWheres(w.Rh); wh != nil {
		dt.match(wh)
	}
	return dt.routing(w)
}

func (dt *DataTable) routing(w *Wheres) *DataTable {
	switch w.Op {
	case "and", "or":
		w.R = dt.logic(w)
	case "=", "<>", "!=", ">", "<", ">=", "<=":
		w.R = dt.contrast(w)
	}
	return w.R.(*DataTable)
}

func (dt *DataTable) logic(w *Wheres) *DataTable {
	var rows []map[string]interface{}
	switch w.Op {
	case "and":
		for _, dr := range getRows(w.Lh) {
			field := getField(w.Lh)
			if field != dr["$key$"] {
				rows = append(rows, dr)
			}
		}
	case "or":
		lr := getRows(w.Lh)
		if lr != nil {
			rows = append(rows, lr...)
		}
		rr := getRows(w.Rh)
		if rr != nil {
			rows = append(rows, rr...)
		}
	}
	return &DataTable{Count: len(rows), Rows: rows}
}

func (dt *DataTable) contrast(w *Wheres) *DataTable {
	lhStr := w.Lh.(string)
	rhStr := w.Rh.(string)
	var number float64
	var use bool
	var err error
	var rows []map[string]interface{}
	var (
		like bool
		regx bool
	)
	switch dt.mode {
	case likeMode:
		if strings.Contains(rhStr, "%") {
			like = true
		}
	case regXMode:
		regx = true
	}
	input := rhStr
	for _, row := range dt.Rows {
		if value, yes := row[lhStr]; yes {
			if like {
				var ok bool
				input, ok = likeValue(value, rhStr)
				if !ok {
					continue
				}
			}
			if regx {
				var ok bool
				input, ok = findValue(value, rhStr)
				if !ok {
					continue
				}
			}
			var verify bool
			switch w.Op {
			case ">", "<", ">=", "<=":
				if !use {
					number, err = strconv.ParseFloat(input, 64)
				}
				data, ok := util.FormatFloat(value)
				if err == nil && ok {
					use = true
					switch w.Op {
					case ">":
						if data > number {
							verify = true
						}
					case "<":
						if data < number {
							verify = true
						}
					case ">=":
						if data >= number {
							verify = true
						}
					case "<=":
						if data <= number {
							verify = true
						}
					}
				}
			case "=":
				if input == util.ToString(value) {
					verify = true
				}
			case "!=", "<>":
				if input != util.ToString(value) {
					verify = true
				}
			}
			if verify {
				row["$key$"] = lhStr
				rows = append(rows, row)
			}
		}
	}
	return &DataTable{Count: len(rows), Rows: rows}
}

func likeValue(v interface{}, value string) (string, bool) {
	dataStr := util.ToString(v)
	args := strings.Split(value, "%")
	switch len(args) {
	case 2:
		if args[0] == "" {
			value = args[1] + "$"
		} else {
			value = "^" + args[0]
		}
	case 3:
		value = strings.Replace(value, "%", "", -1)
	}
	matched, _ := regexp.MatchString(value, dataStr)
	return dataStr, matched
}

func findValue(v interface{}, value string) (string, bool) {
	dataStr := util.ToString(v)
	value = strings.Replace(value, "%", ".", -1)
	matched, _ := regexp.MatchString(value, dataStr)
	return dataStr, matched
}

func getField(h interface{}) string {
	var rs string
	switch r := h.(type) {
	case *Wheres:
		rs = getField(r.Lh)
		if rs != "" {
			return rs
		}
	default:
		rs = r.(string)
	}
	return rs
}

func getWheres(h interface{}) *Wheres {
	switch r := h.(type) {
	case *Wheres:
		return r
	}
	return nil
}

func getRows(h interface{}) []map[string]interface{} {
	if w := getWheres(h); w != nil {
		switch tb := w.R.(type) {
		case *DataTable:
			return tb.Rows
		}
	}
	return nil
}
