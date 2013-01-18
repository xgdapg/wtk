package xgo

import (
	"errors"
	"net/http"
	"net/url"
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
		ControllerType: reflect.TypeOf(*c),
	}
	paramCnt := strings.Count(pattern, ":")
	if paramCnt > 0 {
		re, err := regexp.Compile(`:\w+\(.*?\)`)
		if err != nil {
			return err
		}
		matches := re.FindStringSubmatch(pattern)[1:]
		if len(matches) != paramCnt {
			return errors.New("Regexp match error")
		}
		for _, match := range matches {
			index := strings.Index(match, "(")
			rule.Params = append(rule.Params, match[0:index])
			pattern = strings.Replace(pattern, match, match[index:], 1)
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
	defer func() {
		if err := recover(); err != nil {
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
		}
	}()

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

	cv := reflect.New(routingRule.ControllerType)
	init := cv.MethodByName("Init")
	in := make([]reflect.Value, 2)
	ctx := &Context{
		Response: w,
		Request:  r,
	}
	in[0] = reflect.ValueOf(ctx)
	in[1] = reflect.ValueOf(routingRule.ControllerType.Name())
	init.Call(in)
	if w.HasOutput {
		return
	}

	var method reflect.Value
	switch r.Method {
	case "GET":
		method = cv.MethodByName("Get")
	case "POST":
		method = cv.MethodByName("Post")
	case "HEAD":
		method = cv.MethodByName("Head")
	case "DELETE":
		method = cv.MethodByName("Delete")
	case "PUT":
		method = cv.MethodByName("Put")
	case "PATCH":
		method = cv.MethodByName("Patch")
	case "OPTIONS":
		method = cv.MethodByName("Options")
	default:
		http.Error(w, "Method Not Allowed", 405)
	}
	method.Call([]reflect.Value{})
}
