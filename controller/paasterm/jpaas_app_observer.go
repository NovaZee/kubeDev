package paasterm

import (
	hanwebv1beta1 "github.com/NovaZee/kubeDev/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"time"
)

// ObservedAppsState holds observed state of a cluster.
type ObservedAppsState struct {
	jpaasApp    *hanwebv1beta1.JPaasApp
	revisions   []*appsv1.ControllerRevision
	configMap   *corev1.ConfigMap
	service     *corev1.Service
	jobPod      *corev1.Pod
	deploy      *appsv1.Deployment
	observeTime time.Time
}
