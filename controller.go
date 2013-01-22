package xgo

import (
	"net/http"
)

type xgoControllerInterface interface {
	Init(*xgoContext, *xgoTemplate, *xgoSession, string)
	Get()
	Post()
	Delete()
	Put()
	Head()
	Patch()
	Options()
	Render()
	Output()
}

type Controller struct {
	Context  *xgoContext
	Template *xgoTemplate
	Session  *xgoSession
}

func (this *Controller) Init(ctx *xgoContext, tpl *xgoTemplate, sess *xgoSession, cn string) {
	this.Context = ctx
	this.Template = tpl
	this.Session = sess
}

func (this *Controller) Get() {
	http.Error(this.Context.Response, "Method Not Allowed", 405)
}

func (this *Controller) Post() {
	http.Error(this.Context.Response, "Method Not Allowed", 405)
}

func (this *Controller) Delete() {
	http.Error(this.Context.Response, "Method Not Allowed", 405)
}

func (this *Controller) Put() {
	http.Error(this.Context.Response, "Method Not Allowed", 405)
}

func (this *Controller) Head() {
	http.Error(this.Context.Response, "Method Not Allowed", 405)
}

func (this *Controller) Patch() {
	http.Error(this.Context.Response, "Method Not Allowed", 405)
}

func (this *Controller) Options() {
	http.Error(this.Context.Response, "Method Not Allowed", 405)
}

func (this *Controller) Render() {
	this.Template.Parse()
}

func (this *Controller) Output() {
	this.Context.WriteString(this.Template.GetResult())
}

type xgoControllerHook struct {
	app   *xgoApp
	hooks []*xgoControllerHookData
}

func (this *xgoControllerHook) AddHook(event string, hookFunc xgoControllerHookFunc) {
	data := &xgoControllerHookData{
		Event: event,
		Func:  hookFunc,
	}
	this.hooks = append(this.hooks, data)

}

func (this *xgoControllerHook) CallHook(event string, url string, hc *HookController) {
	for _, hook := range this.hooks {
		if hook.Event == event {
			hook.Func(url, hc)
			if hc.Context.Response.(*xgoResponseWriter).HasOutput {
				return
			}
		}
	}
}

type xgoControllerHookFunc func(string, *HookController)

type xgoControllerHookData struct {
	Event string
	Func  xgoControllerHookFunc
}

type HookController struct {
	Context  *xgoContext
	Template *xgoTemplate
	Session  *xgoSession
}
