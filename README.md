# Abstract
本文主要阐述了coral框架的实现理念和使用方法。这一基于go实现的超轻量级框架更适合作为服务端基础数据服务使用，它抛弃了模板这一为了前端表现而存在的妥协产物（因为前端的事情完全可以交给更专业的javascript来完成），通过json协议包装request和response，从而使得框架更轻更纯粹，也更容易维护。
coral实现了路由和路由组的包装，实现了参数校验和过滤器链，另外还利用go的log模块和database模块实现了日志和数据库插件。
# Server & Router
每个server都可以指定一个监听端口提供web服务。
```
server := coral.NewServer(":8080")
```
server和router都提供了NewRouter方法，可以为每个指定的Path定义不同的处理策略。router可以指定多个filter链式处理。
```
baseRouter := server.NewRouter("/", api.Index)

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
package filter

import (
	. "github.com/coral"
)

func Index(context *Context) bool {
	context.Data = "Hello coral"
	context.Raw = true
	return true
}

func Param(context *Context) bool {
	context.Data = context.Params
	return true
}

func ParamGet(context *Context) bool {
	ret := make(map[string]interface{})
	ret["intVal"] = Int(context.Params["int"])
	ret["strVal"] = String(context.Params["string"])

	data := Map(context.Params["data"])

	arrInt := Array(data["array"])
	for i, arrIntVal := range arrInt {
		ret["arrInt"+String(i)] = Int(arrIntVal)
	}
	arrEle := Array(data["list"])
	for i, arrEleVal := range arrEle {
		arrEleValMap := Map(arrEleVal)
		ret["arrEleVal"+String(i)] = arrEleValMap["ele"]
	}
	context.Data = ret
	return true
}
```
在filter中从contex.Params中提取参数值做进一步操作时，通常需要指定类型，coral实现了强制类型转换的方法，在上面代码的ParamGet方法中，可以看到这些方法的使用方式。这个filter的路由定义在下面的代码中。
# Doc
coral支持通过预定义的doc信息，生成api doc，同时也会根据doc校验输入和输出。
```
// /doc-example?a=aa&b={"c":1}&data={"list":[{"e":"2"},{"e":"0"}],"pages":[0,2,3]}
	doc := &coral.Doc{
		Path:        "doc-example",
		Description: "a example api",
		Input: coral.Checker{
			"a": coral.Rule("string(2)", STATUS_INVALID_INPUT, "字符串"),
			"b": coral.Checker{
				"c": coral.Rule(
					"int[1,10]",
					STATUS_INVALID_INPUT,
					"1-10的int")},
			"data": coral.Checker{
				"list": []coral.Checker{
					coral.Checker{
						"e": coral.Rule(
							"string",
							STATUS_INVALID_INPUT,
							"数组里每个元素都是这样的对象")}},
				"pages": []string{coral.Rule(
					"int",
					STATUS_INVALID_INPUT,
					"素组每个元素都是int")}}},
		Output: coral.Checker{
			"status": coral.Rule(
				"int",
				STATUS_INVALID_OUTPUT,
				"对应的说明"),
			"data": coral.Checker{
				"a": "string(2)", // 也可以直接写
				"b": coral.Checker{
					"c": coral.Rule("int[1,10]", 0, "")}, // 也可以省略说明
				"data": coral.Checker{
					"list": []coral.Checker{
						coral.Checker{"e": "string"}},
					"pages": []string{"int"}}},
			"errmsg": "string"}}
	baseRouter.NewDocRouter(doc, filter.Param)

	// for param get
	baseRouter.NewDocRouter(&coral.Doc{
		Path:        "param-get",
		Description: "取param示例",
		Input: coral.Checker{
			"int":    "int",
			"string": "string",
			"data": coral.Checker{
				"array": []string{"int"},
				"list": []coral.Checker{
					coral.Checker{
						"ele": "string"}}}}},
		filter.ParamGet)
```
当server运行时，访问/doc可以看到全部路由doc，也可以点击对应的doc节点查看子路由的doc。从上面路由定义的代码中，还可以看到当需要传递的参数较为复杂时，使用data包装json的形式更为妥当。
# Config
coral支持配置文件读入，目前实现了ini文件的读取。
```
var conf config.Configer
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
```
这里通过flag传入文件路径，初始化了一个configer，所以启动server的命令需要加上对应的参数。
```
go run index.go --ini config/config.ini
```
# Mysql
Mysql驱动选用了github.com/go-sql-driver/mysql，框架db包对其操作进行了封装，用户需要在启动server之前初始化并添加自己的DB，然后通过一个全局变量DB就可以调用对应sql方法进行数据库操作。
```
func initDB() {
	// add default db
	db.DB.AddDB(
		DEF_CORAL_DB,
		conf.Get("db.DEF_CORAL_DB_DSN"),
		conf.Int("db.DEF_CORAL_DB_MAX_CONNECTION"),
		conf.Int("db.DEF_CORAL_DB_MAX_IDLE"))

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
// 插入数据，可以直接使用DBPool对象操作数据库
func Insert(context *Context) bool {
	ret := DB.Insert(
		DEF_CORAL_DB,
		`INSERT INTO coral (name, type, status, flag, rate, additional, time)
		VALUES (?,?,?,?,?,?,?)`,
		"coral", "a", 1, true, 0.99, "中文", time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05"))
	context.Data = ret
	return true
}
// 查询数据，也可以用DBquery对象操作数据库
func Select(context *Context) bool {
	conn := DB.UseDB(DEF_CORAL_DB)
	context.Data = conn.Select(
		"SELECT * FROM coral WHERE name = ?",
		"coral")
	return true
}
// 事物
func TransCommit(context *Context) bool {
	trans := DB.Begin(DEF_CORAL_DB)
	ret := trans.Update(
		"UPDATE coral SET status = ? WHERE name = ?",
		1, "coral")
	if ret < 1 {
		context.Data = "update faild rollback"
		trans.Rollback()
		return false
	}
	trans.Commit()
	return Select(context)
}
```
# Redis
Redis驱动选用了github.com/garyburd/redigo/redis，框架cache包对其进行了封装，用户需要再启动server之前初始化并添加自己的redis，然后通过全局变量Cache就可以调用Set或者Get进行操作。
```
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
	ret := Cache.Set(DEF_CORAL_REDIS, key, val)
	context.Data = ret
	return true
}

func Get(context *Context) bool {
	param := context.Params
	key := param["key"].(string)
	context.Data = Cache.Get(DEF_CORAL_REDIS, key)
	return true
}
```
# Log
Log模块实现了日志分级输出，日志文件限制大小，自动循环切分等。
其用法与db模块类似，在启动server的时候初始化一次，在程序中使用全局变量Log或者全局方法Info等输出日志。
```
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
