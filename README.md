# gsql
golang sql orm
``` golang
import (
    "fmt"
    "github.com/BlueStorm001/gsql"
)
```

``` golang
var serve = gsql.NewServer("127.0.0.1", 3306, "database", gsql.MySql).
NewAuth("username", "password").
NewConfig(100, 60)
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
    if orm.Select("*").OrderBy("id desc").Pagination(20,1).Execute().Error == nil {
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