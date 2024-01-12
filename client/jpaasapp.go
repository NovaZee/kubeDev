package client

import (
	"context"
	hanwebv1beta1 "github.com/NovaZee/kubeDev/api/v1beta1"
	utils "github.com/NovaZee/kubeDev/internal/util"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type JPaasAppCR struct {
	kubeClient client.Client
	scheme     *runtime.Scheme
	paasApp    *hanwebv1beta1.JPaasApp
	context    context.Context
	request    ctrl.Request
	log        logr.Logger
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
	}, nil
}

func (jac *JPaasAppCR) PaasAppReconcile() (ctrl.Result, error) {
	var log = jac.log
	// ========step one==========
	err := jac.inspectionInit()
	if err != nil {
		log.Error(err, "Failed to inspection the current app state")
		return ctrl.Result{}, err
	}
	// cr被删除了，需要重新创建
	// ========step two==========
	return ctrl.Result{}, nil
}

func (jac *JPaasAppCR) inspectionInit() error {
	var jpaasApp = new(hanwebv1beta1.JPaasApp)
	var log = jac.log
	var ctx = jac.context
	var r = jac.kubeClient
	err := jac.inspectionCr(jpaasApp)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Error(err, "Failed to get the jpaasApp resource")
			return nil
		}
		log.Info("inspection Application", "jpaasApp", "nil")
		jpaasApp = nil
		//cr被删除了,更改jpaas cr 中该app status
	} else {
		log.Info("inspection Application", "jpaasApp", *jpaasApp)
		jac.paasApp = jpaasApp
		err = jac.finalizer()
		if err != nil {
			return err
		}
	}
	var deployment = &v1.Deployment{}
	var service = &corev1.Service{}
	deploymentErr := jac.inspectionComponents(deployment)
	if deploymentErr != nil && errors.IsNotFound(deploymentErr) {
		deployment = utils.NewDeployment(jpaasApp)
		// init deployment
		if err = controllerutil.SetControllerReference(jpaasApp, deployment, jac.scheme); err != nil {
			return err
		}
		if err = r.Create(ctx, deployment); err != nil {
			log.Error(err, "create deploy failed")
			return err
		}
	}
	serviceErr := jac.inspectionService(service)
	if serviceErr != nil && errors.IsNotFound(serviceErr) {
		// init svc
		service = utils.NewService(jpaasApp)
		if err = controllerutil.SetControllerReference(jpaasApp, service, jac.scheme); err != nil {
			return err
		}
		if err = r.Create(ctx, service); err != nil {
			log.Error(err, "create service failed")
			return err
		}
	}
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

// ToOwnerReference the paas as owner reference for its child resources.
func ToOwnerReference(
	paasApp *hanwebv1beta1.JPaasApp) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion:         paasApp.APIVersion,
		Kind:               paasApp.Kind,
		Name:               paasApp.Name,
		UID:                paasApp.UID,
		Controller:         &[]bool{true}[0],
		BlockOwnerDeletion: &[]bool{false}[0],
	}
}

func (jac *JPaasAppCR) finalizer() error {
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
