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
	app          *xgoApp
	handlerHooks map[string][]HookHandlerFunc
}

type HookHandler struct {
	Context  *xgoContext
	Template *xgoTemplate
	Session  *xgoSession
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
			if hc.Context.Response.Finished {
				return
			}
		}
	}
}
