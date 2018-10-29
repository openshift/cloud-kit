/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dnsrecord

import (
	cloudkitv1 "github.com/openshift/cloud-kit/pkg/apis/cloudkit/v1alpha1"
	"github.com/openshift/cloud-kit/pkg/controller/util"
)

type fakeActuator struct {
	util.FakeActuator
}

func (a *fakeActuator) Create(r *cloudkitv1.DNSRecord) error {
	return a.FakeActuator.Create(r)
}
func (a *fakeActuator) Delete(r *cloudkitv1.DNSRecord) error {
	return a.FakeActuator.Delete(r)
}
func (a *fakeActuator) Update(r *cloudkitv1.DNSRecord) error {
	return a.FakeActuator.Update(r)
}
func (a *fakeActuator) Exists(r *cloudkitv1.DNSRecord) (bool, error) {
	return a.FakeActuator.Exists(r)
}

func NewFakeActuator() (Actuator, util.FakeActuator) {
	a := &fakeActuator{
		FakeActuator: util.NewFakeActuator(),
	}
	return a, a.FakeActuator
}
