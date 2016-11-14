# Abstract
本文主要阐述了一个go web框架的实现理念和使用方法。这一框架更适合作为服务端基础数据服务，通过json协议包装request和response，抛弃了模板这一为了前端表现而存在的妥协产物（因为前端的事情完全可以交给更专业的javascript来完成），从而使得框架更轻更纯粹，也更容易维护。
所谓web服务，path和param是入口，所以实现路由策略和参数校验功能是web框架必不可少的。而数据持久化是数据服务的核心，为了使框架简洁，本文仅以最常用的mysql数据库作为数据持久化实现方法示例，如果有不同的需要，读者完全可以在此基础上实现个人的数据持久化方法。
对于代码编写者而言，良好的组织架构总是能产生事半功倍的效果，本文介绍了如何以MVC的组织方式实现一个服务。此外，本文后半部分还涉及了生产环境部署相关问题。
# Introduce
## go web的优势
## go web的劣势
## go web的现状
## go web的目标
# Base
go的http包支持快速搭建一个web server。
```
package main

import (
    "net/http"
)

func sayHello(w http.ResponseWriter, req *http.Request) {
    w.Write([]byte("Hello"))
}

func main() {
    http.HandleFunc("/hello", sayHello)
    http.ListenAndServe(":80", nil)
}
```
```
go build server.go
./server
curl http://localhost/hello
> Hello
```
通过上面的实践可以发现，go通过监听指定端口，接管了所有请求数据，不需要额外的http服务器以及web容器的支持。
# Server & Router
每个server都可以指定一个监听端口提供web服务。
```
server := common.NewServer()
```
router可以为每个指定的Path定义不同的处理策略。
```
router := common.NewRouter("/", api.Index)
```
最后需要将router添加到server中，并启动server
```
server.AddRoute(router)
server.Run()
```
# Filter & Context
Filter是api接管请求，添加进一步逻辑处理的入口，对于每个Filter方法，都有一个Context对象作为参数。
Context对象是请求处理过程中贯穿始终的上下文数据，用户在使用框架的任何filter中都可以对其中数据加工处理。
框架最终会将Context对象中的response返回给请求者。
```
package api

import (
	. "coral/common"
)

func Index(context *Context) {
	context.Response = "hello coral"
}
```
# Param
TODO 主要是参数校验，支持各种类型参数校验。
# Mysql
TODO 主要考虑在不使用orm的前提下如何防止sql注入
TODO 研究一种支持mysql prepare的第三方库
# AOP
TODO 通过例子展示一种AOP形式的代码组织方案
# Log
TODO 考虑日志分级输出可配置、输出路径可配置、大小限制可切分等
# Online
TODO 考虑如何保证服务性能、稳定性、上线流程等
