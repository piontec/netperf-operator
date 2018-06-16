package stub

import (
	"context"

	"github.com/operator-framework/operator-sdk/pkg/sdk"

	"github.com/piontec/netperf-operator/pkg/apis/app/v1alpha1"
	operator "github.com/piontec/netperf-operator/pkg/netperf-operator"

	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
)

func NewHandler(operator operator.Netperfer) sdk.Handler {
	return &Handler{
		operator: operator,
	}
}

type Handler struct {
	operator operator.Netperfer
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch event.Object.(type) {
	case *v1alpha1.Netperf:
		netperf := event.Object.(*v1alpha1.Netperf)
		return h.operator.HandleNetperf(netperf, event.Deleted)
	case *v1.Pod:
		pod := event.Object.(*v1.Pod)
		return h.operator.HandlePod(pod, event.Deleted)
	default:
		logrus.Warnf("unknown event received: %s", event)
	}
	return nil
}
