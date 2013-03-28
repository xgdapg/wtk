package xgo

import (
	"fmt"
	"time"
)

type SessionStorageInterface interface {
	Init(int64)
	CreateSessionID() string
	Set(string, map[string]string)
	Get(string) map[string]string
	Delete(string)
}

type xgoSessionManager struct {
	sessionStorage SessionStorageInterface
	inited         bool
}

func (this *xgoSessionManager) RegisterStorage(storage SessionStorageInterface) {
	if storage == nil {
		return
	}
	this.sessionStorage = storage
	this.inited = false
}

func (this *xgoSessionManager) checkInit() {
	if !this.inited {
		this.sessionStorage.Init(SessionTTL)
		this.inited = true
	}
}

func (this *xgoSessionManager) CreateSessionID() string {
	this.checkInit()
	return this.sessionStorage.CreateSessionID()
}

func (this *xgoSessionManager) Set(sid string, data map[string]string) {
	this.checkInit()
	this.sessionStorage.Set(sid, data)
}

func (this *xgoSessionManager) Get(sid string) map[string]string {
	this.checkInit()
	return this.sessionStorage.Get(sid)
}

func (this *xgoSessionManager) Delete(sid string) {
	this.checkInit()
	this.sessionStorage.Delete(sid)
}

type Session struct {
	hdlr           *Handler
	sessionManager *xgoSessionManager
	sessionId      string
	ctx            *Context
	data           map[string]string
}

func (this *Session) init() {
	if this.sessionId == "" {
		this.sessionId = this.sessionManager.CreateSessionID()
		this.ctx.SetSecureCookie(SessionName, this.sessionId, 0)
	}
	if this.data == nil {
		this.data = this.sessionManager.Get(this.sessionId)
	}
}

func (this *Session) Get(key string) string {
	this.init()
	if data, exist := this.data[key]; exist {
		return data
	}
	return ""
}

func (this *Session) Set(key string, data string) {
	this.init()
	this.data[key] = data
	this.sessionManager.Set(this.sessionId, this.data)
}

func (this *Session) Delete(key string) {
	this.init()
	delete(this.data, key)
	this.sessionManager.Set(this.sessionId, this.data)
}

type xgoDefaultSessionStorage struct {
	ttl   int64
	datas map[string]xgoDefaultSessionStorageData
	incr  *xgoAutoIncr
}

type xgoDefaultSessionStorageData struct {
	expires int64
	data    map[string]string
}

func (this *xgoDefaultSessionStorage) Init(ttl int64) {
	if this.datas != nil {
		return
	}
	this.ttl = ttl
	this.datas = make(map[string]xgoDefaultSessionStorageData)
	go this.gc()
	this.incr = newAutoIncr(1, 1)
}

func (this *xgoDefaultSessionStorage) gc() {
	for {
		if len(this.datas) > 0 {
			now := time.Now().Unix()
			for sid, data := range this.datas {
				if data.expires <= now {
					delete(this.datas, sid)
				}
			}
		}
		time.Sleep(time.Second)
	}
}

func (this *xgoDefaultSessionStorage) CreateSessionID() string {
	t := time.Now()
	return "SESS" + fmt.Sprintf("%d%d", t.Unix(), this.incr.Fetch())
}

func (this *xgoDefaultSessionStorage) Set(sid string, data map[string]string) {
	d := xgoDefaultSessionStorageData{
		expires: time.Now().Unix() + this.ttl,
		data:    data,
	}
	this.datas[sid] = d
}

func (this *xgoDefaultSessionStorage) Get(sid string) map[string]string {
	if data, exist := this.datas[sid]; exist {
		data.expires = time.Now().Unix() + this.ttl
		return data.data
	}
	return make(map[string]string)
}

func (this *xgoDefaultSessionStorage) Delete(sid string) {
	delete(this.datas, sid)
}
