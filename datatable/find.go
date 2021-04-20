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
