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
	Map(v interface{}, ifaces ...interface{})
	MapValue(val reflect.Value, typ reflect.Type)
}

type Invoker interface {
	Invoke(fn interface{}) []interface{}
}

type Injector interface {
	Mapper
	Invoker
}

type injector struct {
	parent Injector
	values map[reflect.Type]reflect.Value
}

func New(parent Injector) Injector {
	return &injector{
		parent: parent,
		values: make(map[reflect.Type]reflect.Value),
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

func (j *injector) Map(v interface{}, ifaces ...interface{}) {
	if ifaces != nil {
		for _, iface := range ifaces {
			typ := reflect.TypeOf(iface)
			if typ.Kind() == reflect.Ptr {
				if subtyp := typ.Elem(); subtyp.Kind() == reflect.Interface {
					typ = typ.Elem()
				}
			}
			j.MapValue(reflect.ValueOf(v), typ)
		}
	} else {
		j.MapValue(reflect.ValueOf(v), reflect.TypeOf(v))
	}
}

func (j *injector) MapValue(val reflect.Value, typ reflect.Type) {
	j.values[typ] = val
}

func (j *injector) Invoke(fn interface{}) []interface{} {
	if fn == nil {
		return nil
	}
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
