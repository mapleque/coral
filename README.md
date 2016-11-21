# Abstract
本文主要阐述了coral框架的实现理念和使用方法。这一基于go实现的超轻量级框架更适合作为服务端基础数据服务使用，它抛弃了模板这一为了前端表现而存在的妥协产物（因为前端的事情完全可以交给更专业的javascript来完成），通过json协议包装request和response，从而使得框架更轻更纯粹，也更容易维护。
coral实现了路由和路由组的包装，实现了参数校验和过滤器链，另外还利用go的log模块和database模块实现了日志和数据库插件。
# Server & Router
每个server都可以指定一个监听端口提供web服务。
```
server := coral.NewServer(":8080")
```
server和router都提供了NewRouter方法，可以为每个指定的Path定义不同的处理策略。router同样也支持链式处理策略，这个在后面的部分将会被看到。
```
// curl http://localhost:8080/
// hello coral
baseRouter := server.NewRouter("/", api.Index)

// curl http://localhost:8080/param?<params>
// <params>
paramRouter := baseRouter.NewRouter("param", api.Param)
```
其中由router创建的router属于子路径，path将会自动加上父router的path。
最后启动server。
```
server.Run()
```
# Filter & Context
Filter是api接管请求，添加进一步逻辑处理的入口，对于每个Filter方法，都有一个Context对象作为参数。
当Filter返回false时，系统将不在处理后面的filter，直接给用户返回数据。
Context对象是请求处理过程中贯穿始终的上下文数据，用户在使用框架的任何filter中都可以对其中数据加工处理。
系统最终会将Context对象中的Status,Data,Errmsg返回给请求者。
```
package api

import (
	. "coral/common"
)

func Index(context *Context) {
	context.Data = "hello coral"
    return true
}

func Param(context *Context) bool {
	context.Data = context.Params
	return true
}
```
# Param
coral提供了基本参数校验方法以及一些列用于校验的方法。
```
r := &R{}

// curl http://localhost:8080/param/check?a=1&b=2&c=1
// {"a":"1", "b":2, "c":"1"}
paramRouter.NewRouter("check", r.Check(V{"a": r.IsString, "b": r.IsInt, "c": r.IsBool}), api.Param)
```
事实上，http包对于从请求获取的参数，都是string类型，因此在参数校验的方法里边，特意加入了强制类型转换逻辑，校验方法会根据用户所要求的类型尝试转换参数，如果成功就赋值给context.Params否则直接返回参数错误提示。这样一来，用户在自己的Filter中就可以直接使用期望的参数类型了。
这里的check方法，实际上返回的就是一个Filter，因此用户完全可以自己实现参数校验，就像Filter所干的事情一样。
值得注意的是，这里使用了前面提到的router对Filter的链式调用。
# Mysql
Mysql驱动选用了github.com/go-sql-driver/mysql，框架db包对其操作进行了封装，用户需要在启动server之前初始化并添加自己的DB，然后通过一个全局变量DB就可以调用对应sql方法进行数据库操作。
```
// 初始化db的方法，在启动server的时候调用一次即可
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
```
在router中添加mysql操作的路由
```
	// /mysql
	mysqlRouter := baseRouter.NewRouter("mysql", filter.Mysql)
	// /mysql/select
	mysqlRouter.NewRouter("select", filter.Select)
```
实现对应的filter
```
func Select(context *Context) bool {
	context.Data = DB.Select(
		DEF_DEFAULT_DB,
		"SELECT * FROM user WHERE id = ?",
		1)
	return true
}
TODO 批量操作mysql用prepare
TODO 事物
```
# Redis
Redis驱动选用了github.com/garyburd/redigo/redis，框架cache包对其进行了封装，用户需要再启动server之前初始化并添加自己的redis，然后通过全局变量Cache就可以调用Set或者Get进行操作。
```
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
```
添加redis路由
```
	// /redis
	redisRouter := baseRouter.NewRouter("redis", filter.Redis)
	redisRouter.NewRouter("set", filter.Set)
	redisRouter.NewRouter("get", filter.Get)
```
实现对应的filter
```
func Set(context *Context) bool {
	param := context.Params
	key := param["key"].(string)
	val := param["val"]
	ret := Cache.Set(DEFAULT_REDIS, key, val)
	context.Data = ret
	return true
}

func Get(context *Context) bool {
	param := context.Params
	key := param["key"].(string)
	context.Data = Cache.Get(DEFAULT_REDIS, key)
	return true
}
```
# Log
Log模块实现了日志分级输出，日志文件限制大小，自动循环切分等。
其用法与db模块类似，在启动server的时候初始化一次，在程序中使用全局变量Log或者全局方法Info等输出日志。
```
func initLog() {
	logPool := log.InitLog()
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
```
添加log的路由：
```
	// log
	baseRouter.NewRouter("log", filter.Log)
```
实现对应的filter
```
func Log(context *Context) bool {
	log.Debug(context.Params)
	return true
}
```
