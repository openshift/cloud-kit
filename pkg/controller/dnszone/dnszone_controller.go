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

	log "github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	cloudkitv1 "github.com/openshift/cloud-kit/pkg/apis/cloudkit/v1alpha1"
	"github.com/openshift/cloud-kit/pkg/controller/util"
)

// AddWithActuator creates a new DNSZone Controller and adds it to the Manager with default RBAC.
// The Manager will set fields on the Controller and Start it when the Manager is Started.
func AddWithActuator(mgr manager.Manager, actuator Actuator) error {
	return Add(mgr, NewReconciler(mgr, actuator))
}

// NewReconciler returns a new reconcile.Reconciler
func NewReconciler(mgr manager.Manager, actuator Actuator) reconcile.Reconciler {
	return &ReconcileDNSZone{
		Client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		actuator: actuator,
		log:      log.WithField("controller", "dnszone"),
	}
}

// Add adds a new Controller to mgr with r as the reconcile.Reconciler
func Add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("dnszone-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		log.WithError(err).Error("error creating a new dnszone controller")
		return err
	}

	// Watch for changes to DNSZone
	err = c.Watch(&source.Kind{Type: &cloudkitv1.DNSZone{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		log.WithError(err).Error("error starting watch of DNSZone")
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileDNSZone{}

// ReconcileDNSZone reconciles a DNSZone object
type ReconcileDNSZone struct {
	client.Client
	scheme   *runtime.Scheme
	actuator Actuator
	log      log.FieldLogger
}

// Reconcile reads that state of the cluster for a DNSZone object and makes changes based on the state read
// and what is in the DNSZone.Spec
// Automatically generate RBAC rules to allow the Controller to read and write DNSZones
// +kubebuilder:rbac:groups=cloudkit.openshift.io,resources=dnszones,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileDNSZone) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the DNSZone instance
	logger := r.log.WithField("dnszone", request.NamespacedName.String())
	logger.Info("syncing dnszone")
	zone := &cloudkitv1.DNSZone{}
	err := r.Get(context.TODO(), request.NamespacedName, zone)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			logger.Warning("resource was not found, nothing to do")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.WithError(err).Error("error retrieving resource")
		return reconcile.Result{}, err
	}

	// If DNS zone hasn't been deleted and doesn't have a finalizer, add one
	if zone.ObjectMeta.DeletionTimestamp.IsZero() &&
		!util.HasFinalizer(zone, cloudkitv1.DNSZoneFinalizer) {
		util.AddFinalizer(zone, cloudkitv1.DNSZoneFinalizer)
		if err = r.Update(context.Background(), zone); err != nil {
			logger.WithError(err).Error("error adding finalizer")
			return reconcile.Result{}, err
		}
		logger.Debug("finalizer added")
		return reconcile.Result{}, nil
	}

	if !zone.ObjectMeta.DeletionTimestamp.IsZero() {
		// no-op if finalizer has been removed.
		if !util.HasFinalizer(zone, cloudkitv1.DNSZoneFinalizer) {
			logger.Debug("deleted resource with no finalizer present, nothing to do")
			return reconcile.Result{}, nil
		}
		logger.Debug("reconciling dnszone triggers actuator delete.")
		if err := r.actuator.Delete(zone); err != nil {
			logger.WithError(err).Error("error deleting dnszone object")
			return reconcile.Result{}, err
		}

		// Remove finalizer on successful deletion.
		logger.Debug("dnszone deletion successful, removing finalizer.")
		util.DeleteFinalizer(zone, cloudkitv1.DNSZoneFinalizer)
		if err := r.Client.Update(context.Background(), zone); err != nil {
			logger.WithError(err).Error("error removing finalizer from dnszone object")
			return reconcile.Result{}, err
		}
		logger.Debug("dnszone finalizer removed")
		return reconcile.Result{}, nil
	}

	exist, err := r.actuator.Exists(zone)
	if err != nil {
		logger.WithError(err).Error("error checking existence of dnszone")
		return reconcile.Result{}, err
	}
	if exist {
		logger.Debug("dnszone exists, calling idempotent update")
		err := r.actuator.Update(zone)
		if err != nil {
			logger.WithError(err).Error("error updating dnszone")
			return reconcile.Result{}, err
		}
		logger.Debug("dnszone updated successfully")
		return reconcile.Result{}, nil
	}
	logger.Debug("dnszone does not exist, calling create")
	if err := r.actuator.Create(zone); err != nil {
		logger.WithError(err).Error("error creating dnszone")
		return reconcile.Result{}, err
	}
	logger.Debug("dnszone created successfully")

	return reconcile.Result{}, nil
}
