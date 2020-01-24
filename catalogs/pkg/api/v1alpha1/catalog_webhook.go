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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var cataloglog = logf.Log.WithName("catalog-resource")

func (r *Catalog) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-catalog-tekton-dev-v1alpha1-catalog,mutating=true,failurePolicy=fail,groups=catalog.tekton.dev,resources=catalogs,verbs=create;update,versions=v1alpha1,name=mcatalog.kb.io

var _ webhook.Defaulter = &Catalog{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Catalog) Default() {
	cataloglog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// +kubebuilder:webhook:path=/validate-catalog-tekton-dev-v1alpha1-catalog,mutating=false,failurePolicy=fail,groups=catalog.tekton.dev,resources=catalogs,verbs=create;update,versions=v1alpha1,name=vcatalog.kb.io

var _ webhook.Validator = &Catalog{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Catalog) ValidateCreate() error {
	cataloglog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Catalog) ValidateUpdate(old runtime.Object) error {
	cataloglog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}
