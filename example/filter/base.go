package filter

import (
	. "github.com/coral"
)

func Index(context *Context) bool {
	context.Data = "hello coral"
	return true
}

func Param(context *Context) bool {
	context.Data = context.Params
	return true
}
