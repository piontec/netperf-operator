package kube

import "k8s.io/client-go/kubernetes"
import "k8s.io/apimachinery/pkg/runtime"

type Provider interface {
	Create(object runtime.Object) error
	Update(object runtime.Object) error
	Get(object runtime.Object) error
	Delete(object runtime.Object) error
	GetKubeClient() kubernetes.Interface
}
