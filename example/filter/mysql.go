package filter

import (
	. "github.com/coral"
	. "github.com/coral/db"

	. "github.com/coral/example/config"
)

func Mysql(context *Context) bool {
	context.Data = "using db " + DEFAULT_DB
	return true
}
func Insert(context *Context) bool {
	ret := DB.Insert(
		DEFAULT_DB,
		"INSERT INTO user (username,password) VALUSES (?,?)",
		"aaa", "aaa")
	context.Data = ret
	return true
}
func Update(context *Context) bool {
	ret := DB.Update(
		DEFAULT_DB,
		"UPDATE user SET username = ? WHERE id = ?",
		"bbb", 1)
	context.Data = ret
	return true
}
func Select(context *Context) bool {
	context.Data = DB.Select(
		DEFAULT_DB,
		"SELECT * FROM user WHERE id = ?",
		1)
	return true
}
