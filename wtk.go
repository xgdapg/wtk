package wtk

import (
	"os"
	"path/filepath"
)

var (
	server            *Server
	servers           map[int]*Server
	serverIdGen       *wtkAutoIncr
	util              *wtkUtil
	cfg               *wtkConfig
	cfgFile           string
	AppRoot           string
	ListenAddr        string
	ListenPort        int
	RunMode           string
	EnableStats       bool
	CookieSecret      string
	SessionName       string
	SessionTTL        int64
	EnablePprof       bool
	EnableGzip        bool
	EnableRouteCache  bool
	GzipMinLength     int
	GzipTypes         []string
	SslCertificate    string
	SslCertificateKey string
)

func init() {
	util = &wtkUtil{}
	AppRoot, err := os.Getwd()
	if err != nil {
		AppRoot = util.getDefaultRootPath()
	}
	defaultCfg := &wtkDefaultConfig{
		AppRoot:           AppRoot,
		ListenAddr:        "",
		ListenPort:        80,
		RunMode:           "http",
		EnableStats:       true,
		CookieSecret:      "foobar",
		SessionName:       "wtkSESSID",
		SessionTTL:        60 * 15,
		EnablePprof:       true,
		EnableGzip:        true,
		EnableRouteCache:  true,
		GzipMinLength:     1024,
		GzipTypes:         []string{"text", "js", "css", "xml"},
		SslCertificate:    "",
		SslCertificateKey: "",
	}

	cfgFile = filepath.Join(AppRoot, "app.conf")
	// Check the first argument of cmd line,
	// if it is not a flag (begin with '-'),
	// try to use it as the config file path.
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg[0] != '-' {
			cfgFile = arg
		}
	}

	cfg = &wtkConfig{}
	cfg.LoadFile(cfgFile)
	cfg.RegisterConfig(defaultCfg)
	servers = make(map[int]*Server)
	serverIdGen = newAutoIncr(1, 1)
	server = NewServer()
}

func NewServer() *Server {
	id := serverIdGen.Fetch()
	a := new(Server).init(id)
	servers[id] = a
	return a
}

func MainServer() *Server {
	return server
}

func AddRoute(pattern string, c HandlerInterface) *Route {
	return server.AddRoute(pattern, c)
}

func RemoveRoute(pattern string) {
	server.RemoveRoute(pattern)
}

func SetPrefixPath(prefix string) {
	server.SetPrefixPath(prefix)
}

func AddHandlerHook(event string, hookFunc HookHandlerFunc) {
	server.AddHandlerHook(event, hookFunc)
}

func AddStaticFileDir(dirs ...string) {
	server.AddStaticFileDir(dirs...)
}

func RemoveStaticFileDir(dirs ...string) {
	server.RemoveStaticFileDir(dirs...)
}

func AddStaticFileType(ext ...string) {
	server.AddStaticFileType(ext...)
}

func RemoveStaticFileType(ext ...string) {
	server.RemoveStaticFileType(ext...)
}

func RegisterSessionStorage(storage SessionStorageInterface) {
	server.RegisterSessionStorage(storage)
}

func RegisterCustomHttpStatus(code int, filePath string) {
	server.RegisterCustomHttpStatus(code, filePath)
}

func Run() error {
	return server.Run(RunMode, ListenAddr, ListenPort)
}

func LoadConfig(conf interface{}) error {
	return cfg.RegisterConfig(conf)
}

func ReloadConfig() error {
	return cfg.ReloadFile()
}
