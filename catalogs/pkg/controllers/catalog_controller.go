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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/tektoncd/experimental/catalogs/pkg/api/v1alpha1"
	catalogv1alpha1 "github.com/tektoncd/experimental/catalogs/pkg/api/v1alpha1"
	"github.com/tektoncd/experimental/catalogs/pkg/git"
)

// CatalogReconciler reconciles a Catalog object
type CatalogReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=catalog.tekton.dev,resources=catalogs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=catalog.tekton.dev,resources=catalogs/status,verbs=get;update;patch

func (r *CatalogReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("req", req.NamespacedName)

	// your logic here
	ctg := v1alpha1.Catalog{}
	log.Info("finding req")
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

func (r *CatalogReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&catalogv1alpha1.Catalog{}).
		Complete(r)
}

func (r *CatalogReconciler) markError(c v1alpha1.Catalog, err error) (ctrl.Result, error) {
	cat := c.DeepCopy()

	now := metav1.Now()
	cat.Status.LastSync = v1alpha1.SyncInfo{Time: &now}

	cat.Status.Condition = v1alpha1.CatalogCondition{
		Code:    v1alpha1.ErrorCondition,
		Details: err.Error(),
	}

	if updateErr := r.Client.Update(context.Background(), cat); updateErr != nil {
		r.Log.Error(updateErr, "error setting catalog status to error")
	}

	return ctrl.Result{}, err
}

func (r *CatalogReconciler) reconcileCatalog(cat v1alpha1.Catalog) (ctrl.Result, error) {
	log := r.Log.WithValues("catalog", cat.Name)
	spec := cat.Spec
	status := cat.Status

	log.Info(">>> cat", "url", spec.URL, "context", spec.ContextPath, "version", spec.Revision)

	// download the repo
	repo, err := git.Fetch(git.FetchSpec{
		URL:      spec.URL,
		Revision: spec.Revision,
		Path:     "/tmp/catalogs",
	})

	if err != nil {
		return r.markError(cat, err)
	}

	log.Info(">>> finding resources in repo", "url", spec.URL, "context", spec.ContextPath, "version", spec.Revision)

	if repo.Head() == cat.Status.LastSync.Revision &&
		status.Condition.Is(v1alpha1.SuccessfullSync) {
		log.Info("Already at latest HEAD")
		return ctrl.Result{}, nil
	}

	synced := cat.DeepCopy()
	now := metav1.Now()
	synced.Status = v1alpha1.CatalogStatus{
		Tasks:        nil,
		ClusterTasks: nil,
		// default status successfull
		Condition: v1alpha1.CatalogCondition{Code: v1alpha1.SuccessfullSync},

		LastSync: v1alpha1.SyncInfo{
			Time:     &now,
			Revision: repo.Head(),
		},
	}

	tasks, err := repo.Tasks()
	if err != nil {
		r.Log.Error(err, "error finding tasks")
		synced.Status.Condition.SetError("unable to find  tasks")
	} else {
		synced.Status.Tasks = tasks
	}

	clusterTasks, err := repo.ClusterTasks()
	if err != nil {
		r.Log.Error(err, "error finding cluster tasks")
		synced.Status.Condition.SetError("enabled to find clustertasks")
	} else {
		synced.Status.ClusterTasks = clusterTasks
	}

	r.Client.Update(context.Background(), synced)
	return ctrl.Result{}, nil

}

func (r *CatalogReconciler) reconcileDeletion(req ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}
