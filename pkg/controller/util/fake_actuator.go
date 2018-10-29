package util

import (
	"k8s.io/apimachinery/pkg/runtime"
)

type FakeActuator interface {
	CallCount() int
	Called(string) bool

	OnCreate(error)
	OnUpdate(error)
	OnDelete(error)
	OnExists(bool, error)

	Create(runtime.Object) error
	Delete(runtime.Object) error
	Update(runtime.Object) error
	Exists(runtime.Object) (bool, error)
}

func NewFakeActuator() FakeActuator {
	return &fakeActuator{}
}

type methodCall struct {
	method string
}

type fakeActuator struct {
	calls        []methodCall
	createErr    error
	updateErr    error
	deleteErr    error
	existsErr    error
	existsResult bool
}

var _ FakeActuator = &fakeActuator{}

func (a *fakeActuator) Create(obj runtime.Object) error {
	a.calls = append(a.calls, methodCall{method: "Create"})
	return a.createErr
}

func (a *fakeActuator) Delete(obj runtime.Object) error {
	a.calls = append(a.calls, methodCall{method: "Delete"})
	return a.deleteErr
}

func (a *fakeActuator) Update(obj runtime.Object) error {
	a.calls = append(a.calls, methodCall{method: "Update"})
	return a.updateErr
}

func (a *fakeActuator) Exists(obj runtime.Object) (bool, error) {
	a.calls = append(a.calls, methodCall{method: "Exists"})
	return a.existsResult, a.existsErr
}

func (a *fakeActuator) CallCount() int {
	return len(a.calls)
}

func (a *fakeActuator) Called(method string) bool {
	for _, call := range a.calls {
		if call.method == method {
			return true
		}
	}
	return false
}

func (a *fakeActuator) OnCreate(err error) {
	a.createErr = err
}

func (a *fakeActuator) OnUpdate(err error) {
	a.updateErr = err
}

func (a *fakeActuator) OnDelete(err error) {
	a.deleteErr = err
}

func (a *fakeActuator) OnExists(result bool, err error) {
	a.existsResult = result
	a.existsErr = err
}
