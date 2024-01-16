/*
Copyright 2024.

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

package controller

import (
	"context"
	hanwebv1beta1 "github.com/NovaZee/kubeDev/api/v1beta1"
	"github.com/NovaZee/kubeDev/controller/paasterm"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// JPaasReconciler reconciles a paasterm object
type JPaasReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=hanweb.jpaas.deploy,resources=jpaas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=hanweb.jpaas.deploy,resources=jpaas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=hanweb.jpaas.deploy,resources=jpaas/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the paasterm object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *JPaasReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//logge1r := log.FromContext(ctx)
	//defer1 logger.Info("JPaasReconciler End")
	var log = r.Log.WithValues(
		"JPaas", req.NamespacedName)
	log.Info("JPaasReconciler Start")

	defer log.Info("JPaasReconciler End")

	paasCr, err := paasterm.NewJPaasCR(ctx, req, log, r.Client)
	if err != nil {
		return reconcile.Result{}, err
	}
	paasReconcile, err := paasCr.PaasReconcile()
	if err != nil {
		return paasReconcile, err
	}
	return paasReconcile, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *JPaasReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hanwebv1beta1.JPaas{}).
		Complete(r)
}
