package xgo

import (
	"reflect"
)

var (
	app              *App
	util             *Util
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
	app := &App{}
	util := &Util{}
}

func RegisterController(pattern string, c ControllerInterface) {
	app.RegisterController(pattern, c)
}

func Run() {
	if EnableDaemon {
		method := reflect.New(reflect.TypeOf(*util)).MethodByName("SetDaemonMode")
		if method.Kind() == reflect.Func {
			in := make([]reflect.Value, 2)
			in[0] = reflect.ValueOf(1)
			in[1] = reflect.ValueOf(0)
			method.Call(in)
		}
	}
	app.Run(ListenAddr, ListenPort)
}

func LoadConfig() {

}

func GetConfig(key string) string {
	return ""
}
