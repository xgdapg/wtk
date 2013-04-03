package xgo

import (
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

type xgoResponseWriter struct {
	app      *App
	writer   http.ResponseWriter
	request  *http.Request
	Closed   bool
	Finished bool
}

func (this *xgoResponseWriter) Header() http.Header {
	return this.writer.Header()
}

func (this *xgoResponseWriter) Write(p []byte) (int, error) {
	if this.Closed {
		return 0, nil
	}

	var writer io.Writer = this.writer

	useGzip := false
	if EnableGzip &&
		strings.Contains(this.request.Header.Get("Accept-Encoding"), "gzip") &&
		len(p) >= GzipMinLength {
		ctype := this.writer.Header().Get("Content-Type")
		for _, t := range GzipTypes {
			if strings.Contains(ctype, t) {
				useGzip = true
				break
			}
		}
	}
	if useGzip {
		this.Header().Set("Content-Encoding", "gzip")
		this.Header().Del("Content-Length")

		gz := gzip.NewWriter(this.writer)
		defer gz.Close()
		writer = gz
	}

	return writer.Write(p)
}

func (this *xgoResponseWriter) WriteHeader(code int) {
	if this.Closed {
		return
	}

	if code != http.StatusOK {
		this.writer.WriteHeader(code)
	}

	if filepath, ok := app.customHttpStatus[code]; ok {
		content, err := ioutil.ReadFile(filepath)
		if err != nil {
			content = []byte(http.StatusText(code))
		}
		this.Write(content)
		this.Close()
	}
}

func (this *xgoResponseWriter) Close() {
	this.Closed = true
}

type xgoRoutingRule struct {
	Pattern     string
	Regexp      *regexp.Regexp
	Params      []string
	HandlerType reflect.Type
}

type xgoRouter struct {
	app            *App
	Rules          []*xgoRoutingRule
	StaticRules    map[string]reflect.Type
	StaticDir      map[string]string
	StaticFileType map[string]int
	lock           *sync.Mutex
}

func (this *xgoRouter) AddStaticPath(sPath, fPath string) {
	this.lock.Lock()
	defer this.lock.Unlock()

	this.StaticDir[sPath] = fPath
}

func (this *xgoRouter) RemoveStaticPath(sPath string) {
	this.lock.Lock()
	defer this.lock.Unlock()

	delete(this.StaticDir, sPath)
}

func (this *xgoRouter) AddStaticFileType(exts ...string) {
	this.lock.Lock()
	defer this.lock.Unlock()

	for _, ext := range exts {
		if ext[0] != '.' {
			ext = "." + ext
		}
		this.StaticFileType[ext] += 1
	}
}

func (this *xgoRouter) RemoveStaticFileType(exts ...string) {
	this.lock.Lock()
	defer this.lock.Unlock()

	for _, ext := range exts {
		if ext[0] != '.' {
			ext = "." + ext
		}
		this.StaticFileType[ext] -= 1
		if this.StaticFileType[ext] <= 0 {
			delete(this.StaticFileType, ext)
		}
	}
}

func (this *xgoRouter) AddRule(pattern string, c HandlerInterface) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	handlerType := reflect.Indirect(reflect.ValueOf(c)).Type()
	paramCnt := strings.Count(pattern, ":")
	if paramCnt != strings.Count(pattern, "(") || paramCnt != strings.Count(pattern, ")") {
		paramCnt = 0
	}
	if paramCnt > 0 {
		re, err := regexp.Compile(`:\w+\(.*?\)`)
		if err != nil {
			return err
		}
		matches := re.FindAllStringSubmatch(pattern, paramCnt)
		if len(matches) != paramCnt {
			return errors.New("Regexp match error")
		}
		rule := &xgoRoutingRule{
			Pattern:     pattern,
			Regexp:      nil,
			Params:      []string{},
			HandlerType: handlerType,
		}
		for _, match := range matches {
			m := match[0]
			index := strings.Index(m, "(")
			rule.Params = append(rule.Params, m[0:index])
			pattern = strings.Replace(pattern, m, m[index:], 1)
		}
		re, err = regexp.Compile(pattern)
		if err != nil {
			return err
		}
		rule.Regexp = re
		this.Rules = append(this.Rules, rule)
	} else {
		this.StaticRules[pattern] = handlerType
	}
	return nil
}

func (this *xgoRouter) RemoveRule(pattern string) {
	this.lock.Lock()
	defer this.lock.Unlock()

	paramCnt := strings.Count(pattern, ":")
	if paramCnt != strings.Count(pattern, "(") || paramCnt != strings.Count(pattern, ")") {
		paramCnt = 0
	}
	if paramCnt > 0 {
		length := len(this.Rules)
		for i, rule := range this.Rules {
			if rule.Pattern == pattern {
				if i == length-1 {
					this.Rules = this.Rules[:i]
				} else {
					this.Rules = append(this.Rules[:i], this.Rules[i+1:]...)
				}
				break
			}
		}
	} else {
		delete(this.StaticRules, pattern)
	}
}

func (this *xgoRouter) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	w := &xgoResponseWriter{
		app:      this.app,
		writer:   rw,
		request:  r,
		Closed:   false,
		Finished: false,
	}
	var handlerType reflect.Type
	urlPath := r.URL.Path

	//static file server
	if r.Method == "GET" || r.Method == "HEAD" {
		dotIndex := strings.LastIndex(urlPath, ".")
		if dotIndex != -1 {
			if _, ok := this.StaticFileType[urlPath[dotIndex:]]; ok {
				http.ServeFile(w, r, filepath.Join(AppRoot, urlPath))
				return
			}
		}
		for sPath, fPath := range this.StaticDir {
			if strings.HasPrefix(urlPath, sPath) {
				file := filepath.Join(AppRoot, fPath+urlPath[len(sPath):])
				http.ServeFile(w, r, file)
				return
			}
		}
	}

	//first find path from the fixrouters to Improve Performance
	if ht, ok := this.StaticRules[urlPath]; ok {
		handlerType = ht
	}

	if handlerType == nil {
		for _, rule := range this.Rules {
			if !rule.Regexp.MatchString(urlPath) {
				continue
			}
			matches := rule.Regexp.FindStringSubmatch(urlPath)
			if matches[0] != urlPath {
				continue
			}
			matches = matches[1:]
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
			handlerType = rule.HandlerType
			break
		}
	}

	if handlerType == nil {
		http.NotFound(w, r)
		return
	}

	ci := reflect.New(handlerType).Interface()
	ctx := &Context{
		hdlr:           nil,
		response:       w,
		ResponseWriter: w,
		Request:        r,
	}
	tpl := &Template{
		hdlr:      nil,
		tpl:       nil,
		Vars:      make(map[string]interface{}),
		tplResult: nil,
	}
	sess := &Session{
		hdlr:           nil,
		sessionManager: this.app.session,
		sessionId:      ctx.GetSecureCookie(SessionName),
		ctx:            ctx,
		data:           nil,
	}
	util.CallMethod(ci, "Init", this.app, ctx, tpl, sess, handlerType.Name())
	if w.Finished {
		return
	}

	hc := &HookHandler{
		Context:  ctx,
		Template: tpl,
		Session:  sess,
	}

	this.app.callHandlerHook("AfterInit", hc)
	if w.Finished {
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

	this.app.callHandlerHook("BeforeMethod"+method, hc)
	if w.Finished {
		return
	}

	util.CallMethod(ci, method)
	if w.Finished {
		return
	}

	this.app.callHandlerHook("AfterMethod"+method, hc)
	if w.Finished {
		return
	}

	util.CallMethod(ci, "Render")
	if w.Finished {
		return
	}

	util.CallMethod(ci, "Output")
	if w.Finished {
		return
	}
}
