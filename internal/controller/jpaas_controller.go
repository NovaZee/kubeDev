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
	utils "github.com/NovaZee/kubeDev/internal/util"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	hanwebv1beta1 "github.com/NovaZee/kubeDev/api/v1beta1"
)

// JPaasReconciler reconciles a JPaas object
type JPaasReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=hanweb.jpaas.deploy,resources=jpaas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=hanweb.jpaas.deploy,resources=jpaas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=hanweb.jpaas.deploy,resources=jpaas/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the JPaas object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *JPaasReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("JPaas Reconcile Start")
	defer logger.Info("JPaas Reconcile End")

	JPaas := &hanwebv1beta1.JPaas{}
	err := r.Get(ctx, req.NamespacedName, JPaas)
	if err != nil {
		logger.Error(err, "get JPaas error")
		return ctrl.Result{}, err
	}
	// condition 1: 部署platform，根据platform的依赖关系，部署基础应用
	if JPaas.Spec.Type == hanwebv1beta1.Platform {
		// 获取 base apps
		baseApps := JPaas.Spec.AppRefs
		for _, app := range baseApps {
			// 初始化crd
			crd := utils.InitCrd(JPaas, &app)
			if err = r.Create(ctx, crd); err != nil {
				if errors.IsAlreadyExists(err) {
					// 如果CRD已经存在，那么更新它
					if err = r.Update(ctx, crd); err != nil {
						logger.Error(err, "failed to update CRD", "crd", crd)
						return ctrl.Result{}, err
					}
				} else {
					logger.Error(err, "failed to create CRD", "crd", crd)
					return ctrl.Result{}, err
				}
			}
			// 初始化deployment
			deployment := utils.NewDeployment(JPaas)
			if err := controllerutil.SetControllerReference(JPaas, deployment, r.Scheme); err != nil {
				return ctrl.Result{}, err
			}
			d := &v1.Deployment{}
			if err := r.Get(ctx, req.NamespacedName, d); err != nil {
				if errors.IsNotFound(err) {
					if err := r.Create(ctx, deployment); err != nil {
						logger.Error(err, "create deploy failed")
						return ctrl.Result{}, err
					}
				}
			} else {
				if err := r.Update(ctx, deployment); err != nil {
					return ctrl.Result{}, err
				}
			}
			// 初始化svc
		}
	}

	for i := range JPaas.Name {
		logger.Info(string(JPaas.Name[i]))
	}
	//// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *JPaasReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hanwebv1beta1.JPaas{}).
		Complete(r)
}
