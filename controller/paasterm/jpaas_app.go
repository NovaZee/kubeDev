package paasterm

import (
	"context"
	"fmt"
	hanwebv1beta1 "github.com/NovaZee/kubeDev/api/v1beta1"
	utils "github.com/NovaZee/kubeDev/controller/util"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

type JPaasAppCR struct {
	kubeClient client.Client
	scheme     *runtime.Scheme
	paasApp    *hanwebv1beta1.JPaasApp
	context    context.Context
	request    ctrl.Request
	log        logr.Logger

	observer ObservedAppsState
}

// JPaasAppFinalizerName name of our custom finalizer
var JPaasAppFinalizerName = "hanweb.jpaas.deploy/finalizer"

// NewJPaasAppCR - Create a new JPaasAppCR.
func NewJPaasAppCR(ctx context.Context, req ctrl.Request, log logr.Logger, c client.Client, scheme *runtime.Scheme) (*JPaasAppCR, error) {

	return &JPaasAppCR{
		context:    ctx,
		request:    req,
		log:        log,
		kubeClient: c,
		scheme:     scheme,
		observer:   ObservedAppsState{},
	}, nil
}

func (jac *JPaasAppCR) PaasAppReconcile() (ctrl.Result, error) {
	// 每一步执行按照先后顺序,如果有错误,则返回错误,整个系统在当前节点存在问题,如果当前节点状态需要提交,则及时更新,避免在下一阶段失败时丢弃(或者在defer中更新)
	// update和status的更新,提交后都会拿到最新的状态,update的结果不会有修改的status的结果,因为提交的是spec不是status

	// step zero: 优先级最高,调合前置判断 初始化/删除 status:state:creating/failed
	//	1: 获取当前的cr,获取不到直接返回,不进行后续操作（后续判断是否在获取不到的时候检查父级CRD中的关联关系,如果存在则进行CR的初始化)
	//	2: 判断DeletionTimestamp,判断当前操作是否为删除操作,如果是删除操作,则执行删除操作,并且返回,不进行后续操作
	// step one: 资源初始化 status:state:deleted
	//	1: 轮询检查资源是否存在,如果不存在,则创建,如果资源存在,不对改资源进行创建操作,轮询结束后会进行下一步操作,
	//     及时更新状态代表当前阶段完成,否则在下一阶段异常时会丢弃当前阶段的状态
	// step two: 运行时调合  status:state:upgrading/running status:health:healthFalse/healthTrue
	//	1: 调合当前资源的状态,平台应用通过health接口即可判断应用是否正常,http调用超时时间为1s,调用health接口返回200则为正常,其他状态码为异常,
	//	   如果异常,则返回错误,延迟15s后再次调用(该时间在java程序初始化的时候,会有一定的延迟,需要较大的延迟时间来回调)
	// step final: 更新状态
	//	1: 更新当前cr的状态,如果更新失败,则返回错误,不进行后续操作
	//	2: 更新当前cr,如果更新失败,则返回错误,不进行后续操作
	var log = jac.log
	// ========step zero:调合前置判断 初始化/删除==========
	var jpaasApp = new(hanwebv1beta1.JPaasApp)
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Errorf("%v", err), "panic")
		}
	}()
	err := jac.inspectionCr(jpaasApp)
	if err != nil {
		// 当前业务应用的cr 不存在
		if errors.IsNotFound(err) {
			log.Error(err, "Failed to get the jpaasApp resource")
			return ctrl.Result{}, nil
		}
		log.Info("inspection Application", "jpaasApp", "nil")
		jpaasApp = nil
		return ctrl.Result{}, err
	} else {
		log.Info("inspection Application", "jpaasApp", *jpaasApp)
		jac.paasApp = jpaasApp
		err = jac.Finalizer()
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	// ========step one:资源初始化==========
	err = jac.ResourceInitCheck()
	if err != nil {
		log.Error(err, "Failed to inspection the current app state")
		return ctrl.Result{}, err
	}
	// 备份,扩容等操作
	err = jac.ResourceRuntimeCheck()
	if err != nil {
		log.Error(err, "Failed to inspection the current app state")
		return ctrl.Result{}, err
	}
	// RUNTIME
	healthy := jac.CheckHealthy()
	if !utils.Healthy(healthy) {
		log.Info("CheckHealthy", "result", "App healthy False,RequeueAfter 15s")
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}
	log.Info("CheckHealthy", "result", healthy)
	if jac.paasApp.Status.Health != healthy {
		jac.paasApp.Status.Health = healthy
		err = jac.UpdateStatus(hanwebv1beta1.ConditionAvailable, hanwebv1beta1.HealthTrue)
		if err != nil {
			return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
		}
	}
	// ========step two==========

	// ========step final更新状态==========
	//err = jac.UpdateStatus(hanwebv1beta1.ConditionAvailable, healthy)
	//if err != nil {
	//	log.Error(err, "Failed to update status app status")
	//	return ctrl.Result{}, err
	//}

	return ctrl.Result{}, nil
}

func (jac *JPaasAppCR) ResourceInitCheck() error {
	// 初始化检查
	var jpaasApp = jac.paasApp
	var log = jac.log
	var ctx = jac.context
	var r = jac.kubeClient
	var err error
	var deployment = &v1.Deployment{}
	var service = &corev1.Service{}
	var deploymentState = jpaasApp.Status.Components.AppDeployment.State
	var svcState = jpaasApp.Status.Components.AppService.State
	var oldSvcPort = jpaasApp.Spec.EmbeddedResource.Ports.FinalPort
	oldStatus := jpaasApp.Status.DeepCopy()
	//deployment
	deploymentErr := jac.inspectionComponents(deployment)
	if deploymentErr == nil {
		jac.observer.deploy = deployment
		if !deployment.DeletionTimestamp.IsZero() {
			deploymentState = hanwebv1beta1.ConditionDeleting
		}
	} else if deploymentErr != nil && errors.IsNotFound(deploymentErr) {
		deployment = utils.NewDeployment(jpaasApp)
		// init deployment
		if err = controllerutil.SetControllerReference(jpaasApp, deployment, jac.scheme); err != nil {
			return err
		}
		jpaasApp.Status.Components.AppDeployment.Name = deployment.Name
		deploymentState = hanwebv1beta1.ConditionCreating
		if err = r.Create(ctx, deployment); err != nil {
			log.Error(err, "create deploy failed")
			deploymentState = hanwebv1beta1.ConditionFailed
			return err
		}
		jpaasApp.Status.Components.AppDeployment.State = deploymentState
	}
	// service
	serviceErr := jac.inspectionService(service)
	if serviceErr == nil {
		jac.observer.service = service
		if !service.DeletionTimestamp.IsZero() {
			svcState = hanwebv1beta1.ConditionDeleting
		}
	} else if serviceErr != nil && errors.IsNotFound(serviceErr) {
		// init svc
		service = utils.NewService(jpaasApp)
		if err = controllerutil.SetControllerReference(jpaasApp, service, jac.scheme); err != nil {
			return err
		}
		jpaasApp.Status.Health = hanwebv1beta1.HealthFalse
		jpaasApp.Status.Components.AppService.Name = service.Name
		svcState = hanwebv1beta1.ConditionCreating
		if err = r.Create(ctx, service); err != nil {
			log.Error(err, "create service failed")
			svcState = hanwebv1beta1.ConditionFailed
			return err
		}
		svcState = hanwebv1beta1.ConditionAvailable
		jpaasApp.Status.Components.AppService.NodePort = utils.GetHealthClusterIpPort(service)
		jpaasApp.Status.Components.AppService.State = svcState
		jpaasApp.Status.Health = hanwebv1beta1.HealthFalse
	}

	if reflect.DeepEqual(*oldStatus, jpaasApp.Status) {
		log.Info(" 资源初始化状态没有变化，不需要再次触发初始化reconcile")
		return nil
	} else {
		jpaasApp.Status.LastUpdateTime = time.Now().Format(time.RFC3339)
		jpaasApp.Status.State = hanwebv1beta1.ConditionAvailable
		err = jac.kubeClient.Status().Update(jac.context, jpaasApp)
		if err != nil {
			log.Error(err, "Failed to update app status")
			return err
		}
		if oldSvcPort != utils.GetHealthClusterIpPort(service) {
			jpaasApp.Spec.EmbeddedResource.Ports.FinalPort = utils.GetHealthClusterIpPort(service)
			err = jac.kubeClient.Update(ctx, jpaasApp)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (jac *JPaasAppCR) ResourceRuntimeCheck() error {
	return nil
}

func (jac *JPaasAppCR) inspectionCr(
	jpaasApp *hanwebv1beta1.JPaasApp) error {
	return jac.kubeClient.Get(
		jac.context, jac.request.NamespacedName, jpaasApp)
}

func (jac *JPaasAppCR) inspectionComponents(
	component *v1.Deployment) error {
	return jac.kubeClient.Get(
		jac.context, jac.request.NamespacedName, component)
}

func (jac *JPaasAppCR) inspectionService(
	service *corev1.Service) error {
	return jac.kubeClient.Get(
		jac.context, jac.request.NamespacedName, service)
}

// Finalizer app cr being deleted modify paas cr  APP status
func (jac *JPaasAppCR) Finalizer() error {
	var jpassApp = jac.paasApp
	var ctx = jac.context
	var r = jac.kubeClient
	// examine DeletionTimestamp to determine if object is under deletion
	if jpassApp.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !controllerutil.ContainsFinalizer(jpassApp, JPaasAppFinalizerName) {
			controllerutil.AddFinalizer(jpassApp, JPaasAppFinalizerName)
			if err := r.Update(ctx, jpassApp); err != nil {
				return err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(jpassApp, JPaasAppFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := jac.JPaasAppFinalizerProcessing(); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(jpassApp, JPaasAppFinalizerName)
			if err := r.Update(ctx, jpassApp); err != nil {
				return err
			}
			jac.paasApp.Status.State = hanwebv1beta1.ConditionDeleted
			if err := r.Status().Update(ctx, jpassApp); err != nil {
				return err
			}
		}

		// Stop reconciliation as the item is being deleted
		return nil
	}

	return nil
}

func (jac *JPaasAppCR) JPaasAppFinalizerProcessing() error {
	var jpassApp, c, ctx = jac.paasApp, jac.kubeClient, jac.context
	// modify paas cr status
	var jpaasList hanwebv1beta1.JPaasList
	listOpts := []client.ListOption{
		client.InNamespace(jpassApp.Namespace),
	}
	if err := c.List(ctx, &jpaasList, listOpts...); err != nil {
		return err
	}
	for _, jpaas := range jpaasList.Items {
		for i, app := range jpaas.Spec.AppRefs {
			if app.Name == jpassApp.Name {
				jpaas.Spec.AppRefs[i].AppCrStatus = hanwebv1beta1.ConditionDeleted
			}
		}
		if err := c.Update(ctx, &jpaas); err != nil {
			return err
		}
	}
	return nil
}

func (jac *JPaasAppCR) CheckHealthy() hanwebv1beta1.Healthy {
	// 创建一个具有超时设置的 http.Client
	client := &http.Client{
		Timeout: time.Second * 1, // 设置超时时间为10秒
	}
	// 获取svc的ip
	url := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d/%s/health", jac.request.Name, jac.request.Namespace, jac.paasApp.Status.Components.AppService.NodePort, jac.request.Name)
	resp, err := client.Get(url)
	if err != nil {
		jac.log.Info("CheckHealthy", "err", err)
		return hanwebv1beta1.HealthFalse
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return hanwebv1beta1.HealthFalse
	}

	return hanwebv1beta1.HealthTrue
}

func (jac *JPaasAppCR) UpdateStatus(state hanwebv1beta1.ConditionType, health hanwebv1beta1.Healthy) error {
	var jpaasApp = jac.paasApp
	var log = jac.log
	jpaasApp.Status.State = state
	jpaasApp.Status.LastUpdateTime = time.Now().Format(time.RFC3339)
	jpaasApp.Status.Health = health
	err := jac.kubeClient.Status().Update(jac.context, jac.paasApp)
	if err != nil {
		log.Error(err, "Failed to update app status")
		return err
	}
	return nil
}
