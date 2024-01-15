package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"
)

type JPaasHandler struct {
	//k8sServices k8s.Services todo: k8s内置资源操作
	Check JPaasCheck
}

type Reconciler interface {
	PaasReconcile() (ctrl.Result, error)
}

type Process interface {
	InspectionInit() error
	Finalizer() error
}
