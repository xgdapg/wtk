package wtk

import (
	"compress/gzip"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type wtkResponseWriter struct {
	server     *Server
	request    *http.Request
	writer     http.ResponseWriter
	gzipWriter *gzip.Writer
	httpStatus int
	Closed     bool
	Finished   bool
}

func (this *wtkResponseWriter) Header() http.Header {
	return this.writer.Header()
}

func (this *wtkResponseWriter) Write(p []byte) (int, error) {
	if this.Closed {
		return 0, nil
	}

	if this.gzipWriter != nil {
		this.Header().Set("Content-Encoding", "gzip")
		this.Header().Del("Content-Length")
	}

	if this.httpStatus > 0 {
		this.writer.WriteHeader(this.httpStatus)
		this.httpStatus = 0
	}

	if this.gzipWriter != nil {
		return this.gzipWriter.Write(p)
	}
	return this.writer.Write(p)
}

func (this *wtkResponseWriter) WriteHeader(code int) {
	if this.Closed {
		return
	}
	this.httpStatus = code

	handler := &Handler{}
	handler.init(this.server, this, this.request)
	hc := handler.getHookHandler()
	this.server.callHandlerHook("HttpStatus"+strconv.Itoa(code), hc)
	if this.Closed {
		return
	}
	if code != http.StatusOK {
		this.writer.WriteHeader(code)
		this.Close()
	}
}

func (this *wtkResponseWriter) Close() {
	this.Closed = true
}

type Route struct {
	pattern  string
	slashCnt int
	regexp   *regexp.Regexp
	params   []string
	scheme   string
	handler  HandlerInterface
}

func (this *Route) Scheme(scheme string) {
	this.scheme = scheme
}

type wtkRouteCache struct {
	Route *Route
	Vars  url.Values
}

type wtkRouter struct {
	server         *Server
	Routes         []*Route
	StaticRoutes   map[string]*Route
	StaticFileDir  map[string]int
	StaticFileType map[string]int
	PrefixPath     string
	lock           *sync.Mutex
	routeCache     map[string]*wtkRouteCache
}

func (this *wtkRouter) ClearRouteCache() {
	this.routeCache = make(map[string]*wtkRouteCache)
}

func (this *wtkRouter) AddStaticFileDir(dirs ...string) {
	this.lock.Lock()
	defer this.lock.Unlock()

	for _, dir := range dirs {
		this.StaticFileDir[dir] += 1
	}
}

func (this *wtkRouter) RemoveStaticFileDir(dirs ...string) {
	this.lock.Lock()
	defer this.lock.Unlock()

	for _, dir := range dirs {
		this.StaticFileDir[dir] -= 1
		if this.StaticFileDir[dir] <= 0 {
			delete(this.StaticFileDir, dir)
		}
	}
}

func (this *wtkRouter) AddStaticFileType(exts ...string) {
	this.lock.Lock()
	defer this.lock.Unlock()

	for _, ext := range exts {
		if ext[0] != '.' {
			ext = "." + ext
		}
		this.StaticFileType[ext] += 1
	}
}

func (this *wtkRouter) RemoveStaticFileType(exts ...string) {
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

func (this *wtkRouter) AddRoute(pattern string, handler HandlerInterface) *Route {
	this.lock.Lock()
	defer this.lock.Unlock()

	if pattern[0] != '/' {
		pattern = "/" + pattern
	}
	route := &Route{
		pattern:  pattern,
		slashCnt: strings.Count(pattern, "/"),
		regexp:   nil,
		params:   []string{},
		scheme:   "",
		handler:  handler,
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
		for _, match := range matches {
			m := match[0][1 : len(match[0])-1]
			index := strings.Index(m, "(")
			if index == -1 {
				index = len(m)
				m = m + "([^/]+)"
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
		if EnableRouteCache {
			this.ClearRouteCache()
		}
	}
	return route
}

func (this *wtkRouter) RemoveRoute(pattern string) {
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
		if EnableRouteCache {
			this.ClearRouteCache()
		}
	} else {
		delete(this.StaticRoutes, pattern)
	}
}

func (this *wtkRouter) SetPrefixPath(prefix string) {
	if prefix[0] != '/' {
		prefix = "/" + prefix
	}
	if prefix[len(prefix)-1] == '/' {
		prefix = prefix[:len(prefix)-1]
	}
	this.PrefixPath = prefix
}

func (this *wtkRouter) getFileSize(name string) (int64, error) {
	dir, file := filepath.Split(name)
	f, err := http.Dir(dir).Open(file)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	s, err := f.Stat()
	if err != nil {
		return 0, err
	}
	return s.Size(), nil
}

func (this *wtkRouter) serveFile(w *wtkResponseWriter, r *http.Request, name string, fileType string) {
	if EnableGzip {
		fileType = fileType[1:]
		useGzip := false
		for _, t := range GzipTypes {
			if fileType == t {
				useGzip = true
				break
			}
		}
		if useGzip {
			fileSize, err := this.getFileSize(name)
			if err == nil && fileSize < int64(GzipMinLength) {
				useGzip = false
			}
		}
		if !useGzip {
			w.gzipWriter = nil
		}
	}
	http.ServeFile(w, r, name)
}

func (this *wtkRouter) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	w := &wtkResponseWriter{
		server:     this.server,
		request:    r,
		writer:     rw,
		gzipWriter: nil,
		httpStatus: 0,
		Closed:     false,
		Finished:   false,
	}
	if EnableGzip && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.gzipWriter = gzip.NewWriter(w.writer)
		defer func(w *wtkResponseWriter) {
			if w.gzipWriter != nil {
				w.gzipWriter.Close()
			}
		}(w)
	}

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
		fileType := ""
		if dotIndex != -1 {
			fileType = urlPath[dotIndex:]
			if _, ok := this.StaticFileType[fileType]; ok {
				this.serveFile(w, r, filepath.Join(AppRoot, urlPath), fileType)
				return
			}
		}
		dir := urlPath[1:]
		if slashIndex := strings.Index(dir, "/"); slashIndex > 0 {
			dir := dir[:slashIndex]
			if _, ok := this.StaticFileDir[dir]; ok {
				this.serveFile(w, r, filepath.Join(AppRoot, urlPath), fileType)
				return
			}
		}
	}

	var handler HandlerInterface

	if route, ok := this.StaticRoutes[urlPath]; ok {
		if route.scheme == "" || urlScheme == route.scheme {
			handler = route.handler
		}
	}

	pathVars := make(url.Values)
	if EnableRouteCache {
		if rc, ok := this.routeCache[urlPath]; ok {
			handler = rc.Route.handler
			pathVars = rc.Vars
		}
	}
	if handler == nil {
		slashCnt := strings.Count(urlPath, "/")
		for _, route := range this.Routes {
			if slashCnt != route.slashCnt {
				continue
			}
			if route.scheme != "" && urlScheme != route.scheme {
				continue
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
				for i, match := range matches {
					pathVars.Add(route.params[i], match)
				}
			}
			handler = route.handler
			if EnableRouteCache {
				this.routeCache[urlPath] = &wtkRouteCache{
					Route: route,
					Vars:  pathVars,
				}
			}
			break
		}
	}

	if handler == nil {
		http.NotFound(w, r)
		return
	}

	handlerType := reflect.Indirect(reflect.ValueOf(handler)).Type()
	newHandler := reflect.New(handlerType).Interface().(HandlerInterface)

	newHandler.init(this.server, w, r)
	newHandler.context().pathVars = pathVars

	if w.Finished {
		return
	}

	hc := newHandler.getHookHandler()

	this.server.callHandlerHook("AfterInit", hc)
	if w.Finished {
		return
	}

	var methodFunc func()
	var method string
	switch r.Method {
	case "GET":
		method = "Get"
		methodFunc = newHandler.Get
	case "POST":
		method = "Post"
		methodFunc = newHandler.Post
	case "HEAD":
		method = "Head"
		methodFunc = newHandler.Head
	case "DELETE":
		method = "Delete"
		methodFunc = newHandler.Delete
	case "PUT":
		method = "Put"
		methodFunc = newHandler.Put
	case "PATCH":
		method = "Patch"
		methodFunc = newHandler.Patch
	case "OPTIONS":
		method = "Options"
		methodFunc = newHandler.Options
	default:
		http.Error(w, "Method Not Allowed", 405)
		return
	}

	this.server.callHandlerHook("BeforeMethod"+method, hc)
	if w.Finished {
		return
	}

	methodFunc()
	if w.Finished {
		return
	}

	this.server.callHandlerHook("AfterMethod"+method, hc)
	if w.Finished {
		return
	}

	newHandler.Render()
	if w.Finished {
		return
	}

	newHandler.Output()
	if w.Finished {
		return
	}
}
