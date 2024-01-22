package paasterm

import (
	"context"
	hanwebv1beta1 "github.com/NovaZee/kubeDev/api/v1beta1"
	utils "github.com/NovaZee/kubeDev/controller/util"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
)

type JPaasCR struct {
	kubeClient client.Client
	paas       *hanwebv1beta1.JPaas
	context    context.Context
	request    ctrl.Request
	log        logr.Logger
}

// NewJPaasCR - Create a new JPaasCR.
func NewJPaasCR(ctx context.Context, req ctrl.Request, log logr.Logger, c client.Client) (*JPaasCR, error) {
	instance := &hanwebv1beta1.JPaas{}
	err := c.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		return nil, err
	}
	return &JPaasCR{
		context:    ctx,
		request:    req,
		log:        log,
		paas:       instance,
		kubeClient: c,
	}, nil
}

func (jc *JPaasCR) PaasReconcile() (ctrl.Result, error) {
	err := jc.Finalizer()
	if err != nil {
		return ctrl.Result{}, err
	}
	if !jc.paas.Spec.Initialized {
		return ctrl.Result{}, jc.ResourceInitCheck()
	}
	err = jc.checkAppCr()
	if err != nil {
		return ctrl.Result{}, err
	}
	// crd列表中有 app的crd not found 代表 app的crd没有创建成功或者被删除了

	return ctrl.Result{}, nil
}

func (jc *JPaasCR) ResourceInitCheck() error {
	// condition 1: 部署platform,根据platform的依赖关系，部署基础应用
	// 获取 base apps
	var instance = jc.paas
	var ctx = jc.context
	var log = jc.log
	var c = jc.kubeClient
	baseApps := instance.Spec.AppRefs
	for i, app := range baseApps {
		// 创建失败的err
		_ = initAppCr(app, instance, ctx, c, log, i)
	}
	instance.Spec.Initialized = true
	if err := c.Update(ctx, instance); err != nil {
		log.Error(err, "failed to update JPaas CRD", "crd", instance)
		return nil
	}
	return nil
}

func (jc *JPaasCR) checkAppCr() error {
	var ctx, kubeClient, log, namespace, jpaas = jc.context, jc.kubeClient, jc.log, jc.request.NamespacedName, jc.paas
	err := kubeClient.Get(ctx, namespace, jpaas)
	if jpaas.Spec.Initialized {
		// Find the JPaasApp in apprefs and set JPaasAppSuccess to false
		for i, appref := range jpaas.Spec.AppRefs {
			if appref.AppCrStatus == hanwebv1beta1.ConditionUnInit || appref.AppCrStatus == hanwebv1beta1.ConditionFailed {
				_ = initAppCr(appref, jpaas, ctx, kubeClient, log, i)
			}
		}
		if err = kubeClient.Update(ctx, jpaas); err != nil {
			log.Error(err, "failed to update JPaas CRD", "crd", jpaas)
			return err
		}
	}

	return nil
}

func (jc *JPaasCR) ResourceRuntimeCheck() error {
	return nil
}

func initAppCr(app hanwebv1beta1.AppSpec, instance *hanwebv1beta1.JPaas, ctx context.Context, c client.Client, log logr.Logger, i int) error {
	// 初始化crd
	if strings.HasPrefix(app.Name, "common-") {
		app.Type = hanwebv1beta1.BaseApp
	} else {
		app.Type = hanwebv1beta1.Application // or any other default value
	}
	crd := utils.InitApp(*instance, app)
	if err := c.Create(ctx, crd); err != nil {
		if errors.IsAlreadyExists(err) {
			log.Info(" CRD IsAlreadyExists", "crd", crd)
			// 如果CRD已经存在，那么更新它
		} else {
			log.Error(err, "failed to create CRD", "crd", crd)
			return err
		}
	} else {
		// 如果CRD创建成功，更新JPaas CRD中AppRefs集合该资源的status为true
		instance.Spec.AppRefs[i].AppCrStatus = hanwebv1beta1.ConditionAvailable
	}
	return nil
}

func (jc *JPaasCR) Finalizer() error {
	var jpass = jc.paas
	var ctx = jc.context
	var r = jc.kubeClient
	if jpass.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !controllerutil.ContainsFinalizer(jpass, JPaasAppFinalizerName) {
			controllerutil.AddFinalizer(jpass, JPaasAppFinalizerName)
			if err := r.Update(ctx, jpass); err != nil {
				return err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(jpass, JPaasAppFinalizerName) {
			//// our finalizer is present, so lets handle any external dependency
			//if err := jc.JPaasAppFinalizerProcessing(); err != nil {
			//	// if fail to delete the external dependency here, return with error
			//	// so that it can be retried
			//	return err
			//}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(jpass, JPaasAppFinalizerName)
			if err := r.Update(ctx, jpass); err != nil {
				return err
			}
		}

		// Stop reconciliation as the item is being deleted
		return nil
	}
	return nil
}
