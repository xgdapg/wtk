package xgo

import (
	"os"
	"path/filepath"
)

var (
	app               *App
	apps              map[int]*App
	appIdGen          *xgoAutoIncr
	util              *xgoUtil
	cfg               *xgoConfig
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
	GzipMinLength     int
	GzipTypes         []string
	SslCertificate    string
	SslCertificateKey string
)

func init() {
	util = &xgoUtil{}
	AppRoot, err := os.Getwd()
	if err != nil {
		AppRoot = util.getDefaultRootPath()
	}
	defaultCfg := &xgoDefaultConfig{
		AppRoot:           AppRoot,
		ListenAddr:        "",
		ListenPort:        80,
		RunMode:           "http",
		EnableStats:       true,
		CookieSecret:      "foobar",
		SessionName:       "XGOSESSID",
		SessionTTL:        60 * 15,
		EnablePprof:       true,
		EnableGzip:        true,
		GzipMinLength:     1024,
		GzipTypes:         []string{"text", "javascript", "css", "xml"},
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

	cfg = &xgoConfig{}
	cfg.LoadFile(cfgFile)
	cfg.RegisterConfig(defaultCfg)
	apps = make(map[int]*App)
	appIdGen = newAutoIncr(1, 1)
	app = NewApp()
}

func NewApp() *App {
	id := appIdGen.Fetch()
	a := new(App).init(id)
	apps[id] = a
	return a
}

func GetMainApp() *App {
	return app
}

func AddRoutingRule(pattern string, c HandlerInterface) {
	app.AddRoutingRule(pattern, c)
}

func RemoveRoutingRule(pattern string) {
	app.RemoveRoutingRule(pattern)
}

func AddHandlerHook(event string, hookFunc HookHandlerFunc) {
	app.AddHandlerHook(event, hookFunc)
}

func AddStaticPath(sPath, fPath string) {
	app.AddStaticPath(sPath, fPath)
}

func RemoveStaticPath(sPath string) {
	app.RemoveStaticPath(sPath)
}

func AddStaticFileType(ext ...string) {
	app.AddStaticFileType(ext...)
}

func RemoveStaticFileType(ext ...string) {
	app.RemoveStaticFileType(ext...)
}

func RegisterSessionStorage(storage SessionStorageInterface) {
	app.RegisterSessionStorage(storage)
}

func RegisterCustomHttpStatus(code int, filePath string) {
	app.RegisterCustomHttpStatus(code, filePath)
}

func Run() error {
	return app.Run(RunMode, ListenAddr, ListenPort)
}

func LoadConfig(conf interface{}) error {
	return cfg.RegisterConfig(conf)
}

func ReloadConfig() error {
	return cfg.ReloadFile()
}
