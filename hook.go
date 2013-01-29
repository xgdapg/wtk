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
	app             *xgoApp
	controllerHooks map[string][]HookControllerFunc
}

type HookController struct {
	Context  *xgoContext
	Template *xgoTemplate
	Session  *xgoSession
}

type HookControllerFunc func(*HookController)

func (this *xgoHook) AddControllerHook(event string, hookFunc HookControllerFunc) {
	if this.controllerHooks == nil {
		this.controllerHooks = make(map[string][]HookControllerFunc)
	}
	if _, ok := this.controllerHooks[event]; !ok {
		this.controllerHooks[event] = []HookControllerFunc{}
	}
	this.controllerHooks[event] = append(this.controllerHooks[event], hookFunc)
}

func (this *xgoHook) CallControllerHook(event string, hc *HookController) {
	if funcList, ok := this.controllerHooks[event]; ok {
		for _, hookFunc := range funcList {
			hookFunc(hc)
			if hc.Context.Response.Finished {
				return
			}
		}
	}
}
