package xgo

import (
	"html/template"
	"io/ioutil"
)

type xgoTemplate struct {
	tpl       *template.Template
	tplVars   map[string]interface{}
	tplResult *xgoTemplateResult
}

func (this *xgoTemplate) SetVar(name string, value interface{}) {
	this.tplVars[name] = value
}

func (this *xgoTemplate) SetTemplate(file string) bool {
	this.tpl = template.New("")
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return false
	}
	this.tpl.Parse(string(content))
	return true
}

func (this *xgoTemplate) SetSubTemplate(name, file string) bool {
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

func (this *xgoTemplate) Parse() bool {
	if this.tpl == nil {
		return false
	}
	if this.tplResult != nil {
		return false
	}
	this.tplResult = &xgoTemplateResult{data: ""}
	err := this.tpl.Execute(this.tplResult, this.tplVars)
	if err != nil {
		return false
	}
	return true
}

func (this *xgoTemplate) GetResult() string {
	if this.tplResult == nil {
		return ""
	}
	return this.tplResult.GetData()
}

type xgoTemplateResult struct {
	data string
}

func (this *xgoTemplateResult) Write(p []byte) (n int, err error) {
	this.data = this.data + string(p)
	return len(p), nil
}

func (this *xgoTemplateResult) GetData() string {
	return this.data
}
