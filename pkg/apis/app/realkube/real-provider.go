package realkube

import (
	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/piontec/netperf-operator/pkg/apis/app/kube"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

type RealProvider struct {
}

func NewRealProvider() kube.Provider {
	return &RealProvider{}
}

func (r *RealProvider) Create(object runtime.Object) error {
	return sdk.Create(object)
}

func (r *RealProvider) Update(object runtime.Object) error {
	return sdk.Update(object)
}

func (r *RealProvider) Get(object runtime.Object) error {
	return sdk.Get(object)
}

func (r *RealProvider) Delete(object runtime.Object) error {
	return sdk.Delete(object)
}

func (r *RealProvider) GetKubeClient() kubernetes.Interface {
	return k8sclient.GetKubeClient()
}
