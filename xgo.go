package xgo

var (
	app           *App
	util          Util
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
	util = Util{}
}

func NewApp() *App {
	return new(App).init()
}

func RegisterController(pattern string, c ControllerInterface) {
	app.RegisterController(pattern, c)
}

func AddControllerHook(event string, hookFunc controllerHookFunc) {
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
