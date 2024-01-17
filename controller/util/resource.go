package utils

import (
	"bytes"
	"github.com/NovaZee/kubeDev/api/v1beta1"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"text/template"
)

func parseTemplate(templateName string, app *v1beta1.JPaasApp) []byte {
	tmpl, err := template.ParseFiles("controller/template/" + templateName + ".yaml")
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

func NewDeployment(app *v1beta1.JPaasApp) *appv1.Deployment {
	d := &appv1.Deployment{}
	err := yaml.Unmarshal(parseTemplate("deployment", app), d)
	if err != nil {
		panic(err)
	}
	return d
}

func NewService(app *v1beta1.JPaasApp) *corev1.Service {
	s := &corev1.Service{}
	err := yaml.Unmarshal(parseTemplate("service", app), s)
	if err != nil {
		panic(err)
	}
	return s
}

func parseAppTemplate(templateName string, app v1beta1.JPaas) []byte {

	tmpl, err := template.ParseFiles("controller/template/" + templateName + ".yaml")
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

func InitApp(paas v1beta1.JPaas, app v1beta1.AppSpec) *v1beta1.JPaasApp {
	paas.Spec.AppRefs = []v1beta1.AppSpec{}
	paas.Spec.AppRefs = append(paas.Spec.AppRefs, app)
	s := &v1beta1.JPaasApp{}
	err := yaml.Unmarshal(parseAppTemplate("jpaasapp", paas), s)
	if err != nil {
		panic(err)
	}
	return s
}
