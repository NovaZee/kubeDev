package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"
)

type Reconciler interface {
	PaasReconcile() (ctrl.Result, error)
}

type Process interface {
	InspectionInit() error
	Finalizer() error
}
