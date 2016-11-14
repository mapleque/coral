package common

import (
	"net/http"
)

// Filter 是一个接口过滤器
// 在创建router的时候传入的filterChains中的每一个元素都必须是这一类型
// context参数可以取到请求处理过程中的任何数据
type Filter func(context *Context)

// Context 是一个请求上下文数据结构定义
// 其中包含了请求处理过程中的所有数据
// 暴露给所有filter方法
// 请求的接收与返回也都通过其记录处理
type Context struct {
	req *http.Request
	w   http.ResponseWriter

	Response string
}

// Router 是一个路由数据结构定义
// 支持子路由
type Router struct {
	path    string
	routers []*Router

	handler func(http.ResponseWriter, *http.Request)
}

// 创建一个新的路由对象并返回引用
func NewRouter(path string, filterChains ...Filter) *Router {
	router := &Router{}
	router.path = path
	router.handler = genHandler(filterChains...)
	return router
}

// 添加一个子路由
func (router *Router) AddRoute(subRouter *Router) {
	router.routers = append(router.routers, subRouter)
}

// 生成路由处理函数
// 该方法将遍历filterChains中的所有方法，并使其在接收到请求后顺序执行
// 在所有filter执行结束后，返回了context中的response数据
func genHandler(filterChains ...Filter) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		context := &Context{}
		context.req = req
		context.w = w
		Info("<-", req.URL)
		for _, filter := range filterChains {
			filter(context)
		}
		w.Write([]byte(context.Response))
	}
}
