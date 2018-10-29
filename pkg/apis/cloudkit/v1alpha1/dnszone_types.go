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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DNSZoneFinalizer is set on DNSZones to invoke Delete on their actuator
const DNSZoneFinalizer = "dnszone.cloudkit.openshift.io"

// DNSZoneSpec defines the desired state of DNSZone
type DNSZoneSpec struct {
	// ZoneName is the DNS name of the Zone
	ZoneName string `json:"zoneName"`

	// ProviderSpec is the spec specific to a particular cloud provider.
	ProviderSpec *runtime.RawExtension `json:"providerSpec,omitempty"`
}

// DNSZoneStatus defines the observed state of DNSZone
type DNSZoneStatus struct {
	//ProviderStatus is the status specific to a particular cloud provider.
	ProviderStatus *runtime.RawExtension `json:"providerStatus,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DNSZone is the Schema for the dnszones API
// +k8s:openapi-gen=true
type DNSZone struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DNSZoneSpec   `json:"spec,omitempty"`
	Status DNSZoneStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DNSZoneList contains a list of DNSZone
type DNSZoneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DNSZone `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DNSZone{}, &DNSZoneList{})
}
