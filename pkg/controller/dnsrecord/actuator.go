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
)

// Actuator is an interface that must be implemented by a specific cloud
// provider to manage DNS records.
type Actuator interface {
	// Create the DNS record.
	Create(*cloudkitv1.DNSRecord) error
	// Delete the DNS record.
	Delete(*cloudkitv1.DNSRecord) error
	// Update the DNS record.
	Update(*cloudkitv1.DNSRecord) error
	// Checks if the DNS record currently exists.
	Exists(*cloudkitv1.DNSRecord) (bool, error)
}
