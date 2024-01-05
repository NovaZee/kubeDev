package utils

import (
	"bytes"
	"github.com/NovaZee/kubeDev/api/v1beta1"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"text/template"
)

func parseTemplate(templateName string, app *v1beta1.JPaas) []byte {
	tmpl, err := template.ParseFiles("internal/template/" + templateName + ".yaml")
	if err != nil {
		panic(err)
	}
	b := new(bytes.Buffer)
	err = tmpl.Execute(b, app)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

func NewDeployment(app *v1beta1.JPaas) *appv1.Deployment {
	d := &appv1.Deployment{}
	err := yaml.Unmarshal(parseTemplate("deployment", app), d)
	if err != nil {
		panic(err)
	}
	return d
}

func NewService(app *v1beta1.JPaas) *corev1.Service {
	s := &corev1.Service{}
	err := yaml.Unmarshal(parseTemplate("service", app), s)
	if err != nil {
		panic(err)
	}
	return s
}

func NewCrd(paas *v1beta1.JPaas) *v1beta1.JPaas {
	s := &v1beta1.JPaas{}
	err := yaml.Unmarshal(parseTemplate("jpaas", paas), s)
	if err != nil {
		panic(err)
	}
	return s
}

func InitCrd(paas *v1beta1.JPaas, app *v1beta1.AppRef) *v1beta1.JPaas {
	var paasType v1beta1.Type
	switch paas.Spec.Type {
	case v1beta1.Platform:
		paasType = v1beta1.BaseApp
	case v1beta1.BaseApp:
		paasType = v1beta1.Application
	case v1beta1.Application:
		paasType = v1beta1.Application
	}
	s := &v1beta1.JPaas{
		Spec: v1beta1.JPaasSpec{
			Namespace:      paas.Spec.Namespace,
			Image:          app.Image,
			Replicas:       paas.Spec.Replicas,
			Version:        app.Version,
			Env:            paas.Spec.Env,
			VersionAligned: paas.Spec.VersionAligned,
			Type:           paasType,
		},
	}
	s.Name = app.Name
	return NewCrd(s)
}
