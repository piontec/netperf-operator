package fakekube

import (
	"github.com/piontec/netperf-operator/pkg/apis/app/kube"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

type FakeProvider struct {
}

func NewFakeProvider() kube.Provider {
	return &FakeProvider{}
}

func (r *FakeProvider) Create(object runtime.Object) error {
	return nil
}

func (r *FakeProvider) Update(object runtime.Object) error {
	return nil
}

func (r *FakeProvider) Get(object runtime.Object) error {
	return nil
}

func (r *FakeProvider) Delete(object runtime.Object) error {
	return nil
}

func (r *FakeProvider) GetKubeClient() kubernetes.Interface {
	return nil
}
