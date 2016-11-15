package api

import (
	. "coral/common"
)

func Index(context *Context) bool {
	context.Data = "hello coral"
	return true
}

func Param(context *Context) bool {
	context.Data = context.Params
	return true
}
