package wtk

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
)

type wtkUtil struct{}
type wtkOsUtil struct{}

func (this *wtkUtil) CallMethod(i interface{}, name string, args ...interface{}) bool {
	t := reflect.TypeOf(i)
	if t.Kind() != reflect.Ptr {
		return false
	}
	argc := len(args)
	method := reflect.ValueOf(i).MethodByName(name)
	if method.Kind() == reflect.Func {
		in := make([]reflect.Value, argc)
		for j, arg := range args {
			in[j] = reflect.ValueOf(arg)
		}
		method.Call(in)
		return true
	}
	return false
}

func (this *wtkUtil) getCookieSig(secret, text string) string {
	hm := hmac.New(sha1.New, []byte(secret))
	hm.Write([]byte(text))
	hex := fmt.Sprintf("%02x", hm.Sum(nil))
	return hex
}
func (this *wtkUtil) AesEncrypt(secret, text []byte) []byte {
	h := sha256.New()
	h.Write(secret)
	key := h.Sum(nil)
	block, _ := aes.NewCipher(key)
	stream := cipher.NewCTR(block, key[:block.BlockSize()])
	stream.XORKeyStream(text, text)
	return text
}

func (this *wtkUtil) AesDecrypt(secret, text []byte) []byte {
	return this.AesEncrypt(secret, text)
}
func (this *wtkUtil) GetAppPath() (string, error) {
	cmd := os.Args[0]
	p, err := filepath.Abs(cmd)
	if err != nil {
		return "", err
	}
	s, err := os.Stat(p)
	if err == nil && !s.IsDir() {
		return p, nil
	}
	if cmd == filepath.Base(cmd) {
		f, err := exec.LookPath(cmd)
		if err != nil {
			return "", err
		}
		return f, nil
	}
	return "", os.ErrNotExist
}

func (this *wtkUtil) getDefaultRootPath() string {
	p, err := this.GetAppPath()
	if err != nil {
		return "./"
	}
	return filepath.Dir(p)
}

type wtkAutoIncr struct {
	start, step int
	queue       chan int
	running     bool
}

func newAutoIncr(start, step int) (ai *wtkAutoIncr) {
	ai = &wtkAutoIncr{
		start:   start,
		step:    step,
		running: true,
		queue:   make(chan int, 4),
	}
	go ai.process()
	return
}

func (ai *wtkAutoIncr) process() {
	defer func() { recover() }()
	for i := ai.start; ai.running; i = i + ai.step {
		ai.queue <- i
	}
}

func (ai *wtkAutoIncr) Fetch() int {
	return <-ai.queue
}

func (ai *wtkAutoIncr) Close() {
	ai.running = false
	close(ai.queue)
}
