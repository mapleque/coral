package main

import (
	api "coral/api"  // an api package
	. "coral/common" // common package
)

func initRouter(server *Server) {
	// curl http://localhost:8080/
	// hello coral
	router := server.NewRouter("/", api.Index)
	r := &R{}

	// curl http://localhost:8080/param?<params>
	// <params>
	param := router.NewRouter("param", api.Param)
	param.NewRouter("check", r.Check(V{"a": r.IsString, "b": r.IsInt, "c": r.IsBool}), api.Param)
}

func main() {
	// new server
	server := NewServer()
	// new router
	initRouter(server)
	// start server
	server.Run()
}
