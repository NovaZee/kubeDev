package client

import (
	"context"
	hanwebv1beta1 "github.com/NovaZee/kubeDev/api/v1beta1"
	utils "github.com/NovaZee/kubeDev/internal/util"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	if !jc.paas.Spec.Initialized {
		return jc.Init()
	}
	err := jc.checkAppCr()
	if err != nil {
		return ctrl.Result{}, err
	}
	// crd列表中有 app的crd not found 代表 app的crd没有创建成功或者被删除了

	return ctrl.Result{}, nil
}

func (jc *JPaasCR) Init() (ctrl.Result, error) {
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
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
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
