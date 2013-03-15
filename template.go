package xgo

import (
	"html/template"
	"io/ioutil"
)

var tplFuncMap template.FuncMap
var tplVars map[string]interface{}

func init() {
	tplFuncMap = make(template.FuncMap)
	tplVars = make(map[string]interface{})
}

func AddTemplateFunc(name string, tplFunc interface{}) {
	tplFuncMap[name] = tplFunc
}

func SetTemplateVar(name string, value interface{}) {
	tplVars[name] = value
}

type xgoTemplate struct {
	hdlr      *Handler
	tpl       *template.Template
	Vars      map[string]interface{}
	tplResult *xgoTemplateResult
}

func (this *xgoTemplate) SetVar(name string, value interface{}) {
	this.Vars[name] = value
}

func (this *xgoTemplate) GetVar(name string) interface{} {
	if val, ok := this.Vars[name]; ok {
		return val
	}
	return nil
}

func (this *xgoTemplate) SetTemplateString(str string) bool {
	this.tpl = template.New("")
	this.tpl.Funcs(tplFuncMap).Parse(str)
	return true
}

func (this *xgoTemplate) SetTemplateFile(filename string) bool {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return false
	}
	return this.SetTemplateString(string(content))
}

func (this *xgoTemplate) SetSubTemplateString(name, str string) bool {
	if this.tpl == nil {
		return false
	}
	tpl := this.tpl.New(name)
	tpl.Funcs(tplFuncMap).Parse(`{{define "` + name + `"}}` + str + `{{end}}`)
	return true
}

func (this *xgoTemplate) SetSubTemplateFile(name, filename string) bool {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return false
	}
	return this.SetSubTemplateString(name, string(content))
}

func (this *xgoTemplate) Parse() bool {
	if this.tpl == nil {
		return false
	}
	if this.tplResult != nil {
		return false
	}

	hc := this.hdlr.getHookHandler()
	this.hdlr.app.callHandlerHook("BeforeRender", hc)
	if this.hdlr.Context.Response.Finished {
		return true
	}

	this.tplResult = &xgoTemplateResult{data: []byte{}}
	err := this.tpl.Execute(this.tplResult, this.Vars)
	if err != nil {
		return false
	}

	this.hdlr.app.callHandlerHook("AfterRender", hc)
	return true
}

func (this *xgoTemplate) GetResult() []byte {
	if this.tplResult == nil {
		return []byte{}
	}
	return this.tplResult.Bytes()
}

func (this xgoTemplate) GetResultString() string {
	return string(this.GetResult())
}

func (this *xgoTemplate) SetResult(p []byte) {
	if this.tplResult == nil {
		return
	}
	this.tplResult.SetBytes(p)
}

func (this *xgoTemplate) SetResultString(s string) {
	this.SetResult([]byte(s))
}

type xgoTemplateResult struct {
	data []byte
}

func (this *xgoTemplateResult) Write(p []byte) (n int, err error) {
	this.data = append(this.data, p...)
	return len(p), nil
}

func (this *xgoTemplateResult) SetBytes(p []byte) {
	this.data = p
}

func (this *xgoTemplateResult) String() string {
	return string(this.data)
}

func (this *xgoTemplateResult) Bytes() []byte {
	return this.data
}
