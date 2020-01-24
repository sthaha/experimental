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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CatalogInstallSpec defines the desired state of CatalogInstall
type CatalogInstallSpec struct {
	CatalogRef CatalogRef        `json:"catalogRef"`
	Tasks      []TaskInstallSpec `json:"tasks"`
}

type CatalogRef string

type TaskInstallSpec struct {
	Name string `json:"name"`
}

// CatalogInstallStatus defines the observed state of CatalogInstall
type CatalogInstallStatus struct {
}

// +kubebuilder:object:root=true

// CatalogInstall is the Schema for the cataloginstalls API
type CatalogInstall struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CatalogInstallSpec   `json:"spec,omitempty"`
	Status CatalogInstallStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CatalogInstallList contains a list of CatalogInstall
type CatalogInstallList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CatalogInstall `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CatalogInstall{}, &CatalogInstallList{})
}
