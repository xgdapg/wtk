package xgo

import (
	"errors"
	"net/http"
	// "net/url"
	"reflect"
	"regexp"
	"strings"
)

type responseWriter struct {
	writer    http.ResponseWriter
	HasOutput bool
}

func (this *responseWriter) Header() http.Header {
	return this.writer.Header()
}

func (this *responseWriter) Write(p []byte) (int, error) {
	this.HasOutput = true
	return this.writer.Write(p)
}

func (this *responseWriter) WriteHeader(code int) {
	this.HasOutput = true
	this.writer.WriteHeader(code)
}

type RoutingRule struct {
	Pattern        string
	Regexp         *regexp.Regexp
	Params         []string
	ControllerType reflect.Type
}

type Router struct {
	Rules       []*RoutingRule
	StaticRules []*RoutingRule
}

func (this *Router) AddRule(pattern string, c ControllerInterface) error {
	rule := &RoutingRule{
		Pattern:        "",
		Regexp:         nil,
		Params:         []string{},
		ControllerType: reflect.Indirect(reflect.ValueOf(c)).Type(),
	}
	paramCnt := strings.Count(pattern, ":")
	if paramCnt > 0 {
		re, err := regexp.Compile(`:\w+\(.*?\)`)
		if err != nil {
			return err
		}
		matches := re.FindAllStringSubmatch(pattern, paramCnt)
		if len(matches) != paramCnt {
			return errors.New("Regexp match error")
		}
		for _, match := range matches {
			m := match[0]
			index := strings.Index(m, "(")
			rule.Params = append(rule.Params, m[0:index])
			pattern = "^" + strings.Replace(pattern, m, m[index:], 1)
		}
		re, err = regexp.Compile(pattern)
		if err != nil {
			return err
		}
		rule.Regexp = re
		this.Rules = append(this.Rules, rule)
	} else {
		rule.Pattern = pattern
		this.StaticRules = append(this.StaticRules, rule)
	}
	return nil
}

func (this *Router) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// defer func() {
	// if err := recover(); err != nil {
	// fmt.Println("RECOVER:", err)
	// if !RecoverPanic {
	// 	panic(err)
	// } else {
	// 	Critical("Handler crashed with error", err)
	// 	for i := 1; ; i += 1 {
	// 		_, file, line, ok := runtime.Caller(i)
	// 		if !ok {
	// 			break
	// 		}
	// 		Critical(file, line)
	// 	}
	// }
	// }
	// }()

	w := &responseWriter{
		writer:    rw,
		HasOutput: false,
	}
	var routingRule *RoutingRule
	urlPath := r.URL.Path
	pathLen := len(urlPath)
	pathEnd := urlPath[pathLen-1]

	//static file server
	// for prefix, staticDir := range StaticDir {
	// 	if strings.HasPrefix(r.URL.Path, prefix) {
	// 		file := staticDir + r.URL.Path[len(prefix):]
	// 		http.ServeFile(w, r, file)
	// 		return
	// 	}
	// }

	//first find path from the fixrouters to Improve Performance
	for _, rule := range this.StaticRules {
		if urlPath == rule.Pattern || (pathEnd == '/' && urlPath[:pathLen-1] == rule.Pattern) {
			routingRule = rule
			break
		}
	}

	if routingRule == nil {
		for _, rule := range this.Rules {
			if !rule.Regexp.MatchString(urlPath) {
				continue
			}
			matches := rule.Regexp.FindStringSubmatch(urlPath)[1:]
			paramCnt := len(rule.Params)
			if paramCnt != len(matches) {
				continue
			}
			if paramCnt > 0 {
				values := r.URL.Query()
				for i, match := range matches {
					values.Add(rule.Params[i], match)
				}
				r.URL.RawQuery = values.Encode()
			}
			routingRule = rule
			break
		}
	}

	if routingRule == nil {
		http.NotFound(w, r)
		return
	}

	//execute middleware filters
	// for _, filter := range p.filters {
	// 	filter(w, r)
	// 	if w.HasOutput {
	// 		return
	// 	}
	// }
	r.ParseForm()
	ci := reflect.New(routingRule.ControllerType).Interface()
	// ci := reflect.Indirect(reflect.New(routingRule.ControllerType)).Interface()
	ctx := &Context{
		Response: w,
		Request:  r,
	}
	util.CallMethod(ci, "Init", ctx, routingRule.ControllerType.Name())
	if w.HasOutput {
		return
	}

	var method string
	switch r.Method {
	case "GET":
		method = "Get"
	case "POST":
		method = "Post"
	case "HEAD":
		method = "Head"
	case "DELETE":
		method = "Delete"
	case "PUT":
		method = "Put"
	case "PATCH":
		method = "Patch"
	case "OPTIONS":
		method = "Options"
	default:
		http.Error(w, "Method Not Allowed", 405)
	}
	util.CallMethod(ci, method)
	if w.HasOutput {
		return
	}

	//If we need other filter to process datas, call it on this step.

	util.CallMethod(ci, "Render")
	if w.HasOutput {
		return
	}

	util.CallMethod(ci, "Output")
	if w.HasOutput {
		return
	}
}
