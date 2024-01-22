package utils

import (
	hanwebv1beta1 "github.com/NovaZee/kubeDev/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

func Healthy(health hanwebv1beta1.Healthy) bool {
	if health == hanwebv1beta1.HealthTrue {
		return true
	}
	return false
}

func GetHealthClusterIpPort(s *corev1.Service) int32 {
	for _, port := range s.Spec.Ports {
		if port.Name == "health" {
			return port.TargetPort.IntVal
			//测试 使用NodePort
			//return port.NodePort
		}
	}
	return 0
}
