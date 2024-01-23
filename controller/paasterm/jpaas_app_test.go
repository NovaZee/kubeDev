package paasterm

import (
	hanwebv1beta1 "github.com/NovaZee/kubeDev/api/v1beta1"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

func TestJPaasAppCR_CheckHealthy(t *testing.T) {
	tests := []struct {
		name     string
		response int
		want     hanwebv1beta1.Healthy
	}{
		{
			name:     "Healthy when status OK",
			response: http.StatusOK,
			want:     hanwebv1beta1.HealthTrue,
		},
		{
			name:     "Unhealthy when status not OK",
			response: http.StatusNotFound,
			want:     hanwebv1beta1.HealthFalse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建一个模拟的 HTTP 服务器
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "test/health" {
					w.WriteHeader(http.StatusOK)
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			// 创建一个 JPaasAppCR 实例
			jac := &JPaasAppCR{
				kubeClient: client.NewDryRunClient(nil),
				//log: logr{T: t},
				paasApp: &hanwebv1beta1.JPaasApp{
					Status: hanwebv1beta1.JPaasAppStatus{
						Components: hanwebv1beta1.ComponentsStatus{
							AppService: hanwebv1beta1.AppComponentServiceStatus{
								NodePort: 8080,
							},
						},
					},
				},
				request: reconcile.Request{
					NamespacedName: client.ObjectKey{
						Namespace: "default",
						Name:      "test",
					},
				},
			}

			// 调用 CheckHealthy 方法并检查结果
			if got := jac.CheckHealthy(); got != tt.want {
				t.Errorf("JPaasAppCR.CheckHealthy() = %v, want %v", got, tt.want)
			}
		})
	}
}
