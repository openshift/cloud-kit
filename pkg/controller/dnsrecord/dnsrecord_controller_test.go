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
	testName      = "dnsrecord"
	testNamespace = "testns"
)

func TestDNSRecordReconcile(t *testing.T) {
	apis.AddToScheme(scheme.Scheme)

	// Utility function to get the test DNS record from the fake client
	getRecord := func(c client.Client) *cloudkitv1.DNSRecord {
		record := &cloudkitv1.DNSRecord{}
		err := c.Get(context.TODO(), client.ObjectKey{Name: testName, Namespace: testNamespace}, record)
		if err == nil {
			return record
		}
		return nil
	}

	tests := []struct {
		name          string
		dnsRecord     *cloudkitv1.DNSRecord
		actuator      Actuator
		expectErr     bool
		setupActuator func(util.FakeActuator)
		validate      func(client.Client, util.FakeActuator)
	}{
		{
			name:      "Add finalizer",
			dnsRecord: testDNSRecordWithoutFinalizer(),
			validate: func(c client.Client, a util.FakeActuator) {
				r := getRecord(c)
				if r == nil || !util.HasFinalizer(r, cloudkitv1.DNSRecordFinalizer) {
					t.Errorf("did not get expected dnsrecord finalizer")
				}
			},
		},
		{
			name:      "Deleted with no finalizer",
			dnsRecord: testDeletedDNSRecordWithoutFinalizer(),
			validate: func(c client.Client, a util.FakeActuator) {
				if a.CallCount() != 0 {
					t.Errorf("expected a no-op for a deleted record with no finalizer")
				}
			},
		},
		{
			name:      "Deleted with finalizer",
			dnsRecord: testDeletedDNSRecord(),
			validate: func(c client.Client, a util.FakeActuator) {
				if !a.Called("Delete") {
					t.Errorf("expected Delete to be called on the actuator")
				}
				r := getRecord(c)
				if util.HasFinalizer(r, cloudkitv1.DNSRecordFinalizer) {
					t.Errorf("expected finalizer to be removed from dnsrecord")
				}
			},
		},
		{
			name:      "Actuator Delete error",
			dnsRecord: testDeletedDNSRecord(),
			setupActuator: func(a util.FakeActuator) {
				a.OnDelete(errors.New("delete error"))
			},
			expectErr: true,
		},
		{
			name:      "Existing DNS record",
			dnsRecord: testDNSRecord(),
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
			name:      "Exists error",
			dnsRecord: testDNSRecord(),
			setupActuator: func(a util.FakeActuator) {
				a.OnExists(true, errors.New("exists error"))
			},
			expectErr: true,
		},
		{
			name:      "Update error",
			dnsRecord: testDNSRecord(),
			setupActuator: func(a util.FakeActuator) {
				a.OnExists(true, nil)
				a.OnUpdate(errors.New("update error"))
			},
			expectErr: true,
		},
		{
			name: "Non-existent DNS record",
			setupActuator: func(a util.FakeActuator) {
				a.OnExists(false, nil)
			},
			dnsRecord: testDNSRecord(),
			validate: func(c client.Client, a util.FakeActuator) {
				if !a.Called("Create") {
					t.Errorf("expected Create to be called on the actuator")
				}
			},
		},
		{
			name:      "Create error",
			dnsRecord: testDNSRecord(),
			setupActuator: func(a util.FakeActuator) {
				a.OnExists(false, nil)
				a.OnCreate(errors.New("create error"))
			},
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeClient := fake.NewFakeClient(test.dnsRecord)
			actuator, fakeActuator := NewFakeActuator()
			if test.setupActuator != nil {
				test.setupActuator(fakeActuator)
			}
			rcd := &ReconcileDNSRecord{
				Client:   fakeClient,
				scheme:   scheme.Scheme,
				actuator: actuator,
				log:      log.WithField("controller", "dnsrecord-test"),
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

func testDNSRecord() *cloudkitv1.DNSRecord {
	r := &cloudkitv1.DNSRecord{}
	r.Name = testName
	r.Namespace = testNamespace
	r.Finalizers = []string{cloudkitv1.DNSRecordFinalizer}
	r.UID = types.UID("1234")
	r.Spec = cloudkitv1.DNSRecordSpec{
		ZoneName:   "zone",
		RecordName: "record",
	}
	return r
}

func testDNSRecordWithoutFinalizer() *cloudkitv1.DNSRecord {
	r := testDNSRecord()
	r.Finalizers = []string{}
	return r
}

func testDeletedDNSRecordWithoutFinalizer() *cloudkitv1.DNSRecord {
	r := testDNSRecordWithoutFinalizer()
	now := metav1.Now()
	r.DeletionTimestamp = &now
	return r
}

func testDeletedDNSRecord() *cloudkitv1.DNSRecord {
	r := testDNSRecord()
	now := metav1.Now()
	r.DeletionTimestamp = &now
	return r
}
