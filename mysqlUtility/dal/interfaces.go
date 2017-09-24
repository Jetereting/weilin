package dal

import (
	"database/sql"
)

type DataContext interface {
	GetDb() *sql.DB
	DbContext
	ExDbContext
}

type DbContext interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type ExDbContext interface {
	Insert(entity Entity) (int64, error)
	Update(entity Entity) (int64, error)
	Delete(entity Entity) (int64, error)
	Search(entity Entity) (data []map[string]string, err error)
	QueryPager(entity Entity) (data map[string]interface{}, err error)
	Query(entity Entity) (interface{}, error)
	GetSingleData(entity Entity) (data map[string]string, err error)
	ExecTransaction(listEntity []Entity) error
	ExecSqlTransaction(sqlList []string) error
	QueryData(sqlText string, values ...interface{}) (data []map[string]string, err error)
}
