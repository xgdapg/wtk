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
