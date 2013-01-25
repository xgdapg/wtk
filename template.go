package xgo

import (
	"html/template"
	"io/ioutil"
)

var tplFuncMap template.FuncMap

func init() {
	tplFuncMap = make(template.FuncMap)
}

func AddTemplateFunc(name string, tplFunc interface{}) {
	tplFuncMap[name] = tplFunc
}

type xgoTemplate struct {
	tpl       *template.Template
	tplVars   map[string]interface{}
	tplResult *xgoTemplateResult
}

func (this *xgoTemplate) SetVar(name string, value interface{}) {
	this.tplVars[name] = value
}

func (this *xgoTemplate) SetTemplateString(str string) bool {
	this.tpl = template.New("")
	this.tpl.Parse(str)
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
	tpl.Parse(`{{define "` + name + `"}}` + str + `{{end}}`)
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
	this.tplResult = &xgoTemplateResult{data: []byte{}}
	this.tpl.Funcs(tplFuncMap)
	err := this.tpl.Execute(this.tplResult, this.tplVars)
	if err != nil {
		return false
	}
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

type xgoTemplateResult struct {
	data []byte
}

func (this *xgoTemplateResult) Write(p []byte) (n int, err error) {
	this.data = append(this.data, p...)
	return len(p), nil
}

func (this *xgoTemplateResult) String() string {
	return string(this.data)
}

func (this *xgoTemplateResult) Bytes() []byte {
	return this.data
}
