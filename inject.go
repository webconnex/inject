// Copyright 2013 Webconnex, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package inject

import (
	"fmt"
	"reflect"
)

type Mapper interface {
	Get(v interface{})
	GetValue(typ reflect.Type) reflect.Value
	Map(v interface{})
	MapTo(v, iface interface{})
	SetValue(reflect.Type, reflect.Value)
}

type NamedMapper interface {
	GetNamed(v interface{}, name string)
	GetNamedValue(name string) reflect.Value
	MapNamed(v interface{}, name string)
	SetNamedValue(name string, val reflect.Value)
}

type Invoker interface {
	Invoke(fn interface{}) []interface{}
}

type NamedInvoker interface {
	InvokeNamed(fn interface{}, args ...string) []interface{}
}

type Injector interface {
	Mapper
	Invoker
	NamedMapper
	NamedInvoker
}

type injector struct {
	parent      Injector
	values      map[reflect.Type]reflect.Value
	namedValues map[string]reflect.Value
}

func New(parent Injector) Injector {
	return &injector{
		parent:      parent,
		values:      make(map[reflect.Type]reflect.Value),
		namedValues: make(map[string]reflect.Value),
	}
}

func (j *injector) Get(v interface{}) {
	dest := reflect.ValueOf(v).Elem()
	val := j.GetValue(dest.Type())
	if val.IsValid() {
		dest.Set(val)
	}
}

func (j *injector) GetValue(typ reflect.Type) reflect.Value {
	val, found := j.values[typ]
	if !found && j.parent != nil {
		val = j.parent.GetValue(typ)
	}
	return val
}

func (j *injector) Map(v interface{}) {
	j.values[reflect.TypeOf(v)] = reflect.ValueOf(v)
}

func (j *injector) MapTo(v, iface interface{}) {
	typ := reflect.TypeOf(iface)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Interface {
		panic(fmt.Sprintf("inject: MapTo expecting interface, got %s", typ.Kind().String()))
	}
	j.values[typ] = reflect.ValueOf(v)
}

func (j *injector) SetValue(typ reflect.Type, val reflect.Value) {
	j.values[typ] = val
}

func (j *injector) Invoke(fn interface{}) []interface{} {
	fval := reflect.ValueOf(fn)
	ftyp := fval.Type()
	if ftyp.Kind() != reflect.Func {
		panic("inject: non-func type")
	}
	in := make([]reflect.Value, ftyp.NumIn())
	for i := range in {
		typ := ftyp.In(i)
		val := j.GetValue(typ)
		if !val.IsValid() {
			panic(fmt.Sprintf("inject: missing %s", typ))
		}
		in[i] = val
	}
	out := make([]interface{}, ftyp.NumOut())
	for i, val := range fval.Call(in) {
		out[i] = val.Interface()
	}
	return out
}

func (j *injector) GetNamed(v interface{}, name string) {
	dest := reflect.ValueOf(v).Elem()
	val := j.GetNamedValue(name)
	if val.IsValid() {
		dest.Set(val)
	}
}

func (j *injector) GetNamedValue(name string) reflect.Value {
	val, found := j.namedValues[name]
	if !found && j.parent != nil {
		val = j.parent.GetNamedValue(name)
	}
	return val
}

func (j *injector) MapNamed(v interface{}, name string) {
	j.namedValues[name] = reflect.ValueOf(v)
}

func (j *injector) SetNamedValue(name string, val reflect.Value) {
	j.namedValues[name] = val
}

func simpleConvert(src reflect.Value, typ reflect.Type) reflect.Value {
	dst := reflect.New(typ).Elem()
	switch typ.Kind() {
	case reflect.String:
		dst.SetString(src.String())
	case reflect.Slice:
		if typ.Elem().Kind() == reflect.Uint8 {
			dst.SetBytes(src.Bytes())
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		dst.SetInt(src.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		dst.SetInt(src.Int())
	case reflect.Float32, reflect.Float64:
		dst.SetFloat(src.Float())
	default:
		panic(fmt.Sprintf("cannot convert %s to %s", src.Type(), typ))
	}
	return dst
}

func (j *injector) InvokeNamed(fn interface{}, names ...string) []interface{} {
	fval := reflect.ValueOf(fn)
	ftyp := fval.Type()
	if ftyp.Kind() != reflect.Func {
		panic("inject: non-func type")
	}
	count := ftyp.NumIn()
	if len(names) != count {
		panic(fmt.Sprintf("inject: expecting %d names, got %d", count, len(names)))
	}
	in := make([]reflect.Value, count)
	for i := range in {
		typ := ftyp.In(i)
		var val reflect.Value
		if name := names[i]; len(name) > 0 {
			val = j.GetNamedValue(name)
			if val.Type() != typ {
				val = simpleConvert(val, typ)
			}
		} else {
			val = j.GetValue(typ)
		}
		if !val.IsValid() {
			panic(fmt.Sprintf("inject: missing %s", typ))
		}
		in[i] = val
	}
	out := make([]interface{}, ftyp.NumOut())
	for i, val := range fval.Call(in) {
		out[i] = val.Interface()
	}
	return out
}
