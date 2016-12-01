package coral

import (
	"encoding/json"
	"net/http"
	"time"

	. "github.com/coral/log"
)

// Server是一个服务的对象定义，一个server对应一个端口监听
type Server struct {
	host    string
	mux     *http.ServeMux
	routers []*Router
}

// Router 是一个路由数据结构定义
// 支持子路由
type Router struct {
	path    string
	docPath string
	routers []*Router

	doc *Doc

	handler    func(http.ResponseWriter, *http.Request)
	docHandler func(http.ResponseWriter, *http.Request)
}

type Doc struct {
	description string
	path        string
	docPath     string
	input       []*DocField
	output      []*DocField
}

type DocField struct {
	name  string
	rule  string
	extra *DocField
}

// Filter 是一个接口过滤器
// 在创建router的时候传入的filterChains中的每一个元素都必须是这一类型
// context参数可以取到请求处理过程中的任何数据
type Filter func(context *Context) bool

// Context 是一个请求上下文数据结构定义
// 其中包含了请求处理过程中的所有数据
// 暴露给所有filter方法
// 请求的接收与返回也都通过其记录处理
type Context struct {
	req *http.Request
	w   http.ResponseWriter

	Host   string
	Path   string
	Params map[string]interface{}
	Data   interface{}
	Status int
	Errmsg string

	Raw bool
}

// Response 是请求返回数据类型
type Response struct {
	Status int         `json:"status"`
	Data   interface{} `json:"data"`
	Errmsg string      `json:"errmsg"`
}

// NewServer返回一个Server对象引用
func NewServer(host string) *Server {
	Info("coral start now ...")
	server := &Server{}
	server.mux = http.NewServeMux()
	server.host = host
	return server
}

// AddRouter为server添加一个router
func (server *Server) AddRoute(router *Router) {
	server.routers = append(server.routers, router)
}

// Run启动server的服务
func (server *Server) Run() {
	server.registerRouters()
	Info("coral listening on", server.host)
	Info("========================================")
	err := http.ListenAndServe(server.host, server.mux)
	if err != nil {
		Error(err)
		Error("server start FAILD!")
	}
}

// 创建一个新的路由对象并返回引用
func (server *Server) NewRouter(path string, filterChains ...Filter) *Router {
	// path head must be "/"
	if len(path) < 1 || path[0] != '/' {
		path = "/" + path
	}
	router := newRouter(path, filterChains...)
	server.AddRoute(router)
	return router
}

// 创建一个带有doc的路由对象
func (server *Server) NewDocRouter(doc *Doc, filterChains ...Filter) *Router {
	// path head must be "/"
	if len(doc.path) < 1 || doc.path[0] != '/' {
		doc.path = "/" + doc.path
	}
	router := newDocRouter(doc, filterChains...)
	server.AddRoute(router)
	return router
}

// registerRouters 注册所有已添加的router
func (server *Server) registerRouters() {
	for _, router := range server.routers {
		server.registerRouter(router)
		server.registerDocRouter(router)
	}
}

// registerRouter 递归注册指定的一个router
func (server *Server) registerRouter(router *Router) {
	Info("register router", router.path)
	server.mux.HandleFunc(router.path, router.handler)
	for _, child := range router.routers {
		server.registerRouter(child)
	}
}

// registerDocRouter 递归注册指定router的doc
func (server *Server) registerDocRouter(router *Router) {
	Info("register router doc", router.docPath)
	server.mux.HandleFunc(router.docPath, router.docHandler)
	for _, child := range router.routers {
		server.registerDocRouter(child)
	}
}

// 添加一个子路由
func (router *Router) NewRouter(path string, filterChains ...Filter) *Router {
	// path head must be "/"
	if len(path) < 1 {
		Error("Empty router path register on", router.path)
	}
	if path[0] != '/' && router.path[len(router.path)-1] != '/' {
		path = "/" + path
	}
	subRouter := newRouter(router.path+path, filterChains...)
	router.routers = append(router.routers, subRouter)
	return subRouter
}

// 添加一个带doc的子路由
func (router *Router) NewDocRouter(doc *Doc, filterChains ...Filter) *Router {
	// path head must be "/"
	if len(doc.path) < 1 {
		Error("Empty router path register on", router.path)
	}
	if doc.path[0] != '/' && router.path[len(router.path)-1] != '/' {
		doc.path = "/" + doc.path
	}
	doc.path = router.path + doc.path
	subRouter := newDocRouter(doc, filterChains...)
	router.routers = append(router.routers, subRouter)
	return subRouter
}

// 生成路由处理函数
// 该方法将遍历filterChains中的所有方法，并使其在接收到请求后顺序执行
// 在所有filter执行结束后，返回了context中的response数据
func (router *Router) genHandler(filterChains ...Filter) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		startTime := time.Now()
		context := &Context{}
		context.req = req
		context.w = w
		context.Host = req.Host
		context.Path = router.path
		// deal params
		context.Params = router.processParams(req)

		response := &Response{}
		for _, filter := range filterChains {
			if !filter(context) {
				Warn("filter break")
				if context.Status == 0 {
					response.Status = STATUS_ERROR_UNKNOWN
				}
				if context.Errmsg == "" {
					response.Errmsg = "filter return false"
				}
				break
			}
		}
		if context.Raw {
			Info(
				"<-",
				time.Now().Sub(startTime),
				context.Host,
				context.Path,
				context.Params,
				"->",
				context.Data)
			w.Write([]byte(context.Data.(string)))
		} else {
			if context.Status != 0 {
				response.Status = context.Status
			}
			response.Data = context.Data
			if context.Errmsg != "" {
				response.Errmsg = context.Errmsg
			}
			out, err := json.Marshal(response)
			if err != nil {
				Error(err)
			}
			Info(
				"<-",
				time.Now().Sub(startTime),
				context.Host,
				context.Path,
				context.Params,
				"->",
				context.Status,
				context.Data,
				context.Errmsg)
			w.Write(out)
		}
	}
}

func (router *Router) processParams(req *http.Request) map[string]interface{} {
	err := req.ParseForm()
	if err != nil {
		Error(err)
	}
	params := make(map[string]interface{})
	for k, vs := range req.Form {
		if len(vs) > 0 {
			params[k] = vs[0]
		} else {
			params[k] = ""
		}
	}
	return params
}

func newRouter(path string, filterChains ...Filter) *Router {
	router := &Router{}
	router.path = path
	docPath := "doc"
	if path[len(path)-1] != '/' {
		docPath = "/" + docPath
	}
	router.docPath = path + docPath

	doc := &Doc{}
	doc.description = "no description in simple router"
	doc.path = router.path
	doc.docPath = router.docPath
	router.doc = doc

	router.handler = router.genHandler(filterChains...)
	router.docHandler = router.genDocHandler()
	return router
}

func newDocRouter(doc *Doc, filterChains ...Filter) *Router {
	router := &Router{}
	router.path = doc.path
	docPath := "doc"
	if doc.path[len(doc.path)-1] != '/' {
		docPath = "/" + docPath
	}
	router.doc = doc
	router.docPath = doc.path + docPath
	router.handler = router.genHandler(filterChains...)
	router.docHandler = router.genDocHandler()
	return router
}

// genDocHandler write a doc view
func (router *Router) genDocHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		docs := genDoc(router)
		ret := ""
		for _, doc := range docs {
			ret = ret + doc.genView()
		}
		ret = "<!doctype html>" +
			"<title>api doc - general by coral</title>" +
			"<h1>Api doc - general by coral</h1>" +
			ret
		w.Write([]byte(ret))
	}
}

func genDoc(router *Router) []*Doc {
	var ret []*Doc
	ret = append(ret, router.doc)
	for _, child := range router.routers {
		docs := genDoc(child)
		for _, d := range docs {
			ret = append(ret, d)
		}
	}
	return ret
}

func (doc *Doc) genView() string {
	if doc == nil {
		return ""
	}
	ret := "<hr>" +
		"<p>" +
		"<a href='" + doc.docPath +
		"' title='click to see sub tree'>@path:</a> " + doc.path +
		"</p>"
	ret = ret + "<p>" + doc.description + "</p>"
	ret = ret + "<p><- input</p>"
	for _, field := range doc.input {
		ret = ret + field.genView()
	}
	ret = ret + "<h4>-> output</h4>"
	for _, field := range doc.output {
		ret = ret + field.genView()
	}
	return ret
}

func (field *DocField) genView() string {
	if field == nil {
		return ""
	}
	return "<li>" + "<p>" + field.name + " " + field.rule + "</p>" +
		"<ul>" + field.extra.genView() + "</ul>" + "</li>"
}
