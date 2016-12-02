package filter

import (
	. "github.com/coral"
)

func Index(context *Context) bool {
	context.Data = "Hello coral"
	context.Raw = true
	return true
}

func Param(context *Context) bool {
	context.Data = context.Params
	return true
}

func ParamGet(context *Context) bool {
	ret := make(map[string]interface{})
	ret["intVal"] = Int(context.Params["int"])
	ret["strVal"] = String(context.Params["string"])

	data := Map(context.Params["data"])

	arrInt := Array(data["array"])
	for i, arrIntVal := range arrInt {
		ret["arrInt"+String(i)] = Int(arrIntVal)
	}
	arrEle := Array(data["list"])
	for i, arrEleVal := range arrEle {
		arrEleValMap := Map(arrEleVal)
		ret["arrEleVal"+String(i)] = arrEleValMap["ele"]
	}
	context.Data = ret
	return true
}
