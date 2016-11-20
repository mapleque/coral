package main

import (
	coral "github.com/coral"
	db "github.com/coral/db"
	log "github.com/coral/log"

	config "github.com/coral/example/config"
	filter "github.com/coral/example/filter"
)

/**
 * initRouter 方法，定义服务全部接口、参数校验、并指定过滤器链
 */
func initRouter(server *coral.Server) {
	// r for param check
	r := &coral.R{}

	// /
	baseRouter := server.NewRouter("/", filter.Index)

	// /param?<params>
	paramRouter := baseRouter.NewRouter("param", filter.Param)
	// /param/check?a=1&b=2&c=1
	paramRouter.NewRouter(
		"check",
		r.Check(coral.V{
			"a": r.IsString,
			"b": r.IsInt,
			"c": r.IsBool}),
		filter.Param,
	)
	// TODO 更复杂的check

	// log
	baseRouter.NewRouter("log", filter.Log)

	// /mysql
	mysqlRouter := baseRouter.NewRouter("mysql", filter.Mysql)
	// /mysql/select
	mysqlRouter.NewRouter("select", filter.Select)
	// /mysql/insert
	mysqlRouter.NewRouter("insert", filter.Insert)
	// /mysql/update
	mysqlRouter.NewRouter("update", filter.Update)

	// /redis
	redisRouter := baseRouter.NewRouter("redis", filter.Redis)
	redisRouter.NewRouter("set", filter.Set)
	redisRouter.NewRouter("get", filter.Get)
}

func initDB() {
	// init db pool
	dbPool := db.InitDB()
	// add default db
	dbPool.AddDB(
		config.DEFAULT_DB,
		config.DEFAULT_DB_DSN,
		config.DEFAULT_DB_MAX_CONNECTION,
		config.DEFAULT_DB_MAX_IDLE)

	// add other db
	// ...
}

func initLog() {
	logPool := log.InitLog()
	logPool.AddLogger(
		config.DEFAULT_LOG,
		config.DEFAULT_LOG_PATH,
		config.DEFAULT_LOG_MAX_NUMBER,
		config.DEFAULT_LOG_MAX_SIZE,
		config.DEFAULT_LOG_MAX_LEVEL,
		config.DEFAULT_LOG_MIN_LEVEL)
}

func main() {
	// init log
	initLog()

	// init db
	initDB()

	// new server
	server := coral.NewServer(config.HOST)

	// new router
	initRouter(server)

	// start server
	server.Run()
}
