package xgo

const (
	HookAfterInit           = "AfterInit"
	HookBeforeMethodGet     = "BeforeMethodGet"
	HookAfterMethodGet      = "AfterMethodGet"
	HookBeforeMethodPost    = "BeforeMethodPost"
	HookAfterMethodPost     = "AfterMethodPost"
	HookBeforeMethodHead    = "BeforeMethodHead"
	HookAfterMethodHead     = "AfterMethodHead"
	HookBeforeMethodDelete  = "BeforeMethodDelete"
	HookAfterMethodDelete   = "AfterMethodDelete"
	HookBeforeMethodPut     = "BeforeMethodPut"
	HookAfterMethodPut      = "AfterMethodPut"
	HookBeforeMethodPatch   = "BeforeMethodPatch"
	HookAfterMethodPatch    = "AfterMethodPatch"
	HookBeforeMethodOptions = "BeforeMethodOptions"
	HookAfterMethodOptions  = "AfterMethodOptions"
	HookBeforeRender        = "BeforeRender"
	HookAfterRender         = "AfterRender"
	HookBeforeOutput        = "BeforeOutput"
	HookAfterOutput         = "AfterOutput"
)

type xgoHook struct {
	app          *App
	handlerHooks map[string][]HookHandlerFunc
}

type HookHandler struct {
	Context  *Context
	Template *Template
	Session  *Session
}

type HookHandlerFunc func(*HookHandler)

func (this *xgoHook) AddHandlerHook(event string, hookFunc HookHandlerFunc) {
	if this.handlerHooks == nil {
		this.handlerHooks = make(map[string][]HookHandlerFunc)
	}
	if _, ok := this.handlerHooks[event]; !ok {
		this.handlerHooks[event] = []HookHandlerFunc{}
	}
	this.handlerHooks[event] = append(this.handlerHooks[event], hookFunc)
}

func (this *xgoHook) CallHandlerHook(event string, hc *HookHandler) {
	if funcList, ok := this.handlerHooks[event]; ok {
		for _, hookFunc := range funcList {
			hookFunc(hc)
			if hc.Context.response.Finished {
				return
			}
		}
	}
}
