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

// DNSRecordFinalizer is set on DNSRecords to invoke Delete on their actuator
const DNSRecordFinalizer = "dnsrecord.cloudkit.openshift.io"

// DNSRecordSpec defines the desired state of DNSRecord
type DNSRecordSpec struct {

	// ZoneName is the name of the zone where this DNS record should live.
	// Depending on the provider, this may not be enough to identify the
	// zone and additional information must be provider in the providerSpec
	ZoneName string `json:"zoneName"`

	// RecordName is the name of the DNS record. This must be a valid DNS name
	// or wildcard.
	RecordName string `json:"recordName"`

	// RecordType is the type of DNS record
	RecordType DNSRecordType `json:"recordType"`

	// Value is the value of the DNS record. The format of the value will depend
	// on the type of DNS Record. If empty, the cloud provider should populate it
	// with the current value on the cloud record if it already exists.
	Value *string `json:"value,omitempty"`

	// ProviderSpec is the spec specific to a particular cloud provider.
	ProviderSpec *runtime.RawExtension `json:"providerSpec,omitempty"`
}

// DNSRecordStatus defines the observed state of DNSRecord
type DNSRecordStatus struct {
	//ProviderStatus is the status specific to a particular cloud provider.
	ProviderStatus *runtime.RawExtension `json:"providerStatus,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DNSRecord is the Schema for the dnsrecords API
// +k8s:openapi-gen=true
type DNSRecord struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DNSRecordSpec   `json:"spec,omitempty"`
	Status DNSRecordStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DNSRecordList contains a list of DNSRecord
type DNSRecordList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DNSRecord `json:"items"`
}

// DNSRecordType is the type of a DNS record
type DNSRecordType string

const (
	DNSRecordA     DNSRecordType = "A"
	DNSRecordCNAME DNSRecordType = "CNAME"
	DNSRecordMX    DNSRecordType = "MX"
	DNSRecordAAAA  DNSRecordType = "AAAA"
	DNSRecordTXT   DNSRecordType = "TXT"
	DNSRecordPTR   DNSRecordType = "PTR"
	DNSRecordSRV   DNSRecordType = "SRV"
	DNSRecordSPF   DNSRecordType = "SPF"
	DNSRecordNAPTR DNSRecordType = "NAPTR"
	DNSRecordCAA   DNSRecordType = "CAA"
	DNSRecordNS    DNSRecordType = "NS"
	DNSRecordSOA   DNSRecordType = "SOA"
)

func init() {
	SchemeBuilder.Register(&DNSRecord{}, &DNSRecordList{})
}
