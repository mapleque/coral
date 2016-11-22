package db

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	. "coral/log"
)

// DBPool 类型， 是一个database容器，用于存储服务可能用到的所有db连接池
type DBPool struct {
	Pool map[string]*sql.DB
}

// DB 全局变量，允许用户可以在任何一个方法中调用并操作数据库
var DB *DBPool

// InitDB 方法，初始化全局变量，DB在使用之前必须初始化，且只能初始化一次
func InitDB() *DBPool {
	DB = &DBPool{}
	DB.Pool = make(map[string]*sql.DB)
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
	db.Pool[name] = dbConn
}

/**
 * Select 方法，返回查询结果数组
 */
func (db *DBPool) Select(
	database, sql string,
	params ...interface{}) map[string]interface{} {
	rows, err := db.Pool[database].Query(sql, params...)
	if err != nil {
		Error("db query error ", sql, err.Error())
		return nil
	}
	defer rows.Close()
	ret, err := processRows(rows)
	if err != nil {
		Error("db query error ", sql, err.Error())
		return nil
	}
	return ret
}

/**
 * Update 方法，返回受影响行数
 */
func (db *DBPool) Update(
	database, sql string,
	params ...interface{}) int64 {
	res, err := db.Pool[database].Exec(sql, params...)
	if err != nil {
		Error("db exec error ", sql, err.Error())
		return -1
	}
	num, err := res.RowsAffected()
	if err != nil {
		Error("db exec error ", sql, err.Error())
		return -1
	}
	return num
}

/**
 * Insert 方法，返回插入id
 */
func (db *DBPool) Insert(
	database, sql string,
	params ...interface{}) int64 {
	res, err := db.Pool[database].Exec(sql, params...)
	if err != nil {
		Error("db exec error ", sql, err.Error())
		return -1
	}
	id, err := res.LastInsertId()
	if err != nil {
		Error("db exec error ", sql, err.Error())
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
