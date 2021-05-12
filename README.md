# gsql

golang SQL ORM

``` golang
import (
    "fmt"
    "github.com/BlueStorm001/gsql"
    mssql "github.com/denisenkom/go-mssqldb"
    _ "github.com/go-sql-driver/mysql"
)
```

``` golang
//使用方式（推荐）
//The first way to use
//MYSQL test
func MySqlConnDrive() (db *sql.DB, err error) {
	connString := "user:pass@tcp(host:3306)/database?charset=utf8"
	db, err = sql.Open("mysql", connString)
	return
}
//更简单的使用sql驱动
//只需要制定的使用的sql类型和驱动即可
//Easier to use sql driver
//gsql.MySql 
//gsql.MSSql
//gsql.Clickhouse
var serve = gsql.NewDrive(gsql.MySql, MySqlConnDrive).Config(100, 60)
```

``` golang
type Options struct {
    Id      int    `json:"id,string" sql:"primary key,auto_increment 1000"`
    Text string    `json:",string" sql:"varchar(20) default null"`
    Value   string
}
```

``` golang
func main() {
    option := &Options{Id:1,Text:"test"}
    orm := serve.NewStruct("table_options", option)
    //orm.Select("*")... or orm.Select("Id","Text")...
    //orm.Select().Page(20,1).Execute() //分页 pagination
    //orm.Count()...
    //orm.Insert()...
    //orm.Update()...
    //orm.Delete()...
    result := orm.Select().Where("Id=?").OrderBy("id desc").Execute()
    //result
    if result.RowsAffected > 0 {
        fmt.Println("row:", orm.Id, option.Id, result.DataTable.Rows[0]["Id"], orm.TC)
        
        // 在结果集里进行搜索
        // Search in the result set
        dt := result.DataTable
        
        // Where 条件匹配 (a=1 and b=2) or (c=2 and d=3) Condition match
        table := dt.Where("Text='CN' and (code='BJS' or code='SHA')").OrderBy("id") 
        for i, row := range table.Rows {
            fmt.Println(i,row)
        }
        
        // 使用模糊搜索 
        // Use fuzzy search
        table = dt.Like("name='CN%' and money=1.2%").OrderBy("id desc")
        
        // 使用正则表达式 
        // Use regular expressions
        table = dt.Find("code='[A-Z]{3}'").OrderBy("id desc")
        
        // 分组
        // Group
        table = dt.GroupBy("name")
        for i, row := range table.Rows {
            newTable := dt.Where("name='" + row["name"] + "' and (code='BJS' or code='SHA')").OrderBy("id") //[id asc , name desc]...
            fmt.Println(newTable)
        }
        
    } else {
        fmt.Println("row:", orm.Id, option.Id, "no data", orm.TC)
    }
   
}
```