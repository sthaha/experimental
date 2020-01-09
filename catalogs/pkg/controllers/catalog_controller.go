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

	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	"github.com/tektoncd/experimental/catalogs/pkg/api/v1alpha1"
	catalogv1alpha1 "github.com/tektoncd/experimental/catalogs/pkg/api/v1alpha1"
	"github.com/tektoncd/experimental/catalogs/pkg/git"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CatalogReconciler reconciles a Catalog object
type CatalogReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=catalog.tekton.dev,resources=catalogs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=catalog.tekton.dev,resources=catalogs/status,verbs=get;update;patch

func (r *CatalogReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&catalogv1alpha1.Catalog{}).
		Complete(r)
}

func (r *CatalogReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("catalog", req.NamespacedName)

	// your logic here
	ctg := v1alpha1.Catalog{}
	err := r.Get(ctx, req.NamespacedName, &ctg)

	if err != nil {
		log.Error(err, "getting catalog failed")
		if errors.IsNotFound(err) {
			log.Info("The catalog got deleted")
			return r.reconcileDeletion(req)
		}
		return ctrl.Result{}, err
	}

	return r.reconcileCatalog(ctg)

}

func (r *CatalogReconciler) reconcileCatalog(cat v1alpha1.Catalog) (ctrl.Result, error) {
	log := r.Log.WithValues("catalog", cat.Name)
	spec := cat.Spec

	log.Info(">>> cat", "url", spec.URL, "context", spec.ContextPath, "version", spec.Revision)

	// download the repo
	repo, err := git.Fetch(log, git.FetchSpec{
		URL:      spec.URL,
		Revision: spec.Revision,
		Path:     "/tmp/catalogs",
	})

	if repo.Head() == cat.Status.LastSync.Checksum {
		log.Info("Already at latest HEAD")
		return ctrl.Result{}, nil
	}

	log.Info("fetch error?", "err", err)
	// get the sha
	synced := cat.DeepCopy()

	now := metav1.Now()
	synced.Status.LastSync = v1alpha1.SyncInfo{
		Time:     &now,
		Checksum: repo.Head(),
	}

	r.Client.Update(context.Background(), synced)

	// get status sha
	// fill in the details

	return ctrl.Result{}, nil
}

func (r *CatalogReconciler) reconcileDeletion(req ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}
