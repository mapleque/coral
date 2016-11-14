package api

import (
	. "coral/common"
)

func Index(context *Context) {
	context.Response = "hello coral"
}
