package dal

import (
	"database/sql"
	"fmt"
)

// Entity entity
type Entity map[string]interface{}

// GetContext 获取数据库操作实例
func GetContext(dbConn *sql.DB) DataContext {
	return &MysqlDal{dbConn}
}

// NewContext Create new context instance.
func NewContext(driver, dsn string) DataContext {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		fmt.Println("===> Database ping error:", err)
	}
	return &MysqlDal{db}
}

// NewEntity Create new entity.
// Table : table tableName
// Operate : Query:{"G": single data "Q": query data "QP": limit query} Insert:"A" Update:"U" Delete:"D"
// FieldsSelect : select 字段
// FieldsPk : where 条件
// Group : group by 字段
// Order By : order by 字段
// LIMIT : limit 直接分页 参数如：0,1
// PageIndex : 分页开始
// PageSize : 分页大小
func NewEntity(tableName, operate string, items ...map[string]interface{}) Entity {
	entity := Entity{
		"Table":   tableName,
		"Operate": operate,
	}
	if len(items) > 0 {
		for key, val := range items[0] {
			entity[key] = val
		}
	}
	return entity
}
