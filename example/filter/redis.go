package filter

import (
	. "github.com/coral"
	. "github.com/coral/cache"
	. "github.com/coral/example/config"
)

func Redis(context *Context) bool {
	context.Data = "using redis"
	return true
}

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
