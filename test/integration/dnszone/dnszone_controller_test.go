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
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/onsi/gomega"
	cloudkitv1alpha1 "github.com/openshift/cloud-kit/pkg/apis/cloudkit/v1alpha1"
	"golang.org/x/net/context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/openshift/cloud-kit/pkg/controller/dnszone"
	"github.com/openshift/cloud-kit/pkg/controller/util"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

var c client.Client

var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: "foo", Namespace: "default"}}
var zoneKey = types.NamespacedName{Name: "foo", Namespace: "default"}

const timeout = time.Second * 5

func TestReconcile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	instance := &cloudkitv1alpha1.DNSZone{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default"}}

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c = mgr.GetClient()

	actuator, fakeActuator := dnszone.NewFakeActuator()
	recFn, requests := SetupTestReconcile(dnszone.NewReconciler(mgr, actuator))
	g.Expect(dnszone.Add(mgr, recFn)).NotTo(gomega.HaveOccurred())

	stopMgr, mgrStopped := StartTestManager(mgr, g)

	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	// Create the DNSZone object and expect the Create method to be called on the actuator
	err = c.Create(context.TODO(), instance)
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	g.Eventually(func() bool {
		zone := &cloudkitv1alpha1.DNSZone{}
		err := c.Get(context.TODO(), zoneKey, zone)
		g.Expect(err).NotTo(gomega.HaveOccurred())
		return util.HasFinalizer(zone, cloudkitv1alpha1.DNSZoneFinalizer)
	}, timeout).Should(gomega.BeTrue())

	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	g.Eventually(fakeActuator.Called("Create"), timeout).Should(gomega.BeTrue())

	// Update the DNSZone and expect that Update is called on the actuator
	fakeActuator.OnExists(true, nil)

	err = c.Get(context.TODO(), zoneKey, instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	instance.Annotations = map[string]string{"foo": "bar"}
	err = c.Update(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	g.Eventually(fakeActuator.Called("Update"), timeout).Should(gomega.BeTrue())

	// Delete the DNSZone and expect that Delete is called on the actuator
	err = c.Delete(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	g.Eventually(fakeActuator.Called("Delete"), timeout).Should(gomega.BeTrue())
}
