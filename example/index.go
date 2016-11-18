package main

import (
	. "github.com/coral"                     // common package
	db "github.com/coral/db"                 // db package
	config "github.com/coral/example/config" // config package
	. "github.com/coral/example/constant"    // constant package
	filter "github.com/coral/example/filter" // filter package
)

/**
 * initRouter 方法，定义服务全部接口、参数校验、并指定过滤器链
 */
func initRouter(server *Server) {
	// r for param check
	r := &R{}

	// curl http://localhost:8080/
	// hello coral
	baseRouter := server.NewRouter("/", filter.Index)

	// curl http://localhost:8080/param?<params>
	// <params>
	paramRouter := baseRouter.NewRouter("param", filter.Param)
	// curl http://localhost:8080/param/check?a=1&b=2&c=1
	// {"a":"1", "b":2, "c":"1"}
	paramRouter.NewRouter(
		"check",
		r.Check(V{
			"a": r.IsString,
			"b": r.IsInt,
			"c": r.IsBool}),
		filter.Param,
	)
	// TODO 更复杂的check

	// curl http://localhost:8080/mysql
	// something happend
	mysqlRouter := baseRouter.NewRouter("mysql", filter.Mysql)
	mysqlRouter.NewRouter("select", filter.Select)
	mysqlRouter.NewRouter("insert", filter.Insert)
	mysqlRouter.NewRouter("update", filter.Update)
}

func initDB() {
	// init db pool
	dbPool := db.InitDB()
	// add default db
	dbPool.AddDB(
		DEF_DEFAULT_DB,
		config.DEFAULT_DB_DSN,
		config.DEFAULT_DB_MAX_CONNECTION,
		config.DEFAULT_DB_MAX_IDLE)

	// add other db
	// ...
}

func main() {
	// new server
	server := NewServer(config.HOST)
	// new router
	initRouter(server)

	// init db
	initDB()

	// start server
	server.Run()
}
