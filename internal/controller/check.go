package controller

import (
	"fmt"
	"github.com/NovaZee/kubeDev/api/v1beta1"
	"github.com/NovaZee/kubeDev/client/jpaas"
	"github.com/go-logr/logr"
	"io"
	"net/http"
	"os"
)

type JPaasCheck interface {
	CheckHealthy(JPaasCluster *v1beta1.JPaas) (bool, error)
}

type JPaasChecker struct {
	jpaas  jpaas.Client
	logger logr.Logger
}

func NewJPaasChecker(logger logr.Logger) *JPaasChecker {
	return &JPaasChecker{
		jpaas:  jpaas.New(),
		logger: logger,
	}
}

func (jc JPaasChecker) CheckHealthy(JPaasCluster *v1beta1.JPaas) (bool, error) {
	podIP := os.Getenv("POD_IP")
	resp, err := http.Get(fmt.Sprintf("http://%s/health", podIP))
	if err != nil {
		return false, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %v", err)
	}

	if string(body) != "UP" {
		return false, fmt.Errorf("service is not healthy: %s", body)
	}

	return true, nil
}
