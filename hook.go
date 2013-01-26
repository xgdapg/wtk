package xgo

const (
	HookEventAfterInit           = "AfterInit"
	HookEventBeforeMethodGet     = "BeforeMethodGet"
	HookEventAfterMethodGet      = "AfterMethodGet"
	HookEventBeforeMethodPost    = "BeforeMethodPost"
	HookEventAfterMethodPost     = "AfterMethodPost"
	HookEventBeforeMethodHead    = "BeforeMethodHead"
	HookEventAfterMethodHead     = "AfterMethodHead"
	HookEventBeforeMethodDelete  = "BeforeMethodDelete"
	HookEventAfterMethodDelete   = "AfterMethodDelete"
	HookEventBeforeMethodPut     = "BeforeMethodPut"
	HookEventAfterMethodPut      = "AfterMethodPut"
	HookEventBeforeMethodPatch   = "BeforeMethodPatch"
	HookEventAfterMethodPatch    = "AfterMethodPatch"
	HookEventBeforeMethodOptions = "BeforeMethodOptions"
	HookEventAfterMethodOptions  = "AfterMethodOptions"
	HookEventBeforeRender        = "BeforeRender"
	HookEventAfterRender         = "AfterRender"
	HookEventBeforeOutput        = "BeforeOutput"
	HookEventAfterOutput         = "AfterOutput"
)

type xgoHook struct {
	app             *xgoApp
	controllerHooks map[string][]HookControllerFunc
}

type HookController struct {
	Context  *xgoContext
	Template *xgoTemplate
	Session  *xgoSession
}

type HookControllerFunc func(string, *HookController)

func (this *xgoHook) AddControllerHook(event string, hookFunc HookControllerFunc) {
	if this.controllerHooks == nil {
		this.controllerHooks = make(map[string][]HookControllerFunc)
	}
	if _, ok := this.controllerHooks[event]; !ok {
		this.controllerHooks[event] = []HookControllerFunc{}
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
