package wtk

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Server struct {
	Id       int
	RunMode  string
	listener net.Listener
	router   *wtkRouter
	hook     *wtkHook
	session  *wtkSessionManager
}

func (this *Server) init(id int) *Server {
	this.Id = id
	this.router = &wtkRouter{
		server:         this,
		Routes:         []*Route{},
		StaticRoutes:   make(map[string]*Route),
		StaticFileDir:  make(map[string]int),
		StaticFileType: make(map[string]int),
		lock:           new(sync.Mutex),
		routeCache:     make(map[string]*wtkRouteCache),
	}
	this.hook = &wtkHook{server: this}
	this.session = new(wtkSessionManager)
	this.session.RegisterStorage(new(wtkDefaultSessionStorage))
	return this
}

func (this *Server) AddRoute(pattern string, c HandlerInterface) *Route {
	return this.router.AddRoute(pattern, c)
}

func (this *Server) RemoveRoute(pattern string) {
	this.router.RemoveRoute(pattern)
}

func (this *Server) SetPrefixPath(prefix string) {
	this.router.SetPrefixPath(prefix)
}

func (this *Server) AddHandlerHook(event string, hookFunc HookHandlerFunc) {
	this.hook.AddHandlerHook(event, hookFunc)
}

func (this *Server) AddHttpStatusHook(statusCode int, hookFunc HookHandlerFunc) {
	this.hook.AddHandlerHook("HttpStatus"+strconv.Itoa(statusCode), hookFunc)
}

func (this *Server) SetHttpStatusPage(statusCode int, pageFile string) {
	this.AddHttpStatusHook(statusCode, func(h *HookHandler) {
		s, err := ioutil.ReadFile(pageFile)
		if err == nil {
			h.Context.WriteBytes(s)
		}
	})
}

func (this *Server) AddStaticFileDir(dirs ...string) {
	this.router.AddStaticFileDir(dirs...)
}

func (this *Server) RemoveStaticFileDir(dirs ...string) {
	this.router.RemoveStaticFileDir(dirs...)
}

func (this *Server) AddStaticFileType(ext ...string) {
	this.router.AddStaticFileType(ext...)
}

func (this *Server) RemoveStaticFileType(ext ...string) {
	this.router.RemoveStaticFileType(ext...)
}

func (this *Server) RegisterSessionStorage(storage SessionStorageInterface) {
	this.session.RegisterStorage(storage)
}

func (this *Server) Run(mode string, addr string, port int) error {
	var tlsConfig *tls.Config
	var err error
	if mode == "https" {
		tlsConfig = &tls.Config{}
		if tlsConfig.NextProtos == nil {
			tlsConfig.NextProtos = []string{"http/1.1"}
		}
		tlsConfig.Certificates = make([]tls.Certificate, 1)
		tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(SslCertificate, SslCertificateKey)
		if err != nil {
			return err
		}
	}

	listenUnix := false
	if strings.HasPrefix(addr, "unix:") {
		listenUnix = true
		addr = addr[5:]
	}

	if listenUnix {
		os.Remove(addr)
		this.listener, err = net.Listen("unix", addr)
		if err == nil {
			os.Chmod(addr, os.FileMode(0666))
			defer os.Remove(addr)
		}
	} else {
		listenAddr := net.JoinHostPort(addr, fmt.Sprintf("%d", port))
		this.listener, err = net.Listen("tcp", listenAddr)
	}
	if err != nil {
		return err
	}
	defer this.listener.Close()
	this.RunMode = mode

	switch mode {
	case "http":
		http.Serve(this.listener, this.router)
	case "fcgi":
		fcgi.Serve(this.listener, this.router)
	case "https":
		http.Serve(tls.NewListener(this.listener, tlsConfig), this.router)
	default:
		http.Serve(this.listener, this.router)
	}
	return nil
}

func (this *Server) Stop() {
	this.listener.Close()
}

func (this *Server) Close() {
	delete(servers, this.Id)
	this.Stop()
}

func (this *Server) Clone() *Server {
	a := NewServer()
	a.router = this.router
	a.hook = this.hook
	a.session = this.session
	return a
}
