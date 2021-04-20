package datatable

import "gsql/util"

func (dt *DataTable) GroupBy(query string) *DataTable {
	if dt.Count <= 1 {
		return dt
	}
	exp := GroupBy([]byte(query))
	count := len(exp.GroupExpr)
	if count == 0 {
		return dt
	}
	dataTable := &DataTable{Name: dt.Name, FindMode: dt.FindMode}
	for _, item := range exp.GroupExpr {
		for _, column := range dt.Columns {
			if column.Name == item.Name {
				dataTable.Columns = append(dataTable.Columns, column)
				break
			}
		}
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
