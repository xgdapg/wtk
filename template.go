package xgo

import (
	"html/template"
	"io/ioutil"
)

const (
	tplRootName string = "__ROOT__"
)

type Template struct {
	tpl       *template.Template
	tplVars   map[string]interface{}
	tplResult *templateResult
}

func (this *Template) SetVar(name string, value interface{}) {
	this.tplVars[name] = value
}

func (this *Template) SetTemplate(file string) bool {
	this.tpl = template.New(tplRootName)
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return false
	}
	this.tpl.Parse(string(content))
	return true
}

func (this *Template) SetSubTemplate(name, file string) bool {
	if this.tpl == nil {
		return false
	}
	tpl := this.tpl.New(name)
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return false
	}
	tpl.Parse(`{{define "` + name + `"}}` + string(content) + `{{end}}`)
	return true
}

func (this *Template) Parse() bool {
	if this.tpl == nil {
		return false
	}
	if this.tplResult != nil {
		return false
	}
	this.tplResult = &templateResult{data: ""}
	err := this.tpl.Execute(this.tplResult, this.tplVars)
	if err != nil {
		return false
	}
	return true
}

func (this *Template) GetResult() string {
	if this.tplResult == nil {
		return ""
	}
	return this.tplResult.GetData()
}

type templateResult struct {
	data string
}

func (this *templateResult) Write(p []byte) (n int, err error) {
	this.data = this.data + string(p)
	return len(p), nil
}

func (this *templateResult) GetData() string {
	return this.data
}
