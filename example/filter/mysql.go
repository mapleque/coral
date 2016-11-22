package filter

import (
	. "github.com/coral"
	. "github.com/coral/db"

	. "github.com/coral/example/config"

	"time"
)

func Mysql(context *Context) bool {
	context.Data = "using db " + DEFAULT_DB
	return true
}
func Insert(context *Context) bool {
	ret := DB.Insert(
		DEFAULT_DB,
		`INSERT INTO coral (name, type, status, flag, rate, additional, time)
		VALUES (?,?,?,?,?,?,?)`,
		"coral", "a", 1, true, 0.99, "中文", time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05"))
	context.Data = ret
	return true
}
func Update(context *Context) bool {
	ret := DB.Update(
		DEFAULT_DB,
		"UPDATE coral SET status = ? WHERE name = ?",
		2, "coral")
	context.Data = ret
	return true
}
func Select(context *Context) bool {
	conn := DB.UseDB(DEFAULT_DB)
	context.Data = conn.Select(
		"SELECT * FROM coral WHERE name = ?",
		"coral")
	return true
}
func TransCommit(context *Context) bool {
	trans := DB.Begin(DEFAULT_DB)
	ret := trans.Update(
		"UPDATE coral SET status = ? WHERE name = ?",
		1, "coral")
	if ret < 1 {
		context.Errmsg = "update faild rollback"
		trans.Rollback()
		return false
	}
	trans.Commit()
	return Select(context)
}

func TransRollback(context *Context) bool {
	trans := DB.Begin(DEFAULT_DB)
	ret := trans.Update(
		"UPDATE coral SET status = ? WHERE name = ?",
		2, "coral")
	if ret < 1 {
		context.Data = ret
		return false
	}
	trans.Rollback()
	return Select(context)
}
