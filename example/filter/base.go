package filter

import (
	. "github.com/coral"
	. "github.com/coral/db"
	. "github.com/coral/example/constant" // constant package
)

func Index(context *Context) bool {
	context.Data = "hello coral"
	return true
}

func Param(context *Context) bool {
	context.Data = context.Params
	return true
}

func Mysql(context *Context) bool {
	context.Data = "using db " + DEF_DEFAULT_DB
	return true
}
func Insert(context *Context) bool {
	ret := DB.Insert(
		DEF_DEFAULT_DB,
		"INSERT INTO user (username,password) VALUSES (?,?)",
		"aaa", "aaa")
	context.Data = ret
	return true
}
func Update(context *Context) bool {
	ret := DB.Update(
		DEF_DEFAULT_DB,
		"UPDATE user SET username = ? WHERE id = ?",
		"bbb", 1)
	context.Data = ret
	return true
}
func Select(context *Context) bool {
	context.Data = DB.Select(
		DEF_DEFAULT_DB,
		"SELECT * FROM user WHERE id = ?",
		1)
	return true
}
