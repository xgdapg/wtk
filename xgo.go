package xgo

import (
	"os"
)

var (
	app              *App
	util             Util
	ListenAddr       string = ""
	ListenPort       int    = 80
	RunMode          string = "http"
	EnableDaemon     bool   = false
	EnableStats      bool   = true
	EnableSession    bool   = true
	EnablePprof      bool   = true
	EnableAutoRender bool   = true
)

func init() {
	app = NewApp()
	util = Util{}
}

func NewApp() *App {
	return &App{
		router: &Router{},
	}
}

func RegisterController(pattern string, c ControllerInterface) {
	app.RegisterController(pattern, c)
}

func Run() {
	if EnableDaemon {
		util.CallMethod(&util, "SetDaemonMode", 1, 0)
	}
	app.Run(ListenAddr, ListenPort)
	os.Exit(0)
}

func LoadConfig() {

}

func GetConfig(key string) string {
	return ""
}
