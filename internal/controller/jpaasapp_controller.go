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
	hanwebv1client "github.com/NovaZee/kubeDev/client"
	"github.com/go-logr/logr"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// JPaasAppReconciler reconciles a JPaasApp object
type JPaasAppReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=hanweb.jpaas.deploy,resources=jpaasapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=hanweb.jpaas.deploy,resources=jpaasapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=hanweb.jpaas.deploy,resources=jpaasapps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the JPaasApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *JPaasAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var log = r.Log.WithValues(
		"JPaas", req.NamespacedName)
	log.Info("JPaasAppReconciler Start")
	defer log.Info("JPaasAppReconciler End")

	paasAppCr, err := hanwebv1client.NewJPaasAppCR(ctx, req, log, r.Client, r.Scheme)
	if err != nil {
		return reconcile.Result{}, err
	}
	paasReconcile, err := paasAppCr.PaasAppReconcile()
	if err != nil {
		return paasReconcile, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *JPaasAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hanwebv1beta1.JPaasApp{}).
		Owns(&appv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
