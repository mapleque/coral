package main

import (
	"flag"

	coral "github.com/coral"
	cache "github.com/coral/cache"
	config "github.com/coral/config"
	db "github.com/coral/db"
	log "github.com/coral/log"

	. "github.com/coral/example/constant"
	filter "github.com/coral/example/filter"
)

var conf config.Configer

/**
 * initRouter 方法，定义服务全部接口、参数校验、并指定过滤器链
 */
func initRouter(server *coral.Server) {
	// /
	baseRouter := server.NewRouter("/", filter.Index)

	// /param?<params>
	paramRouter := baseRouter.NewRouter("param", filter.Param)

	// doc & checker
	// /doc-example?a=aa&b={"c":1}&data={"list":[{"e":"2"},{"e":"0"}],"pages":[0,2,3]}
	doc := &coral.Doc{
		Path:        "doc-example",
		Description: "a example api",
		Input: coral.Checker{
			"a": "string(2)",
			"b": coral.Checker{
				"c": "int[1,10]"},
			"data": coral.Checker{
				"list": []coral.Checker{
					coral.Checker{"e": "string"}},
				"pages": []string{"int"}}},
		Output: coral.Checker{
			"status": "int",
			"data": coral.Checker{
				"a": "string(2)",
				"b": coral.Checker{
					"c": "int[1,10]"},
				"data": coral.Checker{
					"list": []coral.Checker{
						coral.Checker{"e": "string"}},
					"pages": []string{"int"}}},
			"errmsg": "string"}}
	baseRouter.NewDocRouter(doc, filter.Param)

	// log
	baseRouter.NewRouter("log", filter.Log)

	// /mysql
	mysqlRouter := baseRouter.NewRouter("mysql", filter.Mysql)
	// /mysql/*
	mysqlRouter.NewRouter("select", filter.Select)
	mysqlRouter.NewRouter("insert", filter.Insert)
	mysqlRouter.NewRouter("update", filter.Update)
	mysqlRouter.NewRouter("transCommit", filter.TransCommit)
	mysqlRouter.NewRouter("transRollback", filter.TransRollback)

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
	// add default db
	db.DB.AddDB(
		DEF_CORAL_DB,
		conf.Get("db.DEFAULT_DB_DSN"),
		conf.Int("db.DEFAULT_DB_MAX_CONNECTION"),
		conf.Int("db.DEFAULT_DB_MAX_IDLE"))

	// add other db
	// ...
}

func initRedis() {
	// add default cache
	cache.Cache.AddRedis(
		DEF_CORAL_REDIS,
		conf.Get("cache.DEFAULT_REDIS_SERVER"),
		conf.Get("cache.DEFAULT_REDIS_AUTH"),
		conf.Int("cache.DEFAULT_REDIS_MAX_CONNECTION"),
		conf.Int("cache.DEFAULT_REDIS_MAX_IDLE"))

	// add other cache
	// ...
}

func initLog() {
	// add default logger
	log.Log.AddLogger(
		DEF_CORAL_LOG,
		conf.Get("log.DEFAULT_LOG_PATH"),
		conf.Int("log.DEFAULT_LOG_MAX_NUMBER"),
		conf.Int64("log.DEFAULT_LOG_MAX_SIZE"),
		conf.Int("log.DEFAULT_LOG_MAX_LEVEL"),
		conf.Int("log.DEFAULT_LOG_MIN_LEVEL"))

	// add other logger
	// ...
}

func main() {
	confFile := flag.String("ini", "", "your config file")
	flag.Parse()
	if *confFile != "" {
		config.AddConfiger(config.INI, DEF_CORAL_CONF, *confFile)
		conf = config.Use(DEF_CORAL_CONF)

		// init log
		initLog()

		// init db
		initDB()

		// init redis
		initRedis()

		// new server
		server := coral.NewServer(conf.Get("server.HOST"))

		// new router
		initRouter(server)

		// start server
		server.Run()
	} else {
		panic("run with -h to find usage")
	}
}
