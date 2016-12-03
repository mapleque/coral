package coral

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	. "github.com/coral/log"
)

// Server是一个服务的对象定义，一个server对应一个端口监听
type Server struct {
	host    string
	mux     *http.ServeMux
	routers []*Router
}

// Router 是一个路由数据结构定义
// 支持子路由
type Router struct {
	path    string
	docPath string
	routers []*Router

	doc *Doc

	handler    func(http.ResponseWriter, *http.Request)
	docHandler func(http.ResponseWriter, *http.Request)
}

// Doc 用于生成api doc
// 在创建router的时候可以作为参数传入
type Doc struct {
	Path        string
	Description string
	docPath     string
	Input       Checker
	Output      Checker
}

type Checker map[string]interface{}

// Filter 是一个接口过滤器
// 在创建router的时候传入的filterChains中的每一个元素都必须是这一类型
// context参数可以取到请求处理过程中的任何数据
type Filter func(context *Context) bool

// Context 是一个请求上下文数据结构定义
// 其中包含了请求处理过程中的所有数据
// 暴露给所有filter方法
// 请求的接收与返回也都通过其记录处理
type Context struct {
	req *http.Request
	w   http.ResponseWriter

	Host   string
	Path   string
	Params map[string]interface{}
	Data   interface{}
	Status int
	Errmsg string

	Raw bool
}

// Response 是请求返回数据类型
type Response struct {
	Status int         `json:"status"`
	Data   interface{} `json:"data"`
	Errmsg string      `json:"errmsg"`
}

// rule: TYPE + (n)|[m,n]|{a,b,c...} + #STATUS_* + ;NOTE
func Rule(rule string, status int, note string) string {
	ret := rule
	if status > 0 {
		ret = ret + "#" + strconv.Itoa(status)
	}
	if note != "" {
		ret = ret + "<" + note + ">"
	}
	return ret
}

func InStatus(status ...int) string {
	status = append(status, STATUS_INVALID_PARAM)
	status = append(status, STATUS_ERROR_DB)
	status = append(status, STATUS_ERROR_UNKNOWN)
	status = append(status, STATUS_SUCCESS)
	var arr []string
	for _, st := range status {
		arr = append(arr, strconv.Itoa(st))
	}
	ret := strings.Join(arr, ",")
	return Rule("int{"+ret+"}", STATUS_INVALID_STATUS, "")
}

// 类型转换，任何类型转成int
func Int(param interface{}) int {
	switch ret := param.(type) {
	case int:
		return ret
	case int64:
		return int(ret)
	case float64:
		return int(ret)
	case string:
		r, err := strconv.Atoi(ret)
		if err != nil {
			Error("param type change error", ret, err.Error())
		}
		return r
	case bool:
		if ret {
			return 1
		} else {
			return 0
		}
	default:
		Error("param type change to int error",
			ret, fmt.Sprintf("%T", ret))
		return 0
	}
}

// 类型转换，任何类型转成bool
func Bool(param interface{}) bool {
	switch ret := param.(type) {
	case bool:
		return ret
	case int:
		if ret > 0 {
			return true
		} else {
			return false
		}
	case string:
		switch ret {
		case "1", "true", "y", "on", "yes":
			return true
		case "0", "false", "n", "off", "no":
			return false
		default:
			Error("param type change to bool error", ret, "unknown type")
		}
		return false
	default:
		Error("param type change to bool error",
			ret, fmt.Sprintf("%T", ret))
		return false
	}
}

// 类型转换，任何类型转成string
func String(param interface{}) string {
	switch ret := param.(type) {
	case string:
		return ret
	case int:
		return strconv.Itoa(ret)
	case bool:
		if ret {
			return "1"
		} else {
			return "0"
		}
	default:
		Error("param type change to string error",
			ret, fmt.Sprintf("%T", ret))
		return ""
	}
}

func Map(param interface{}) map[string]interface{} {
	switch ret := param.(type) {
	case map[string]interface{}:
		return ret
	default:
		Error("param type change to map error",
			ret, fmt.Sprintf("%T", ret))
		return nil
	}
}

func Array(param interface{}) []interface{} {
	switch ret := param.(type) {
	case []interface{}:
		return ret
	default:
		Error("param type change to map error",
			ret, fmt.Sprintf("%T", ret))
		return nil
	}
}

// NewServer返回一个Server对象引用
func NewServer(host string) *Server {
	Info("coral start now ...")
	server := &Server{}
	server.mux = http.NewServeMux()
	server.host = host
	return server
}

// AddRouter为server添加一个router
func (server *Server) AddRoute(router *Router) {
	server.routers = append(server.routers, router)
}

// Run启动server的服务
func (server *Server) Run() {
	server.registerRouters()
	Info("coral listening on", server.host)
	Info("========================================")
	err := http.ListenAndServe(server.host, server.mux)
	if err != nil {
		Error(err)
		Error("server start FAILD!")
	}
}

// 创建一个新的路由对象并返回引用
func (server *Server) NewRouter(path string, filterChains ...Filter) *Router {
	// path head must be "/"
	if len(path) < 1 || path[0] != '/' {
		path = "/" + path
	}
	router := newRouter(path, filterChains...)
	server.AddRoute(router)
	return router
}

// 创建一个带有doc的路由对象
func (server *Server) NewDocRouter(doc *Doc, filterChains ...Filter) *Router {
	// path head must be "/"
	if len(doc.Path) < 1 || doc.Path[0] != '/' {
		doc.Path = "/" + doc.Path
	}
	router := newDocRouter(doc, filterChains...)
	server.AddRoute(router)
	return router
}

// registerRouters 注册所有已添加的router
func (server *Server) registerRouters() {
	for _, router := range server.routers {
		server.registerRouter(router)
		server.registerDocRouter(router)
	}
}

// registerRouter 递归注册指定的一个router
func (server *Server) registerRouter(router *Router) {
	Info("register router", router.path)
	server.mux.HandleFunc(router.path, router.handler)
	for _, child := range router.routers {
		server.registerRouter(child)
	}
}

// registerDocRouter 递归注册指定router的doc
func (server *Server) registerDocRouter(router *Router) {
	Info("register router doc", router.docPath)
	server.mux.HandleFunc(router.docPath, router.docHandler)
	for _, child := range router.routers {
		server.registerDocRouter(child)
	}
}

// 添加一个子路由
func (router *Router) NewRouter(path string, filterChains ...Filter) *Router {
	// path head must be "/"
	if len(path) < 1 {
		Error("Empty router path register on", router.path)
	}
	if path[0] != '/' && router.path[len(router.path)-1] != '/' {
		path = "/" + path
	}
	subRouter := newRouter(router.path+path, filterChains...)
	router.routers = append(router.routers, subRouter)
	return subRouter
}

// 添加一个带doc的子路由
func (router *Router) NewDocRouter(doc *Doc, filterChains ...Filter) *Router {
	// path head must be "/"
	if len(doc.Path) < 1 {
		Error("Empty router path register on", router.path)
	}
	if doc.Path[0] != '/' && router.path[len(router.path)-1] != '/' {
		doc.Path = "/" + doc.Path
	}
	doc.Path = router.path + doc.Path
	subRouter := newDocRouter(doc, filterChains...)
	router.routers = append(router.routers, subRouter)
	return subRouter
}

// 生成路由处理函数
// 该方法将遍历filterChains中的所有方法，并使其在接收到请求后顺序执行
// 在所有filter执行结束后，返回了context中的response数据
func (router *Router) genHandler(filterChains ...Filter) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		startTime := time.Now()
		context := &Context{}
		context.req = req
		context.w = w
		context.Host = req.Host
		context.Path = router.path
		// deal params
		context.Params = router.processParams(req)

		ret := true
		response := &Response{}

		// param check if need
		if router.doc.Input != nil {
			ret, context.Status = router.doc.Input.check(context.Params)
			if !ret {
				Debug("input check faild")
			}
		}

		if ret {
			for _, filter := range filterChains {
				ret = filter(context)
				if !ret {
					Warn("filter break", filter)
					if context.Status == 0 {
						response.Status = STATUS_ERROR_UNKNOWN
					}
					if context.Errmsg == "" {
						response.Errmsg = "filter return false"
					}
					break
				}
			}
		}

		if context.Raw {
			Info(
				"<-",
				time.Now().Sub(startTime),
				context.Host,
				context.Path,
				context.Params,
				"->",
				context.Data)
			w.Write([]byte(context.Data.(string)))
		} else {
			if context.Status != 0 {
				response.Status = context.Status
			}
			response.Data = context.Data
			if context.Errmsg != "" {
				response.Errmsg = context.Errmsg
			}
			// check response
			if ret && router.doc.Output != nil {
				resp := map[string]interface{}{
					"status": response.Status,
					"data":   response.Data,
					"errmsg": response.Errmsg}
				ret, context.Status = router.doc.Output.check(resp)
				if !ret {
					Debug("output check faild")
				}
			}
			if context.Status != 0 {
				response.Status = context.Status
			}

			out, err := json.Marshal(response)
			if err != nil {
				Error(err)
			}
			Info(
				"<-",
				time.Now().Sub(startTime),
				context.Host,
				context.Path,
				context.Params,
				"->",
				context.Status,
				context.Data,
				context.Errmsg)
			w.Write(out)
		}
	}
}

// 处理参数，从请求中提取所有参数
func (router *Router) processParams(req *http.Request) map[string]interface{} {
	err := req.ParseForm()
	if err != nil {
		Error(err)
	}
	params := make(map[string]interface{})
	for k, vs := range req.Form {
		if len(vs) > 0 {
			var dat map[string]interface{}
			if err := json.Unmarshal([]byte(vs[0]), &dat); err != nil {
				params[k] = vs[0]
			} else {
				params[k] = dat
			}
		} else {
			params[k] = ""
		}
	}
	return params
}

// 创建一个路由
func newRouter(path string, filterChains ...Filter) *Router {
	router := &Router{}
	router.path = path
	docPath := "doc"
	if path[len(path)-1] != '/' {
		docPath = "/" + docPath
	}
	router.docPath = path + docPath

	doc := &Doc{}
	doc.Path = router.path
	doc.docPath = router.docPath
	router.doc = doc

	router.handler = router.genHandler(filterChains...)
	router.docHandler = router.genDocHandler()
	return router
}

// 创建一个带有doc的路由
func newDocRouter(doc *Doc, filterChains ...Filter) *Router {
	router := &Router{}
	router.path = doc.Path
	docPath := "doc"
	if doc.Path[len(doc.Path)-1] != '/' {
		docPath = "/" + docPath
	}
	router.doc = doc
	router.docPath = doc.Path + docPath
	router.handler = router.genHandler(filterChains...)
	router.docHandler = router.genDocHandler()
	doc.docPath = router.docPath
	return router
}

// genDocHandler 生成一个doc的页面
func (router *Router) genDocHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		docs := genDoc(router)
		ret := ""
		for _, doc := range docs {
			ret = ret + doc.genView()
		}
		ret = "<!doctype html>" +
			"<title>api doc - general by coral</title>" +
			"<h1>Api doc</h1>" +
			"<pre>" +
			`
#STATUS_*		若参数不满足要求，则返回错误码STATUS_*
<NOTE>			参数相关说明

string			任意字符串
string(n)		长度为n的字符串
string[m,n]		长度不小于m不大于n的字符串
string{a,b,c}	a,b,c其中一个字符串
int				任意整数
int(n)			整数n
int[m,n]		不小于m不大于n的整数
int{a,b,c}		a,b,c其中一个整数
mobile
md5
` +
			"</pre>" +
			ret +
			"<hr><p>@general by coral</p>"
		w.Write([]byte(ret))
	}
}

// 递归生成路由的doc
func genDoc(router *Router) []*Doc {
	var ret []*Doc
	ret = append(ret, router.doc)
	for _, child := range router.routers {
		docs := genDoc(child)
		for _, d := range docs {
			ret = append(ret, d)
		}
	}
	return ret
}

// 生成doc的html
func (doc *Doc) genView() string {
	if doc == nil {
		return ""
	}
	ret := "<hr>" +
		"<p>" +
		"<a href='" + doc.docPath +
		"' title='click to see sub tree'>@path:</a> " + doc.Path +
		"</p>"
	if doc.Description != "" {
		ret = ret + "<p>" + doc.Description + "</p>"
	}
	if doc.Input != nil {
		ret = ret + "<p>:- input</p>"
		ret = ret + "<pre>{\n" + doc.Input.genView("\t") + "}</pre>"
	}
	if doc.Output != nil {
		ret = ret + "<p>:- output</p>"
		ret = ret + "<pre>{\n" + doc.Output.genView("\t") + "}</pre>"
	}
	return ret
}

// 生成checker的doc
func (field Checker) genView(prefix string) string {
	if field == nil {
		return ""
	}
	ret := ""
	var keys []string
	for key := range field {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := field[key]
		if ret != "" {
			ret = ret + ",\n"
		}
		switch value := value.(type) {
		case Checker:
			ret = ret + prefix + key +
				": {\n" + value.genView(prefix+"\t") + prefix + "}"
			break
		case string:
			ret = ret + prefix + key + ": " + value
			break
		case []string:
			ret = ret + prefix + key + ": ["
			list := ""
			for _, sf := range value {
				if list != "" {
					list = list + ","
				}
				list = list + sf
			}
			ret = ret + list + ",...]"
			break
		case []Checker:
			ret = ret + prefix + key + ": ["
			list := ""
			for _, sf := range value {
				if list != "" {
					list = list + ",\n" + prefix
				}
				list = list + "{\n" + sf.genView(prefix+"\t") + prefix + "}"
			}
			ret = ret + list + ",...]"
			break
		default:
			Error("doc build error: unexpect rule", key, value)
		}
	}
	return ret + "\n"
}

// check 检查参数是否合法
func (field Checker) check(params map[string]interface{}) (bool, int) {
	for key, value := range field {
		switch value := value.(type) {
		case Checker:
			switch ele := params[key].(type) {
			case map[string]interface{}:
				// 如果是嵌套，那么参数必须也是嵌套的
				ret, status := value.check(ele)
				if !ret {
					return ret, status
				}
			default:
				Debug("param check", key, ele)
				return false, STATUS_INVALID_PARAM
			}
			break
		case string:
			ret, status := checkRule(params[key], value)
			if !ret {
				return ret, status
			}
			break
		case []string:
			if len(value) < 1 {
				Error("unexpect checker rule", key, value)
				break
			}
			switch eles := params[key].(type) {
			case []interface{}:
				for _, ele := range eles {
					ret, status := checkRule(ele, value[0])
					if !ret {
						return ret, status
					}
				}
				break
			default:
				Debug("param check", key, eles)
				return false, STATUS_INVALID_PARAM
			}
			break
		case []Checker:
			// 如果是数组嵌套，那么只检查数组第一项的规则
			if len(value) < 1 {
				Error("unexpect checker rule", key, value)
				break
			}
			switch eles := params[key].(type) {
			case []interface{}:
				// 还要保证要检验的参数也是数组
				for _, ele := range eles {
					switch ele := ele.(type) {
					case map[string]interface{}:
						// 数组里边的数也要求是嵌套的
						ret, status := value[0].check(ele)
						if !ret {
							return ret, status
						}
					default:
						Debug("param check", key, ele)
						return false, STATUS_INVALID_PARAM
					}
				}
				break
			default:
				Debug("param check", key, eles)
				return false, STATUS_INVALID_PARAM
			}
			break
		default:
			Error("unexpect checker rule", key, value)
		}
	}
	return true, STATUS_SUCCESS
}

func checkRule(param interface{}, rule string) (bool, int) {
	rules := strings.Split(rule, "|")
	for _, singleRule := range rules {
		if ret, status := checkSingleRule(param, singleRule); !ret {
			return ret, status
		}
	}
	return true, STATUS_SUCCESS
}

// string			任意字符串
// string(n)		长度为n的字符串
// string[m,n]		长度不小于m不大于n的字符串
// string{a,b,c}	a,b,c其中一个字符串
// int				任意整数
// int(n)			整数n
// int[m,n]			不小于m不大于n的整数
// int{a,b,c}		a,b,c其中一个整数
// mobile
// md5
func checkSingleRule(param interface{}, singleRule string) (bool, int) {
	var status int
	// 提取;后边的注释
	tmparr := strings.Split(singleRule, "<")
	// 提取#后面的错误码
	tmparr = strings.Split(tmparr[0], "#")
	if len(tmparr) > 1 {
		singleRule = tmparr[0]
		st, err := strconv.Atoi(tmparr[1])
		if err != nil {
			Debug("status not a int number", singleRule)
			status = STATUS_INVALID_PARAM
		} else {
			status = st
		}
	} else {
		status = STATUS_INVALID_PARAM
	}

	// 根据rule选择处理方式
	// 只处理已知类型，位置类型的全部不能通过
	switch {
	case len(singleRule) >= 6 && singleRule[0:6] == "string":
		// string只能是string
		switch param := param.(type) {
		case string:
			if checkString(singleRule, param) {
				return true, STATUS_SUCCESS
			}
			break
		}
		break
	case len(singleRule) >= 3 && singleRule[0:3] == "int":
		// int 可以是int, float64，也可以是string转int
		switch param := param.(type) {
		case float64:
			if checkInt(singleRule, int(param)) {
				return true, STATUS_SUCCESS
			}
			break
		case int64:
			if checkInt(singleRule, int(param)) {
				return true, STATUS_SUCCESS
			}
			break
		case int32:
			if checkInt(singleRule, int(param)) {
				return true, STATUS_SUCCESS
			}
			break
		case int:
			if checkInt(singleRule, param) {
				return true, STATUS_SUCCESS
			}
		case string:
			if param, err := strconv.Atoi(param); err != nil {
				Debug("check rule faild", singleRule, param, err.Error())
			} else {
				if checkInt(singleRule, param) {
					return true, STATUS_SUCCESS
				}
			}
		default:
			Debug("check rule faild", "unexpect int type", param, fmt.Sprintf("%T", param))
		}
		break
	case singleRule == "mobile":
		switch param := param.(type) {
		case string:
			if len(param) == 11 {
				return true, STATUS_SUCCESS
			}
		default:
			Debug("check rule faild", singleRule, param, "mobile must be string")
		}
	case singleRule == "md5":
		switch param := param.(type) {
		case string:
			if len(param) == 32 || len(param) == 64 {
				return true, STATUS_SUCCESS
			}
		default:
			Debug("check rule faild", singleRule, param, "md5 must be string")
		}
	default:
		Error("unknown rule", singleRule)
	}
	Debug("check rule faild", singleRule, param, fmt.Sprintf("%T", param))
	return false, status
}

func checkString(rule, param string) bool {
	rule = rule[6:]
	if len(rule) > 0 {
		switch rule[0] {
		case '(':
			return checkPoint(rule, len(param))
		case '[':
			return checkRange(rule, len(param))
		case '{':
			return checkIn(rule, param)
		}
	}
	return true
}
func checkInt(rule string, param int) bool {
	rule = rule[3:]
	if len(rule) > 0 {
		switch rule[0] {
		case '(':
			return checkPoint(rule, param)
		case '[':
			return checkRange(rule, param)
		case '{':
			return checkIn(rule, strconv.Itoa(param))
		}
	}
	return true
}
func checkPoint(rule string, point int) bool {
	if len(rule) > 2 && rule[0] == '(' && rule[len(rule)-1] == ')' {
		rule = rule[1 : len(rule)-1]
		if strconv.Itoa(point) != rule {
			Debug("check point faild", rule, point)
			return false
		}
	}
	return true
}
func checkRange(rule string, point int) bool {
	if len(rule) > 5 && rule[0] == '[' && rule[len(rule)-1] == ']' {
		rule = rule[1 : len(rule)-1]
		tmparr := strings.Split(rule, ",")
		if len(tmparr) != 2 {
			Debug("check range faild, invalid rule", rule, point)
			return false
		}
		min, err := strconv.Atoi(tmparr[0])
		if err != nil {
			Debug("check range faild, invalid rule", rule, point, err.Error())
			return false
		}
		max, err := strconv.Atoi(tmparr[1])
		if err != nil {
			Debug("check range faild, invalid rule", rule, point, err.Error())
			return false
		}
		ret := true
		if min >= 0 {
			ret = point >= min
		}
		if max >= 0 {
			ret = point <= max
		}
		if !ret {
			Debug("check range faild", rule, point)
			return false
		}
	}
	return true
}
func checkIn(rule string, point string) bool {
	if len(rule) > 2 && rule[0] == '{' && rule[len(rule)-1] == '}' {
		rule = rule[1 : len(rule)-1]
		tmparr := strings.Split(rule, ",")
		if len(tmparr) < 1 {
			return false
		}
		ret := false
		for _, ex := range tmparr {
			if point == ex {
				ret = true
			}
		}
		if !ret {
			Debug("check in faild", rule, point)
		}
	}
	return true
}
