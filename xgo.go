package xgo

var (
	app          *xgoApp
	util         xgoUtil
	ListenAddr   string = ""
	ListenPort   int    = 80
	RunMode      string = "http"
	EnableDaemon bool   = false
	EnableStats  bool   = true
	SessionName  string = "XGOSESSID"
	SessionTTL   int64  = 60 * 15
	EnablePprof  bool   = true
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

func RegisterControllerHook(event string, hookFunc xgoControllerHookFunc) {
	app.RegisterControllerHook(event, hookFunc)
}

func SetStaticPath(sPath, fPath string) {
	app.SetStaticPath(sPath, fPath)
}

func RegisterSessionStorage(storage xgoSessionStorageInterface) {
	app.RegisterSessionStorage(storage)
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
