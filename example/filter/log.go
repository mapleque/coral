package filter

import (
	. "github.com/coral"
	log "github.com/coral/log"
)

func Log(context *Context) bool {
	log.Debug(context.Params)
	return true
}
