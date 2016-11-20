package filter

import (
	. "github.com/coral"
	//. "github.com/coral/cache"
	//. "github.com/coral/example/constant"
)

func Redis(context *Context) bool {
	context.Data = "using redis"
	return true
}

func Set(context *Context) bool {
	context.Data = "using redis"
	return true
}

func Get(context *Context) bool {
	context.Data = "using redis"
	return true
}
