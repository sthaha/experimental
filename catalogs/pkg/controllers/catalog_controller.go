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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/tektoncd/experimental/catalogs/pkg/api/v1alpha1"
	catalogv1alpha1 "github.com/tektoncd/experimental/catalogs/pkg/api/v1alpha1"
)

// CatalogReconciler reconciles a Catalog object
type CatalogReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=catalog.tekton.dev,resources=catalogs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=catalog.tekton.dev,resources=catalogs/status,verbs=get;update;patch

func (r *CatalogReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("catalog", req.NamespacedName)

	// your logic here
	ctg := v1alpha1.Catalog{}
	err := r.Get(ctx, req.NamespacedName, &ctg)
	if err != nil {
		log.Error(err, "getting catalog failed")
		return ctrl.Result{}, err
	}

	log.Info("got from", "cat", ctg)

	return ctrl.Result{}, nil
}

func (r *CatalogReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&catalogv1alpha1.Catalog{}).
		Complete(r)
}
