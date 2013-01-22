package xgo

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strconv"
)

type xgoConfig struct {
	datas map[string]string
}

func (this *xgoConfig) GetConfig(key string) *xgoConfigValue {
	val := &xgoConfigValue{
		value: "",
		exist: false,
	}
	if data, exist := this.datas[key]; exist {
		val.value = data
		val.exist = true
	}
	return val
}

func (this *xgoConfig) LoadConfig(filename string) error {
	this.datas = make(map[string]string)

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := bufio.NewReader(file)
	for {
		line, _, err := buf.ReadLine()
		if err == io.EOF {
			break
		}
		if bytes.Equal(line, []byte{}) {
			continue
		}
		line = bytes.TrimSpace(line)
		if bytes.HasPrefix(line, []byte{'#'}) {
			continue
		}
		s := bytes.SplitN(line, []byte{'='}, 2)
		k := string(bytes.TrimSpace(s[0]))
		v := string(bytes.TrimSpace(s[1]))
		this.datas[k] = v
	}
	return nil
}

type xgoConfigValue struct {
	value string
	exist bool
}

func (this *xgoConfigValue) String() (string, bool) {
	if !this.exist {
		return "", false
	}
	return this.value, true
}

func (this *xgoConfigValue) Int() (int, bool) {
	if !this.exist {
		return 0, false
	}
	i, err := strconv.Atoi(this.value)
	if err != nil {
		return 0, false
	}
	return i, true
}

func (this *xgoConfigValue) Float64() (float64, bool) {
	if !this.exist {
		return 0, false
	}
	f, err := strconv.ParseFloat(this.value, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

func (this *xgoConfigValue) Bool() (bool, bool) {
	if !this.exist {
		return false, false
	}
	b, err := strconv.ParseBool(this.value)
	if err != nil {
		return false, false
	}
	return b, true
}
