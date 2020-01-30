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
	apis "github.com/tektoncd/experimental/catalogs/pkg/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CatalogInstallReconciler reconciles a CatalogInstall object
type CatalogInstallReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=apis.tekton.dev,resources=cataloginstalls,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apis.tekton.dev,resources=cataloginstalls/status,verbs=get;update;patch

func (r *CatalogInstallReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("req", req.NamespacedName)

	ci := apis.CatalogInstall{}
	err := r.Get(ctx, req.NamespacedName, &ci)

	if err != nil {
		log.Error(err, "getting resource failed")
		if errors.IsNotFound(err) {
			log.Info("Reconcile deletion of resource")
			return r.reconcileDeletion(req)
		}
		return ctrl.Result{}, err
	}

	return r.reconcileInstall(ci)
}

func (r *CatalogInstallReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apis.CatalogInstall{}).
		Complete(r)
}

func (r *CatalogInstallReconciler) reconcileDeletion(req ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *CatalogInstallReconciler) reconcileInstall(ci apis.CatalogInstall) (ctrl.Result, error) {
	// does the catalog exist?
	cat, err := r.catalogForRef(ci.Spec.CatalogRef)
	if err != nil {
		r.markError(ci, err)
		return ctrl.Result{}, nil
	}
	// can I get the git repo?
	// does the tasks exist?
	// can I apply the task?

	if res, err := r.reconcileTasks(ci, *cat); err != nil {
		return res, err
	}
	return ctrl.Result{}, nil
}

func (r *CatalogInstallReconciler) reconcileTasks(ci apis.CatalogInstall, cat apis.Catalog) (ctrl.Result, error) {
	log := r.Log.WithValues("install-tasks", ci.Namespace)

	tasks := ci.Spec.Tasks
	log.Info("Install ", "tasks", tasks)
	if tasks == nil || len(tasks) == 0 {
		return ctrl.Result{}, nil
	}

	//for _, t := range tasks {

	//}

	return ctrl.Result{}, nil
}

func (r *CatalogInstallReconciler) catalogForRef(ref string) (*apis.Catalog, error) {
	log := r.Log.WithValues("ref", ref)

	log.Info("finding ref")
	res := types.NamespacedName{Name: ref}
	cat := apis.Catalog{}
	if err := r.Get(context.Background(), res, &cat); err != nil {
		log.Error(err, "failed wtf!")
		return nil, err
	}

	return &cat, nil
}

func (r *CatalogInstallReconciler) markError(c apis.CatalogInstall, err error) (ctrl.Result, error) {
	ci := c.DeepCopy()
	ci.Status.Condition = apis.CatalogInstallCondition{
		Code:    apis.InstallError,
		Details: err.Error(),
	}

	if updateErr := r.Client.Update(context.Background(), ci); updateErr != nil {
		r.Log.Error(updateErr, "error setting catalog status to error")
	}

	return ctrl.Result{}, err
}
