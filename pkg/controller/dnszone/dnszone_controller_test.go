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

package dnszone

import (
	"context"
	"errors"
	"testing"

	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/openshift/cloud-kit/pkg/apis"
	cloudkitv1 "github.com/openshift/cloud-kit/pkg/apis/cloudkit/v1alpha1"
	"github.com/openshift/cloud-kit/pkg/controller/util"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

const (
	testName      = "dnszone"
	testNamespace = "testns"
)

func TestDNSZoneReconcile(t *testing.T) {
	apis.AddToScheme(scheme.Scheme)

	// Utility function to get the test DNS zone from the fake client
	getZone := func(c client.Client) *cloudkitv1.DNSZone {
		zone := &cloudkitv1.DNSZone{}
		err := c.Get(context.TODO(), client.ObjectKey{Name: testName, Namespace: testNamespace}, zone)
		if err == nil {
			return zone
		}
		return nil
	}

	tests := []struct {
		name          string
		dnsZone       *cloudkitv1.DNSZone
		actuator      Actuator
		expectErr     bool
		setupActuator func(util.FakeActuator)
		validate      func(client.Client, util.FakeActuator)
	}{
		{
			name:    "Add finalizer",
			dnsZone: testDNSZoneWithoutFinalizer(),
			validate: func(c client.Client, a util.FakeActuator) {
				r := getZone(c)
				if r == nil || !util.HasFinalizer(r, cloudkitv1.DNSZoneFinalizer) {
					t.Errorf("did not get expected dnszone finalizer")
				}
			},
		},
		{
			name:    "Deleted with no finalizer",
			dnsZone: testDeletedDNSZoneWithoutFinalizer(),
			validate: func(c client.Client, a util.FakeActuator) {
				if a.CallCount() != 0 {
					t.Errorf("expected a no-op for a deleted zone with no finalizer")
				}
			},
		},
		{
			name:    "Deleted with finalizer",
			dnsZone: testDeletedDNSZone(),
			validate: func(c client.Client, a util.FakeActuator) {
				if !a.Called("Delete") {
					t.Errorf("expected Delete to be called on the actuator")
				}
				r := getZone(c)
				if util.HasFinalizer(r, cloudkitv1.DNSZoneFinalizer) {
					t.Errorf("expected finalizer to be removed from dnszone")
				}
			},
		},
		{
			name:    "Actuator Delete error",
			dnsZone: testDeletedDNSZone(),
			setupActuator: func(a util.FakeActuator) {
				a.OnDelete(errors.New("delete error"))
			},
			expectErr: true,
		},
		{
			name:    "Existing DNS zone",
			dnsZone: testDNSZone(),
			setupActuator: func(a util.FakeActuator) {
				a.OnExists(true, nil)
			},
			validate: func(c client.Client, a util.FakeActuator) {
				if !a.Called("Update") {
					t.Errorf("expected Update to be called on the actuator")
				}
			},
		},
		{
			name:    "Exists error",
			dnsZone: testDNSZone(),
			setupActuator: func(a util.FakeActuator) {
				a.OnExists(true, errors.New("exists error"))
			},
			expectErr: true,
		},
		{
			name:    "Update error",
			dnsZone: testDNSZone(),
			setupActuator: func(a util.FakeActuator) {
				a.OnExists(true, nil)
				a.OnUpdate(errors.New("update error"))
			},
			expectErr: true,
		},
		{
			name: "Non-existent DNS zone",
			setupActuator: func(a util.FakeActuator) {
				a.OnExists(false, nil)
			},
			dnsZone: testDNSZone(),
			validate: func(c client.Client, a util.FakeActuator) {
				if !a.Called("Create") {
					t.Errorf("expected Create to be called on the actuator")
				}
			},
		},
		{
			name:    "Create error",
			dnsZone: testDNSZone(),
			setupActuator: func(a util.FakeActuator) {
				a.OnExists(false, nil)
				a.OnCreate(errors.New("create error"))
			},
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeClient := fake.NewFakeClient(test.dnsZone)
			actuator, fakeActuator := NewFakeActuator()
			if test.setupActuator != nil {
				test.setupActuator(fakeActuator)
			}
			rcd := &ReconcileDNSZone{
				Client:   fakeClient,
				scheme:   scheme.Scheme,
				actuator: actuator,
				log:      log.WithField("controller", "dnszone-test"),
			}

			_, err := rcd.Reconcile(reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      testName,
					Namespace: testNamespace,
				},
			})

			if test.validate != nil {
				test.validate(fakeClient, fakeActuator)
			}

			if err != nil && !test.expectErr {
				t.Errorf("Unexpected error: %v", err)
			}
			if err == nil && test.expectErr {
				t.Errorf("Expected error but got none")
			}
		})
	}
}

func testDNSZone() *cloudkitv1.DNSZone {
	r := &cloudkitv1.DNSZone{}
	r.Name = testName
	r.Namespace = testNamespace
	r.Finalizers = []string{cloudkitv1.DNSZoneFinalizer}
	r.UID = types.UID("1234")
	r.Spec = cloudkitv1.DNSZoneSpec{
		ZoneName: "zone",
	}
	return r
}

func testDNSZoneWithoutFinalizer() *cloudkitv1.DNSZone {
	r := testDNSZone()
	r.Finalizers = []string{}
	return r
}

func testDeletedDNSZoneWithoutFinalizer() *cloudkitv1.DNSZone {
	r := testDNSZoneWithoutFinalizer()
	now := metav1.Now()
	r.DeletionTimestamp = &now
	return r
}

func testDeletedDNSZone() *cloudkitv1.DNSZone {
	r := testDNSZone()
	now := metav1.Now()
	r.DeletionTimestamp = &now
	return r
}
