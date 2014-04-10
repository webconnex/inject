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
	"github.com/webconnex/tester"
	"reflect"
	"testing"
)

type Foo struct {
	Value string
}

func TestGet(t *testing.T) {
	injector := New(nil).(*injector)
	value := Foo{Value: "Foo"}
	injector.values[reflect.TypeOf(value)] = reflect.ValueOf(value)

	var foo Foo
	injector.Get(&foo)
	if e := tester.DeepEquals(foo, value); e != nil {
		t.Error(e)
	}

	var fooy *Foo
	injector.Get(&fooy)
	if e := tester.DeepEquals(fooy, (*Foo)(nil)); e != nil {
		t.Error(e)
	}
}

func TestMap(t *testing.T) {
	injector := New(nil).(*injector)

	if e := tester.HasLen(injector.values, 0); e != nil {
		t.Error(e)
	}

	injector.Map(Foo{})
	if e := tester.HasLen(injector.values, 1); e != nil {
		t.Error(e)
	}

	injector.MapTo(Foo{}, (*interface{})(nil))
	if e := tester.HasLen(injector.values, 2); e != nil {
		t.Error(e)
	}
}

func TestGetParent(t *testing.T) {
	parent := New(nil).(*injector)
	value := Foo{Value: "Foo"}
	parent.values[reflect.TypeOf(value)] = reflect.ValueOf(value)

	injector := New(parent).(*injector)

	var foo Foo
	injector.Get(&foo)
	if e := tester.DeepEquals(foo, value); e != nil {
		t.Error(e)
	}

	value2 := Foo{Value: "Bar"}
	injector.values[reflect.TypeOf(value2)] = reflect.ValueOf(value2)

	var foo2 Foo
	injector.Get(&foo2)
	if e := tester.DeepEquals(foo2, value2); e != nil {
		t.Error(e)
	}

	var foo3 Foo
	parent.Get(&foo3)
	if e := tester.DeepEquals(foo3, value); e != nil {
		t.Error(e)
	}
}

func TestInvoke(t *testing.T) {
	injector := New(nil).(*injector)
	value := Foo{Value: "Foo"}
	injector.values[reflect.TypeOf(value)] = reflect.ValueOf(value)

	values := injector.Invoke(func(foo Foo) Foo {
		return foo
	})

	if e := tester.HasLen(injector.values, 1); e != nil {
		t.Fatal(e)
	}

	if e := tester.DeepEquals(values[0], value); e != nil {
		t.Error(e)
	}
}

func TestInvokePanicNonFunc(t *testing.T) {
	injector := New(nil).(*injector)
	defer func() {
		err := recover()
		if err == nil {
			t.Fail()
		}
	}()
	injector.Invoke(42)
}
