package dal

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"weilin/utility"
)

type MysqlDal struct {
	db *sql.DB
}

func (dal *MysqlDal) GetDb() *sql.DB {
	return dal.db
}

func (dal *MysqlDal) Exec(sqlText string, args ...interface{}) (sql.Result, error) {
	//输出日志
	// alog.Info("|{SQL_MESSAGE}::[", sqlText, "]|{VALUES}::", args)
	stmt, err := dal.db.Prepare(sqlText)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	return stmt.Exec(args...)
}

func (dal *MysqlDal) Insert(entity Entity) (int64, error) {
	var (
		i            int
		fields       string
		placeHolders string
		result       sql.Result
		err          error
	)
	fieldsValue := entity["FieldsValue"].(map[string]interface{})
	values := make([]interface{}, len(fieldsValue))
	for key, value := range fieldsValue {
		fields += "," + key
		placeHolders += ",?"
		values[i] = value
		i++
	}
	sqlText := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", entity["Table"].(string), fields[1:], placeHolders[1:])
	if tx, ok := entity["TX"]; ok {
		result, err = tx.(*sql.Tx).Exec(sqlText, values...)
	} else {
		result, err = dal.Exec(sqlText, values...)
	}
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (dal *MysqlDal) getCondition(entity Entity, values *[]interface{}) string {
	var condition string
	if c, ok := entity["Condition"]; ok {
		condition = utility.NewT(c).ToString()
		if vals, ok := entity["Values"]; ok && vals != nil {
			*values = append(*values, vals.([]interface{})...)
		}
	} else if f, ok := entity["FieldsPk"]; ok {
		fieldsPk := f.(map[string]interface{})
		arrFields := make([]string, 0, len(fieldsPk))
		for key, value := range fieldsPk {
			arrFields = append(arrFields, fmt.Sprintf("%s=?", key))
			*values = append(*values, value)
		}
		condition = fmt.Sprintf(" WHERE %s ", strings.Join(arrFields, " AND "))
	}
	return condition
}

func (dal *MysqlDal) Update(entity Entity) (int64, error) {
	fieldsValue := entity["FieldsValue"].(map[string]interface{})
	values := make([]interface{}, 0, len(fieldsValue))
	fields := make([]string, 0, len(fieldsValue))
	for key, value := range fieldsValue {
		fields = append(fields, fmt.Sprintf("%s=?", key))
		values = append(values, value)
	}
	condition := dal.getCondition(entity, &values)
	sqlText := fmt.Sprintf("UPDATE %s SET %s %s ", entity["Table"].(string), strings.Join(fields, ","), condition)
	var (
		result sql.Result
		err    error
	)
	if tx, ok := entity["TX"]; ok {
		result, err = tx.(*sql.Tx).Exec(sqlText, values...)
	} else {
		result, err = dal.Exec(sqlText, values...)
	}
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (dal *MysqlDal) Delete(entity Entity) (int64, error) {
	var values []interface{}
	condition := dal.getCondition(entity, &values)
	sqlText := fmt.Sprintf("DELETE FROM %s %s ", entity["Table"].(string), condition)
	var (
		result sql.Result
		err    error
	)
	if tx, ok := entity["TX"]; ok {
		result, err = tx.(*sql.Tx).Exec(sqlText, values...)
	} else {
		result, err = dal.Exec(sqlText, values...)
	}
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (dal *MysqlDal) GetSingleData(entity Entity) (data map[string]string, err error) {
	datas, err := dal.Search(entity)
	if err != nil {
		return
	}
	if len(datas) > 0 {
		data = datas[0]
		return
	}
	return make(map[string]string), nil
}

func (dal *MysqlDal) QueryData(sqlText string, values ...interface{}) (data []map[string]string, err error) {
	rows, err := dal.db.Query(sqlText, values...)
	if err != nil {
		return
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return
	}
	for rows.Next() {
		values := make([]sql.RawBytes, len(columns))
		scanArgs := make([]interface{}, len(columns))
		for i, _ := range values {
			scanArgs[i] = &values[i]
		}
		if err = rows.Scan(scanArgs...); err != nil {
			break
		}
		columnsMp := make(map[string]string, len(columns))
		for i, val := range values {
			var value string
			if val != nil {
				value = string(val)
			}
			columnsMp[columns[i]] = value
		}
		data = append(data, columnsMp)
	}
	if len(data) == 0 {
		data = make([]map[string]string, 0)
	}
	return
}

func (dal *MysqlDal) getQuerySql(entity Entity) (query string, values []interface{}) {
	condition := dal.getCondition(entity, &values)
	fieldsSelect := "*"
	if f, ok := entity["FieldsSelect"]; ok {
		fieldsSelect = f.(string)
	}
	query = fmt.Sprintf("SELECT %s FROM %s %s ", fieldsSelect, entity["Table"].(string), condition)
	if group, ok := entity["Group"]; ok {
		query = fmt.Sprintf("%s GROUP BY %s", query, utility.NewT(group).ToString())
	}
	if order, ok := entity["OrderBy"]; ok {
		query = fmt.Sprintf("%s Order By %s", query, utility.NewT(order).ToString())
	}
	if l, ok := entity["LIMIT"]; ok {
		query = fmt.Sprintf("%s LIMIT %s", query, utility.NewT(l).ToString())
	}
	return
}

func (dal *MysqlDal) Search(entity Entity) (data []map[string]string, err error) {
	sqlText, values := dal.getQuerySql(entity)
	//输出日志
	// alog.Info("|{SQL_MESSAGE}::[", sqlText, "]|{VALUES}::", values)
	return dal.QueryData(sqlText, values...)
}

func (dal *MysqlDal) queryCount(entity Entity, w *sync.WaitGroup, count *int, err *error) {
	defer w.Done()
	var values []interface{}
	condition := dal.getCondition(entity, &values)
	sqlText := fmt.Sprintf("SELECT COUNT(*) FROM %s %s ", entity["Table"].(string), condition)
	stmt, sterr := dal.db.Prepare(sqlText)
	if sterr != nil {
		*err = sterr
		return
	}
	defer stmt.Close()
	*err = stmt.QueryRow(values...).Scan(count)
	return
}

func (dal *MysqlDal) QueryPager(entity Entity) (data map[string]interface{}, err error) {
	var (
		count    int
		datas    []map[string]string
		countErr error
		queryErr error
	)
	w := new(sync.WaitGroup)
	w.Add(2)
	go dal.queryCount(entity, w, &count, &countErr)
	go func(entity Entity, w *sync.WaitGroup, datas *[]map[string]string, errInfo *error) {
		defer w.Done()
		sqlText, values := dal.getQuerySql(entity)
		// if group, ok := entity["Group"]; ok {
		// 	sqlText = fmt.Sprintf("%s GROUP BY %s", sqlText, utility.NewT(group).ToString())
		// }
		pageIndex, indexOK := entity["PageIndex"]
		pageSize, sizeOK := entity["PageSize"]
		if indexOK && sizeOK {
			index := utility.NewT(pageIndex).ToInt64()
			size := utility.NewT(pageSize).ToInt64()
			sqlText = fmt.Sprintf("%s LIMIT %d,%d", sqlText, index, size)
		}
		//输出日志
		// alog.Info("|{SQL_MESSAGE}::[", sqlText, "]|{VALUES}::", values)

		//		fmt.Println("----------------Query pager sql--------------------")
		//		fmt.Println("===> Query pager:", sqlText)
		//		fmt.Println("===> Query value:", values)
		data, err := dal.QueryData(sqlText, values...)
		if err != nil {
			*errInfo = err
			return
		}
		*datas = data
	}(entity, w, &datas, &queryErr)
	w.Wait()
	if countErr != nil {
		err = countErr
		return
	}
	if queryErr != nil {
		err = queryErr
		return
	}
	data = map[string]interface{}{
		"total": count,
		"items": datas,
	}
	return
}

func (dal *MysqlDal) Query(entity Entity) (interface{}, error) {
	if table, ok := entity["Table"]; !ok || table.(string) == "" {
		return nil, fmt.Errorf("Query requires table name.")
	}
	switch utility.NewT(entity["Operate"]).ToString() {
	case "G":
		return dal.GetSingleData(entity)
	case "Q":
		return dal.Search(entity)
	case "QP":
		return dal.QueryPager(entity)
	}
	return nil, fmt.Errorf("No corresponding operation")
}

func (dal *MysqlDal) ExecTransaction(listEntity []Entity) error {
	tx, err := dal.db.Begin()
	if err != nil {
		return err
	}
	var errInfo error
	for _, entity := range listEntity {
		entity["TX"] = tx
		switch oper := entity["Operate"].(string); strings.ToUpper(oper) {
		case "A":
			if _, err := dal.Insert(entity); err != nil {
				errInfo = err
			}
		case "U":
			if _, err := dal.Update(entity); err != nil {
				errInfo = err
			}
		case "D":
			if _, err := dal.Delete(entity); err != nil {
				errInfo = err
			}
		}
		if errInfo != nil {
			break
		}
	}
	if errInfo != nil {
		tx.Rollback()
		return errInfo
	}
	return tx.Commit()
}

func (dal *MysqlDal) ExecSqlTransaction(sqlList []string) error {
	tx, err := dal.db.Begin()
	if err != nil {
		return err
	}
	var errInfo error
	for _, sql := range sqlList {
		if _, err := tx.Exec(sql); err != nil {
			errInfo = err
		}
		if errInfo != nil {
			break
		}
	}
	if errInfo != nil {
		tx.Rollback()
		return errInfo
	}
	return tx.Commit()
}
