package main

import (
	coral "github.com/coral"
	cache "github.com/coral/cache"
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
	redisRouter.NewRouter("set",
		r.Check(coral.V{
			"key": r.IsString,
			"val": r.IsInt}),
		filter.Set)
	redisRouter.NewRouter("get",
		r.Check(coral.V{
			"key": r.IsString}),
		filter.Get)
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

func initRedis() {
	// init cache pool
	cachePool := cache.InitCache()
	// add default cache
	cachePool.AddRedis(
		config.DEFAULT_REDIS,
		config.DEFAULT_REDIS_SERVER,
		config.DEFAULT_REDIS_AUTH,
		config.DEFAULT_REDIS_MAX_CONNECTION,
		config.DEFAULT_REDIS_MAX_IDLE)

	// add other cache
	// ...
}

func initLog() {
	// init log pool
	logPool := log.InitLog()
	// add default logger
	logPool.AddLogger(
		config.DEFAULT_LOG,
		config.DEFAULT_LOG_PATH,
		config.DEFAULT_LOG_MAX_NUMBER,
		config.DEFAULT_LOG_MAX_SIZE,
		config.DEFAULT_LOG_MAX_LEVEL,
		config.DEFAULT_LOG_MIN_LEVEL)

	// add other logger
	// ...
}

func main() {
	// init log
	initLog()

	// init db
	initDB()

	// init redis
	initRedis()

	// new server
	server := coral.NewServer(config.HOST)

	// new router
	initRouter(server)

	// start server
	server.Run()
}
