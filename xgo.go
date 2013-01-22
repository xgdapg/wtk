package xgo

var (
	app           *xgoApp
	util          xgoUtil
	ListenAddr    string = ""
	ListenPort    int    = 80
	RunMode       string = "http"
	EnableDaemon  bool   = false
	EnableStats   bool   = true
	EnableSession bool   = true
	EnablePprof   bool   = true
)

func init() {
	app = NewApp()
	util = xgoUtil{}
}

func NewApp() *xgoApp {
	return new(xgoApp).init()
}

func RegisterController(pattern string, c xgoControllerInterface) {
	app.RegisterController(pattern, c)
}

func AddControllerHook(event string, hookFunc xgoControllerHookFunc) {
	app.AddControllerHook(event, hookFunc)
}

func SetStaticPath(sPath, fPath string) {
	app.SetStaticPath(sPath, fPath)
}

func Run() {
	if EnableDaemon {
		util.CallMethod(&util, "SetDaemonMode", 1, 0)
	}
	app.Run(ListenAddr, ListenPort)
}

func LoadConfig() {

}

func GetConfig(key string) string {
	return ""
}
