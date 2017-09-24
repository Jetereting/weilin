package mysqlUtility

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/astaxie/beego"

	_ "github.com/go-sql-driver/mysql"
	"weilin/mysqlUtility/dal"
)

// begin with capitial wordd so it can be accessed by outer
var (
	DBConn   *sql.DB
	DContext dal.DataContext

	// 数据库配置
	DB_CONF DbConfig
)

// DbConfig 数据库连接参数
type DbConfig struct {
	IP        string
	Port      int64
	User      string
	Pwd       string
	Database  string
	IdleConns int
}

func StartInit() {
	DB_CONF = DbConfig{
		IP:        beego.AppConfig.String("mysql::host"),
		Port:      beego.AppConfig.DefaultInt64("mysql::port", 3306),
		User:      beego.AppConfig.String("mysql::user"),
		Pwd:       beego.AppConfig.String("mysql::pwd"),
		Database:  beego.AppConfig.String("mysql::database"),
		IdleConns: beego.AppConfig.DefaultInt("mysql::idleconns", 100),
	}
	DBConn = ConnectToDB()
}

// OpenDB 打开数据库
func OpenDB() (*sql.DB, error) {
	config := DB_CONF
	if config.IP == "" {
		return nil, errors.New("未配置数据库连接")
	}
	return OpenDBForConfig(config)
}

// OpenDBForConfig 连接数据库
// 通过config参数构建数据库连接
func OpenDBForConfig(config DbConfig) (*sql.DB, error) {
	sqlConn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&multiStatements=true", config.User, config.Pwd, config.IP, config.Port, config.Database)
	db, err := sql.Open("mysql", sqlConn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(config.IdleConns)
	DContext = dal.GetContext(db)
	beego.Info("数据库初始化成功...")
	return db, nil
}

// ConnectToDB 打开数据库
func ConnectToDB() *sql.DB {
	dbConn, err := OpenDB()
	if err != nil {
		beego.Error("数据库初始化错误:", err.Error())
		os.Exit(0)
	}
	return dbConn
}

// PingDB 测试连接
func PingDB(ticker *time.Ticker) {
	for range ticker.C {
		if err := DBConn.Ping(); err != nil {
			DBConn = ConnectToDB()
			ticker.Stop()
			return
		}
	}
}

// StartTransaction 开始事务
//map[string]interface{}{Sql:string ,Values:[][]interface{}}
func StartTransaction(data []map[string]interface{}) error {
	tx, err := DBConn.Begin()
	if err != nil {
		return err
	}
	for _, v := range data {
		stmt, err := tx.Prepare(v["Sql"].(string))
		if err != nil {
			tx.Rollback()
			return err
		}
		for _, v := range v["Values"].([]interface{}) {
			_, err = stmt.Exec(v.([]interface{})...)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	err = tx.Commit()
	return err
}

// StartToTransaction 开始事务
//map[string]interface{}{Sql:string ,Values:[][]interface{}}
func StartToTransaction(data []map[string]interface{}) error {
	tx, err := DBConn.Begin()
	if err != nil {
		return err
	}
	for _, v := range data {
		stmt, err := tx.Prepare(v["Sql"].(string))
		if err != nil {
			tx.Rollback()
			return err
		}
		valuesData := v["Values"].([]interface{})
		value := make([]interface{}, 0, len(valuesData))
		for _, v := range valuesData {
			value = append(value, v.([]interface{})...)
		}
		_, err = stmt.Exec(value...)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	err = tx.Commit()
	return err
}
