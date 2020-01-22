/*
Copyright 2019 The Tekton Authors
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
)

// CatalogSpec defines the desired state of Catalog
type CatalogSpec struct {
	URL         string `json:"url"`
	ContextPath string `json:"contextPath,omitempty"`
	Revision    string `json:"revision,omitempty"`
}

// CatalogStatus defines the observed state of Catalog
type CatalogStatus struct {
	// Information when was the last time the job was successfully scheduled.
	// +optional
	LastSync  SyncInfo         `json:"lastSynced,omitempty"`
	Condition CatalogCondition `json:"condition,omitempty"`

	// +optional
	Tasks        []TaskInfo `json:"tasks,omitempty"`
	ClusterTasks []TaskInfo `json:"clustertasks,omitempty"`
}

type SyncInfo struct {
	// Information when was the last time the job was successfully scheduled.
	// +optional
	Time     *metav1.Time `json:"time,omitempty"`
	Revision string       `json:"revision,omitempty"`
}

type TaskInfo struct {
	Name     string   `json:"name"`
	Versions []string `json:"versions"`
}

// +kubebuilder:object:root=true

// Catalog is the Schema for the catalogs API
type Catalog struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CatalogSpec   `json:"spec,omitempty"`
	Status CatalogStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CatalogList contains a list of Catalog
type CatalogList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Catalog `json:"items"`
}

// CatalogCondition represents the current condition of the catalog
type CatalogCondition string

const UnknownCondition CatalogCondition = "unknown"
const ErrorCondition CatalogCondition = "error"
const SuccessfullSync CatalogCondition = "success"

func init() {
	SchemeBuilder.Register(&Catalog{}, &CatalogList{})
}
