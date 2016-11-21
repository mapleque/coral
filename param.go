package coral

import (
	"fmt"
	"strconv"

	. "github.com/coral/log"
)

type ParamCheckRule interface {
	Check(V) Filter

	Optioanl(func(interface{})) RF
	And(...func(interface{})) RF
	Or(...func(interface{})) RF

	IsAny(interface{}) bool
	IsBool(interface{}) bool
	IsInt(p interface{}) bool
	IsString(interface{}) bool

	Min(int) RF
	Max(int) RF
	MinLen(int) RF
	MaxLen(int) RF
}
type RuleFunc func(*Context, string) bool
type RF RuleFunc
type Validator map[string]RF
type V Validator
type R struct{}

func (r *R) Check(v V) Filter {
	return func(context *Context) bool {
		for k, f := range v {
			if !f(context, k) {
				context.Status = STATUS_INVALID_PARAM
				return false
			}
		}
		for k, _ := range context.Params {
			if v[k] == nil {
				return false
			}
		}
		return true
	}
}

func (r *R) Optioanl(f RF) RF {
	return func(context *Context, key string) bool {
		p := context.Params[key]
		if p != nil {
			return f(context, key)
		}
		return true
	}
}

func (r *R) And(f ...RF) RF {
	return func(context *Context, key string) bool {
		for _, sf := range f {
			if !sf(context, key) {
				return false
			}
		}
		return true
	}
}

func (r *R) Or(f ...RF) RF {
	return func(context *Context, key string) bool {
		for _, sf := range f {
			if sf(context, key) {
				return true
			}
		}
		return false
	}
}

func (r *R) IsAny(context *Context, key string) bool {
	return true
}

func (r *R) IsBool(context *Context, key string) bool {
	p := context.Params[key]
	switch p {
	case "1":
	case "true":
		context.Params[key] = true
		break
	case "0":
	case "false":
		context.Params[key] = false
		break
	default:
		Debug(fmt.Sprintf("param type should be bool but ", p))
		return false
	}
	return true
}

func (r *R) IsInt(context *Context, key string) bool {
	p := context.Params[key]
	np, err := strconv.Atoi(p.(string))
	if err != nil {
		Debug(fmt.Sprintf("param type should be int but %T", p))
		return false
	}
	context.Params[key] = np
	return true
}

func (r *R) IsString(context *Context, key string) bool {
	p := context.Params[key]
	switch p.(type) {
	case string:
		return true
	default:
		Debug(fmt.Sprintf("param type should be string but %T", p))
		return false
	}
}

func (r *R) Min(min int) RF {
	return func(context *Context, key string) bool {
		p := context.Params[key]
		l := p.(int)
		if l < min {
			Debug("param min should be ", min, " but ", l)
			return false
		}
		return true
	}
}

func (r *R) Max(max int) RF {
	return func(context *Context, key string) bool {
		p := context.Params[key]
		l := p.(int)
		if l > max {
			Debug("param max should be ", max, " but ", l)
			return false
		}
		return true
	}
}

func (r *R) MinLen(min int) RF {
	return func(context *Context, key string) bool {
		p := context.Params[key]
		l := len(p.(string))
		if l < min {
			Debug("param min len should be ", min, " but ", l)
			return false
		}
		return true
	}
}

func (r *R) MaxLen(max int) RF {
	return func(context *Context, key string) bool {
		p := context.Params[key]
		l := len(p.(string))
		if l > max {
			Debug("param max len should be ", max, " but ", l)
			return false
		}
		return true
	}
}
