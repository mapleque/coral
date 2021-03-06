package filter

import (
	. "github.com/coral"
	"github.com/coral/db"
	. "github.com/coral/example/constant"

	"time"
)

func Mysql(context *Context) bool {
	context.Data = "using db " + DEF_CORAL_DB
	return true
}
func Insert(context *Context) bool {
	ret := db.Insert(
		DEF_CORAL_DB,
		`INSERT INTO coral (name, type, status, flag, rate, additional, time)
		VALUES (?,?,?,?,?,?,?)`,
		"coral", "a", 1, true, 0.99, "中文", time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05"))
	context.Data = ret
	return true
}
func Update(context *Context) bool {
	ret := db.Update(
		DEF_CORAL_DB,
		"UPDATE coral SET status = ? WHERE name = ?",
		2, "coral")
	context.Data = ret
	return true
}
func Select(context *Context) bool {
	conn := db.UseDB(DEF_CORAL_DB)
	context.Data = conn.Select(
		"SELECT * FROM coral WHERE name = ?",
		"coral")
	return true
}
func TransCommit(context *Context) bool {
	trans := db.Begin(DEF_CORAL_DB)
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
	trans := db.Begin(DEF_CORAL_DB)
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
