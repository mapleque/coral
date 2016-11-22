package db

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	. "coral/log"
)

// DBPool 类型， 是一个database容器，用于存储服务可能用到的所有db连接池
type DBPool struct {
	Pool map[string]*DBQuery
}

// DBQuery 对象，用于直接查询或执行
type DBQuery struct {
	database string
	conn     *sql.DB
}

// DBTransaction 对象，用于事物查询或执行
type DBTransaction struct {
	database string
	conn     *sql.Tx
}

// DB 全局变量，允许用户可以在任何一个方法中调用并操作数据库
var DB *DBPool

// InitDB 方法，初始化全局变量，DB在使用之前必须初始化，且只能初始化一次
func InitDB() *DBPool {
	DB = &DBPool{}
	DB.Pool = make(map[string]*DBQuery)
	return DB
}

/**
 * AddDB 方法，添加一个database，并且在启动前验证其连通性
 */
func (db *DBPool) AddDB(name, dsn string, maxOpenConns, maxIdleConns int) {
	dbConn, err := sql.Open("mysql", dsn)
	if err != nil {
		Error("can not open db ", dsn, err.Error())
		panic(err.Error())
	}
	//	defer db.Close()
	// 如果这里defer，这里刚添加完的db就会被关掉
	err = dbConn.Ping()
	if err != nil {
		Error("can not ping db ", dsn, err.Error())
		panic(err.Error()) // 直接panic，让server无法启动
	}
	dbConn.SetMaxOpenConns(maxOpenConns)
	dbConn.SetMaxIdleConns(maxIdleConns)
	dbQuery := &DBQuery{}
	dbQuery.conn = dbConn
	dbQuery.database = name
	db.Pool[name] = dbQuery
}

// UserDB 方法，返回DBQuery对象
func (db *DBPool) UseDB(database string) *DBQuery {
	return db.Pool[database]
}

// Begin 方法，返回DBTransaction对象
func (db *DBPool) Begin(database string) *DBTransaction {
	return db.Pool[database].Begin()
}

// Select 方法，返回查询结果数组
func (db *DBPool) Select(
	database, sql string,
	params ...interface{}) map[string]interface{} {

	return db.Pool[database].Select(sql, params...)
}

// Update 方法，返回受影响行数
func (db *DBPool) Update(
	database, sql string,
	params ...interface{}) int64 {

	return db.Pool[database].Update(sql, params...)
}

// Insert 方法，返回插入id
func (db *DBPool) Insert(
	database, sql string,
	params ...interface{}) int64 {

	return db.Pool[database].Insert(sql, params...)
}

// Begin 方法，返回DBTransaction对象
func (dbq *DBQuery) Begin() *DBTransaction {
	trans := &DBTransaction{}
	conn, err := dbq.conn.Begin()
	if err != nil {
		Error("db create transaction faild", dbq.database, err.Error())
		return trans
	}
	trans.conn = conn
	return trans
}

// Select 方法，返回查询结果数组
func (dbq *DBQuery) Select(
	sql string,
	params ...interface{}) map[string]interface{} {

	return processQueryRet(dbq.conn.Query(sql, params...))
}

// Update 方法，返回受影响行数
func (dbq *DBQuery) Update(
	sql string,
	params ...interface{}) int64 {

	return processUpdateRet(dbq.conn.Exec(sql, params...))
}

// Insert 方法，返回插入id
func (dbq *DBQuery) Insert(
	sql string,
	params ...interface{}) int64 {

	return processInsertRet(dbq.conn.Exec(sql, params...))
}

// Select 方法，返回查询结果数组
func (dbt *DBTransaction) Select(
	sql string,
	params ...interface{}) map[string]interface{} {

	return processQueryRet(dbt.conn.Query(sql, params...))
}

// Update 方法，返回受影响行数
func (dbt *DBTransaction) Update(
	database, sql string,
	params ...interface{}) int64 {

	return processUpdateRet(dbt.conn.Exec(sql, params...))
}

// Insert 方法，返回插入id
func (dbt *DBTransaction) Insert(
	database, sql string,
	params ...interface{}) int64 {

	return processInsertRet(dbt.conn.Exec(sql, params...))
}

// Commit 方法，提交事物
func (dbt *DBTransaction) Commit() {
	err := dbt.conn.Commit()
	if err != nil {
		Error("db transaction commit faild ", err.Error())
	}
}

// Rollback 方法，回滚事物
func (dbt *DBTransaction) Rollback() {
	err := dbt.conn.Rollback()
	if err != nil {
		Error("db transaction rollback faild ", err.Error())
	}
}

// 返回查询结果数组
func processQueryRet(rows *sql.Rows, err error) map[string]interface{} {
	if err != nil {
		Error("db query error ", err.Error())
		return nil
	}
	defer rows.Close()
	ret, err := processRows(rows)
	if err != nil {
		Error("db query error ", err.Error())
		return nil
	}
	return ret
}

// 返回受影响行数
func processUpdateRet(res sql.Result, err error) int64 {
	if err != nil {
		Error("db update error ", err.Error())
		return -1
	}
	num, err := res.RowsAffected()
	if err != nil {
		Error("db update error ", err.Error())
		return -1
	}
	return num
}

// 返回插入id
func processInsertRet(res sql.Result, err error) int64 {
	if err != nil {
		Error("db exec error ", err.Error())
		return -1
	}
	id, err := res.LastInsertId()
	if err != nil {
		Error("db exec error ", err.Error())
		return -1
	}
	return id
}

/**
 * processRows 方法，将返回的rows封装为字典数组
 */
func processRows(rows *sql.Rows) (map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	var ret map[string]interface{}
	ret = make(map[string]interface{})
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}
		for i, col := range values {
			if col == nil {
				ret[columns[i]] = "null"
			} else {
				switch val := (*scanArgs[i].(*interface{})).(type) {
				case []byte:
					ret[columns[i]] = string(val)
					break
				default:
					ret[columns[i]] = val
				}
			}
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}
