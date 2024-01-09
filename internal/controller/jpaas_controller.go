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
	utils "github.com/NovaZee/kubeDev/internal/util"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)

var paasLog = log.Log.WithName("controller_jpaas")

// JPaasReconciler reconciles a client object
type JPaasReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Log     logr.Logger
	Handler *JPaasHandler
}

//+kubebuilder:rbac:groups=hanweb.jpaas.deploy,resources=jpaas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=hanweb.jpaas.deploy,resources=jpaas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=hanweb.jpaas.deploy,resources=jpaas/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the client object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *JPaasReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//logger := r.Log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	logger := log.FromContext(ctx)
	logger.Info("client Reconcile Start")
	defer logger.Info("client Reconcile End")
	logger.Info("============", "client", req.String())
	instance := &hanwebv1beta1.JPaas{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		logger.Error(err, "get client error")
		return ctrl.Result{}, err
	}
	// 需要进行初始化
	if !instance.Spec.Initialized {
		reconcile, initErr := r.InitReconcile(ctx, instance, req)
		if err != nil {
			return reconcile, initErr
		} else {
			// 更新状态
			instance.Spec.Initialized = true
			if err = r.Update(ctx, instance); err != nil {
				logger.Error(err, "failed to update JPaas CRD", "JPaas", instance)
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *JPaasReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hanwebv1beta1.JPaas{}).
		Owns(&v1.Deployment{}).
		//Owns(&corev1.Service{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}

func (r *JPaasReconciler) InitReconcile(ctx context.Context, instance *hanwebv1beta1.JPaas, req ctrl.Request) (ctrl.Result, error) {
	// condition 1: 部署platform,根据platform的依赖关系，部署基础应用
	// 获取 base apps
	baseApps := instance.Spec.AppRefs
	for i, app := range baseApps {
		// 初始化crd
		if strings.HasPrefix(app.Name, "common-") {
			app.Type = hanwebv1beta1.BaseApp
		} else {
			app.Type = hanwebv1beta1.Application // or any other default value
		}
		crd := utils.InitApp(*instance, app)
		if err := r.Create(ctx, crd); err != nil {
			if errors.IsAlreadyExists(err) {
				// 如果CRD已经存在，那么更新它
				if err = r.Update(ctx, crd); err != nil {
					//todo：更新逻辑 主要是版本升级
					return ctrl.Result{}, nil
				}
			} else {
				r.Log.Error(err, "failed to create CRD", "crd", crd)
				return ctrl.Result{}, err
			}
		} else {
			// 如果CRD创建成功，更新JPaas CRD中AppRefs集合该资源的status为true
			if !instance.Spec.AppRefs[i].JPaasAppSuccess {
				instance.Spec.AppRefs[i].JPaasAppSuccess = true
			}
		}
		// 初始化deployment
		deployment := utils.NewDeployment(crd)
		if err := controllerutil.SetControllerReference(instance, deployment, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		d := &v1.Deployment{}
		if err := r.Get(ctx, req.NamespacedName, d); err != nil {
			if errors.IsNotFound(err) {
				if err = r.Create(ctx, deployment); err != nil {
					r.Log.Error(err, "create deploy failed")
					return ctrl.Result{}, err
				}
			}
		} else {

		}
		// 初始化svc
		service := utils.NewService(crd)
		if err := controllerutil.SetControllerReference(instance, service, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		s := &corev1.Service{}
		if err := r.Get(ctx, req.NamespacedName, s); err != nil {
			if errors.IsNotFound(err) {
				if err = r.Create(ctx, service); err != nil {
					r.Log.Error(err, "create service failed")
					return ctrl.Result{}, err
				}

			}
			if !errors.IsNotFound(err) {
				return ctrl.Result{}, err
			}
		} else {

		}
	}
	return ctrl.Result{}, nil
}
