package xgo

import (
	"reflect"
)

type xgoUtil struct {
}

func (this xgoUtil) CallMethod(i interface{}, name string, args ...interface{}) bool {
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
