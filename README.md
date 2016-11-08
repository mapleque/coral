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
go是一种天生支持并发的语言，以至于通过net/http包实现web服务只需要简单几行代码。
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
# Path
TODO 定义在一个文件中，与参数一起作为服务文档，同时还要研究限制返回数据的方案。
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
