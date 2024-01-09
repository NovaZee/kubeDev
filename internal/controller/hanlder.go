package controller

import (
	"context"
	hanwebv1beta1 "github.com/NovaZee/kubeDev/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type JPaasHandler struct {
	//k8sServices k8s.Services todo: k8s内置资源操作
	Check JPaasCheck
}

type Reconciler interface {
	InitReconcile(ctx context.Context, instance *hanwebv1beta1.JPaas, req ctrl.Request) (ctrl.Result, error)
}
