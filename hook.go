package wtk

const (
	HookReceiveRequest      = "ReceiveRequest"
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

type wtkHook struct {
	server       *Server
	handlerHooks map[string][]HookHandlerFunc
}

type HookHandler struct {
	server   *Server
	Context  *Context
	Template *Template
	Session  *Session
}

func (this *HookHandler) Render() {
	this.Template.Parse()
}

func (this *HookHandler) Output() {
	content := this.Template.GetResult()
	if len(content) > 0 {
		this.Context.WriteBytes(content)
	}
}

func (this *HookHandler) GetServer() *Server {
	return this.server
}

type HookHandlerFunc func(*HookHandler)

func (this *wtkHook) AddHandlerHook(event string, hookFunc HookHandlerFunc) {
	if this.handlerHooks == nil {
		this.handlerHooks = make(map[string][]HookHandlerFunc)
	}
	if _, ok := this.handlerHooks[event]; !ok {
		this.handlerHooks[event] = []HookHandlerFunc{}
	}
	this.handlerHooks[event] = append(this.handlerHooks[event], hookFunc)
}

func (this *wtkHook) CallHandlerHook(event string, hc *HookHandler) {
	if funcList, ok := this.handlerHooks[event]; ok {
		for _, hookFunc := range funcList {
			hookFunc(hc)
			if hc.Context.response.Finished {
				return
			}
		}
	}
}
