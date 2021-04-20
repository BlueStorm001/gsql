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
//第一种使用方式（推荐）
//The first way to use
//MYSQL
func MySqlConnDrive() (db *sql.DB, err error) {
	connString := "user:pass@tcp(host:3306)/database?charset=utf8"
	db, err = sql.Open("mysql", connString)
	return
}
//更简单的使用sql驱动
//只需要制定的使用的sql类型和驱动即可
//Easier to use sql driver
var serve = gsql.NewDrive(gsql.MySql, MySqlConnDrive).Config(100, 60)
```

``` golang
//第二种使用方式
//The second way to use
//SQLSERVER
func MSSqlDrive(s *datatable.Serve) (db *sql.DB, err error) {
    connString := fmt.Sprintf("server=%s;database=%s;user id=%s;password=%s;port=%d", s.Host, s.Database, s.Auth.User, s.Auth.Pass, s.Port)
    conn, err := mssql.NewAccessTokenConnector(connString,
    func() (string, error) {
        return "", nil
    })
    if err != nil {
        return
    }
    db = sql.OpenDB(conn)
    return
}
//MYSQL
func MySqlDrive(s *datatable.Serve) (db *sql.DB, err error) {
    connString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", s.Auth.User, s.Auth.Pass, s.Host, s.Port, s.Database)
    db, err = sql.Open("mysql", connString)
}
//使用用户密码后调用驱动
//Call the sql driver after using the user password
var serve = gsql.NewServer("host", 3306).
	Database(gsql.MySql, "database").
	Login("user", "password").
	Config(connectMax, timeout).NewDrive(MySqlDrive)
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
    if orm.Insert().Execute().Error == nil {
        fmt.Println(orm.LastInsertId)
    }
}
```

``` golang
func main() {
    option := &Options{Id:1,Text:"test"}
    orm := serve.NewStruct("table_options", option)
    if orm.Select().Where("Id=?").OrderBy("id desc").Execute().Error == nil {
        if orm.Result.RowsAffected > 0 {
            fmt.Println("row:", orm.Id, option.Id, orm.Result.DataTable.Rows[0]["Id"], orm.TC)
            
            //在结果集里进行搜索
            //Search in the result set
            dt := orm.Result.DataTable
            
            // Where 条件匹配 (a=1 and b=2) or (c=2 and d=3) Condition match
            table := dt.Where("Text='CN' and (code='BJS' or code='SHA')").OrderBy("id") 
            for i, row := range table.Rows {
                fmt.Println(i,row)
            }
            
            // 使用模糊搜索 Use fuzzy search
            table = dt.Like("name='CN%' and money=1.2%").OrderBy("id desc")
            
            // 使用正则表达式 Use regular expressions
            table = dt.Find("code='[A-Z]{3}'").OrderBy("id desc")
            
            // Group By 分组
            table = dt.GroupBy("name")
            for i, row := range table.Rows {
                newTable := dt.Where("name='" + row["name"] + "' and (code='BJS' or code='SHA')").OrderBy("id") //[id asc , name desc]...
                fmt.Println(newTable)
            }
            
        } else {
            fmt.Println("row:", orm.Id, option.Id, "no data", orm.TC)
        }
    } else {
        fmt.Println(orm.Error)
    }
}
```

``` golang
func main() {
    option := &Options{Id:1,Text:"test"}
    orm := serve.NewStruct("table_options", option)
    if orm.Select("*").OrderBy("id desc").Page(20,1).Execute().Error == nil {
        if orm.Result.RowsAffected > 0 {
            fmt.Println("row:", orm.Id, option.Id, orm.Result.DataTable.Rows[0]["Id"], orm.TC)
        } else {
            fmt.Println("row:", orm.Id, option.Id, "no data", orm.TC)
        }
    } else {
        fmt.Println(orm.Error)
    }
}
```

``` golang
func main() {
    option := &Options{Id:1,Text:"test"}
    orm := serve.NewStruct("table_options", option)
    if orm.Select("Text").GroupBy("Text").Execute().Error == nil {
        if orm.Result.RowsAffected > 0 {
            fmt.Println("row:", orm.Id, option.Id, orm.Result.DataTable.Rows[0]["Id"], orm.TC)
        } else {
            fmt.Println("row:", orm.Id, option.Id, "no data", orm.TC)
        }
    } else {
        fmt.Println(orm.Error)
    }
}
```

``` golang
func main() {
    option := &Options{Id:1,Text:"test"}
    orm := serve.NewStruct("table_options", option)
    if orm.Count().Where("Id=?").Execute().Error == nil {
        fmt.Println(orm.RowsAffected)
    }
}
```

``` golang
func main() {
    option := &Options{Id:1,Text:"test"}
    orm := serve.NewStruct("table_options", option)
    if orm.Update().Where("Id=?").Execute().Error == nil {
        fmt.Println(orm.RowsAffected)
    }
}
```

``` golang
func main() {
    option := &Options{Id:1,Text:"test"}
    orm := serve.NewStruct("table_options", option)
    if orm.Update("value").Where("Id=?").Execute().Error == nil {
        fmt.Println(orm.RowsAffected)
    }
}
```

``` golang
func main() {
    option := &Options{Id:1,Text:"test"}
    orm := serve.NewStruct("table_options", option)
    if orm.UpdateExclude("Text").Where("Id=?").Execute().Error == nil {
        fmt.Println(orm.RowsAffected)
    }
}
```

``` golang
func main() {
    option := &Options{Id:1,Text:"test"}
    orm := serve.NewStruct("table_options", option)
    if orm.Delete().Where("Id>=?","and Text=?").Execute().Error == nil {
        fmt.Println(orm.RowsAffected)
    }
}
```