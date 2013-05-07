package wtk

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

type wtkConfig struct {
	file string
	data []byte
	cfgs []interface{}
}

func (this *wtkConfig) LoadFile(filename string) error {
	var err error
	this.data, err = ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	this.file = filename
	for _, cfg := range this.cfgs {
		this.loadConfig(cfg)
	}
	return nil
}

func (this *wtkConfig) ReloadFile() error {
	return this.LoadFile(this.file)
}

func (this *wtkConfig) RegisterConfig(cfg interface{}) error {
	this.cfgs = append(this.cfgs, cfg)
	return this.loadConfig(cfg)
}

func (this *wtkConfig) loadConfig(cfg interface{}) error {
	if len(this.data) > 0 {
		err := json.Unmarshal(this.data, cfg)
		if err != nil {
			return err
		}
		util.CallMethod(cfg, "OnLoaded")
		return nil
	}
	util.CallMethod(cfg, "OnLoaded")
	return errors.New("config file is not loaded")
}

type wtkDefaultConfig struct {
	AppRoot           string
	ListenAddr        string
	ListenPort        int
	RunMode           string
	EnableStats       bool
	CookieSecret      string
	SessionName       string
	SessionTTL        int64
	EnablePprof       bool
	EnableGzip        bool
	EnableRouteCache  bool
	GzipMinLength     int
	GzipTypes         []string
	SslCertificate    string
	SslCertificateKey string
}

func (this *wtkDefaultConfig) OnLoaded() {
	AppRoot = this.AppRoot
	ListenAddr = this.ListenAddr
	ListenPort = this.ListenPort
	RunMode = this.RunMode
	EnableStats = this.EnableStats
	CookieSecret = this.CookieSecret
	SessionName = this.SessionName
	SessionTTL = this.SessionTTL
	EnablePprof = this.EnablePprof
	EnableGzip = this.EnableGzip
	EnableRouteCache = this.EnableRouteCache
	GzipMinLength = this.GzipMinLength
	GzipTypes = this.GzipTypes
	SslCertificate = this.SslCertificate
	SslCertificateKey = this.SslCertificateKey
}
