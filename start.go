package main

import (
	api "coral/api"       // an api package
	common "coral/common" // common package
)

func main() {
	// new server
	server := common.NewServer()
	// new router
	router := common.NewRouter("/", api.Index)
	// add router to server
	server.AddRoute(router)
	// start server
	server.Run()
}
