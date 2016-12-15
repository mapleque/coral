package filter

import (
	. "github.com/coral"
	"github.com/coral/cache"
	. "github.com/coral/example/constant"
)

func Redis(context *Context) bool {
	context.Data = "using redis"
	return true
}

func Set(context *Context) bool {
	param := context.Params
	key := param["key"].(string)
	val := param["val"]
	ret := cache.Set(DEF_CORAL_REDIS, key, val)
	context.Data = ret
	return true
}

func Get(context *Context) bool {
	param := context.Params
	key := param["key"].(string)
	var ret map[string]interface{}
	ret = make(map[string]interface{})
	ret[key] = cache.Get(DEF_CORAL_REDIS, key)
	context.Data = ret
	return true
}
