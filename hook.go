package xgo

type xgoHook struct {
	app             *xgoApp
	controllerHooks map[string][]xgoControllerHookFunc
}

type HookController struct {
	Context  *xgoContext
	Template *xgoTemplate
	Session  *xgoSession
}

type xgoControllerHookFunc func(string, *HookController)

func (this *xgoHook) AddControllerHook(event string, hookFunc xgoControllerHookFunc) {
	if this.controllerHooks == nil {
		this.controllerHooks = make(map[string][]xgoControllerHookFunc)
	}
	if _, ok := this.controllerHooks[event]; !ok {
		this.controllerHooks[event] = []xgoControllerHookFunc{}
	}
	this.controllerHooks[event] = append(this.controllerHooks[event], hookFunc)
}

func (this *xgoHook) CallControllerHook(event string, url string, hc *HookController) {
	if funcList, ok := this.controllerHooks[event]; ok {
		for _, hookFunc := range funcList {
			hookFunc(url, hc)
			if hc.Context.Response.(*xgoResponseWriter).HasOutput {
				return
			}
		}
	}
}
