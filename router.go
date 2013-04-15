package xgo

import (
	"compress/gzip"
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

type Route struct {
	pattern     string
	slashCnt    int
	regexp      *regexp.Regexp
	params      []string
	staticParts []string
	schemes     []string
	handlerType reflect.Type
}

func (this *Route) Schemes(schemes ...string) {
	for _, scheme := range schemes {
		this.schemes = append(this.schemes, scheme)
	}
}

type xgoRouter struct {
	app            *App
	Routes         []*Route
	StaticRoutes   map[string]*Route
	StaticFileDir  map[string]int
	StaticFileType map[string]int
	PrefixPath     string
	lock           *sync.Mutex
}

func (this *xgoRouter) AddStaticFileDir(dirs ...string) {
	this.lock.Lock()
	defer this.lock.Unlock()

	for _, dir := range dirs {
		this.StaticFileDir[dir] += 1
	}
}

func (this *xgoRouter) RemoveStaticFileDir(dirs ...string) {
	this.lock.Lock()
	defer this.lock.Unlock()

	for _, dir := range dirs {
		this.StaticFileDir[dir] -= 1
		if this.StaticFileDir[dir] <= 0 {
			delete(this.StaticFileDir, dir)
		}
	}
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

func (this *xgoRouter) AddRoute(pattern string, c HandlerInterface) *Route {
	this.lock.Lock()
	defer this.lock.Unlock()

	handlerType := reflect.Indirect(reflect.ValueOf(c)).Type()

	if pattern[0] != '/' {
		pattern = "/" + pattern
	}
	route := &Route{
		pattern:     pattern,
		slashCnt:    strings.Count(pattern, "/"),
		regexp:      nil,
		params:      []string{},
		staticParts: strings.Split(pattern, "/"),
		schemes:     []string{},
		handlerType: handlerType,
	}
	paramCnt := strings.Count(pattern, "{")
	if paramCnt != strings.Count(pattern, "}") {
		paramCnt = 0
	}
	if paramCnt == 0 {
		this.StaticRoutes[pattern] = route
	} else {
		re, err := regexp.Compile(`\{\w+?\(.*?\)\}|\{\w+?\}`)
		if err != nil {
			panic(err)
		}
		matches := re.FindAllStringSubmatch(pattern, paramCnt)
		if len(matches) != paramCnt {
			panic("Regexp match error")
		}
		for i, part := range route.staticParts {
			if strings.Index(part, "{") >= 0 {
				route.staticParts[i] = "/"
			}
		}
		for _, match := range matches {
			m := match[0][1 : len(match[0])-1]
			index := strings.Index(m, "(")
			if index == -1 {
				index = len(m)
				m = m + "(.+)"
			}
			route.params = append(route.params, m[0:index])
			pattern = strings.Replace(pattern, match[0], m[index:], 1)
		}
		re, err = regexp.Compile(pattern)
		if err != nil {
			panic(err)
		}
		route.regexp = re
		this.Routes = append(this.Routes, route)
	}
	return route
}

func (this *xgoRouter) RemoveRoute(pattern string) {
	this.lock.Lock()
	defer this.lock.Unlock()

	paramCnt := strings.Count(pattern, "{")
	if paramCnt != strings.Count(pattern, "}") {
		paramCnt = 0
	}
	if paramCnt > 0 {
		length := len(this.Routes)
		for i, route := range this.Routes {
			if route.pattern == pattern {
				if i == length-1 {
					this.Routes = this.Routes[:i]
				} else {
					this.Routes = append(this.Routes[:i], this.Routes[i+1:]...)
				}
				break
			}
		}
	} else {
		delete(this.StaticRoutes, pattern)
	}
}

func (this *xgoRouter) SetPrefixPath(prefix string) {
	if prefix[0] != '/' {
		prefix = "/" + prefix
	}
	if prefix[len(prefix)-1] == '/' {
		prefix = prefix[:len(prefix)-1]
	}
	this.PrefixPath = prefix
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

	if this.PrefixPath != "" {
		if !strings.HasPrefix(r.URL.Path, this.PrefixPath+"/") {
			http.NotFound(w, r)
			return
		}
		r.URL.Path = r.URL.Path[len(this.PrefixPath):]
	}
	urlPath := r.URL.Path
	urlScheme := r.URL.Scheme
	//static file server
	if r.Method == "GET" || r.Method == "HEAD" {
		dotIndex := strings.LastIndex(urlPath, ".")
		if dotIndex != -1 {
			if _, ok := this.StaticFileType[urlPath[dotIndex:]]; ok {
				http.ServeFile(w, r, filepath.Join(AppRoot, urlPath))
				return
			}
		}
		dir := urlPath[1:]
		if slashIndex := strings.Index(dir, "/"); slashIndex > 0 {
			dir := dir[:slashIndex]
			if _, ok := this.StaticFileDir[dir]; ok {
				http.ServeFile(w, r, filepath.Join(AppRoot, urlPath))
				return
			}
		}
	}

	//first find path from the fixrouters to Improve Performance
	if route, ok := this.StaticRoutes[urlPath]; ok {
		if len(route.schemes) > 0 {
			ok := false
			for _, scheme := range route.schemes {
				if urlScheme == scheme {
					ok = true
					break
				}
			}
			if ok {
				handlerType = route.handlerType
			}
		} else {
			handlerType = route.handlerType
		}
	}

	if handlerType == nil {
		slashCnt := strings.Count(urlPath, "/")
		parts := strings.Split(urlPath, "/")
		for _, route := range this.Routes {
			if slashCnt != route.slashCnt {
				continue
			}
			ok := true
			for i, part := range route.staticParts {
				if part != "/" && part != parts[i] {
					ok = false
					break
				}
			}
			if !ok {
				continue
			}
			if len(route.schemes) > 0 {
				ok := false
				for _, scheme := range route.schemes {
					if urlScheme == scheme {
						ok = true
						break
					}
				}
				if !ok {
					continue
				}
			}
			if !route.regexp.MatchString(urlPath) {
				continue
			}
			matches := route.regexp.FindStringSubmatch(urlPath)
			if matches[0] != urlPath {
				continue
			}
			matches = matches[1:]
			paramCnt := len(route.params)
			if paramCnt != len(matches) {
				continue
			}
			if paramCnt > 0 {
				values := r.URL.Query()
				for i, match := range matches {
					values.Add(route.params[i], match)
				}
				r.URL.RawQuery = values.Encode()
			}
			handlerType = route.handlerType
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
