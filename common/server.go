package common

import (
	"net/http"
)

// Server是一个服务的对象定义，一个server对应一个端口监听
type Server struct {
	host    string
	mux     *http.ServeMux
	routers []*Router
}

// NewServer返回一个Server对象引用
func NewServer() *Server {
	Info("server start now ... ")
	server := &Server{}
	server.mux = http.NewServeMux()
	server.host = "0.0.0.0:8080"
	return server
}

// AddRouter为server添加一个router
func (server *Server) AddRoute(router *Router) {
	server.routers = append(server.routers, router)
}

// registerRouters 注册所有已添加的router
func (server *Server) registerRouters() {
	for _, router := range server.routers {
		server.registerRouter(router)
	}
}

// registerRouter 递归注册指定的一个router
func (server *Server) registerRouter(router *Router) {
	Info("register router " + router.path)
	server.mux.HandleFunc(router.path, router.handler)
	for _, child := range router.routers {
		server.registerRouter(child)
	}
}

// Run启动server的服务
func (server *Server) Run() {
	server.registerRouters()
	Info("server listening on " + server.host)
	err := http.ListenAndServe(server.host, server.mux)
	if err != nil {
		Error(err)
		Error("server start FAILD!")
	}
}
