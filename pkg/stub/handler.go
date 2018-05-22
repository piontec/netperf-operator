package stub

import (
	"context"

	"github.com/piontec/netperf-operator/pkg/apis/app/v1alpha1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.Netperf:
		logrus.Debugf("New Netperf event, name: %s, deleted: %v", o.Name, event.Deleted)
		if event.Deleted {
			return h.deleteNetperfPods(o)
		}
		return h.handleNetperfUpdateEvent(o)
	case *v1.Pod:
		pod := event.Object.(*v1.Pod)
		if pod.ObjectMeta.OwnerReferences[0].Kind != "Netperf" {
			return nil
		}
		logrus.Debugf("New pod event: %s/%s, deleted status: %v", pod.Namespace, pod.Name, event.Deleted)
		return h.handlePodUpdateEvent(pod)
	default:
		logrus.Warnf("unknown event received: %s", event)
	}
	return nil
}

func (h *Handler) deleteNetperfPods(cr *v1alpha1.Netperf) error {
	//TODO: implement
	logrus.Fatalf("deleteNetperfPods Not implemented")
	return nil
}

func (h *Handler) handleNetperfUpdateEvent(cr *v1alpha1.Netperf) error {
	switch cr.Status.Status {
	case v1alpha1.NetperfPhaseInitial:
		return h.startServerPod(cr)
	case v1alpha1.NetperfPhaseServer:
		return h.startClientPod(cr)
	default:
		logrus.Debugf("Nothing needed to do for update even on Netperf %s in state %s",
			cr.Name, cr.Status.Status)
		return nil
	}
}

func (h *Handler) startServerPod(cr *v1alpha1.Netperf) error {
	serverPod := h.newNetperfPod(cr, "netperf-server", []string{})

	err := sdk.Create(serverPod)
	if err != nil && !errors.IsAlreadyExists(err) {
		logrus.Errorf("Failed to create server pod : %v", err)
		return err
	}

	if err := h.registerNetperfServer(cr, serverPod); err != nil {
		return err
	}
	logrus.Debug("New server pod started and registered for netperf: %s", cr.Name)
	return nil
}

func (h *Handler) newNetperfPod(cr *v1alpha1.Netperf, name string, command []string) *v1.Pod {
	labels := map[string]string{
		"app": "netperf-operator",
	}
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: cr.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cr, schema.GroupVersionKind{
					Group:   v1alpha1.SchemeGroupVersion.Group,
					Version: v1alpha1.SchemeGroupVersion.Version,
					Kind:    "Netperf",
				}),
			},
			Labels: labels,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:    name,
					Image:   "alectolytic/netperf",
					Command: command,
				},
			},
		},
	}
	return pod
}

func (h *Handler) registerNetperfServer(cr *v1alpha1.Netperf, serverPod *v1.Pod) error {
	cr.Status.Status = v1alpha1.NetperfPhaseServer
	cr.Status.ServerPod = serverPod.UID
	return sdk.Update(cr)
}

func (h *Handler) handlePodUpdateEvent(pod *v1.Pod) error {
	logrus.Debugf("Trying to bind the pod %s with CR", pod.Name)
	// var serverPodOwner *NetperfInfo = nil

	// for _, info := range h.infos {
	// 	if pod.OwnerReferences[0].UID != info.customResource.UID {
	// 		continue
	// 	}
	// 	if info.clientPod != nil && pod.UID == info.clientPod.UID {
	// 		logrus.Debugf("This is client pod event for CR %s: pod phase: %s, pod IP: %s",
	// 			info.customResource.Name, pod.Status.Phase, pod.Status.PodIP)
	// 	}
	// 	if pod.UID == info.serverPod.UID {
	// 		logrus.Debugf("This is server pod event for CR %s: pod phase: %s, pod IP: %s",
	// 			info.customResource.Name, pod.Status.Phase, pod.Status.PodIP)
	// 		if pod.Status.Phase == v1.PodRunning {
	// 			logrus.Infof("Server pod %s is running with IP %s, starting client pod",
	// 				pod.Name, pod.Status.PodIP)
	// 			serverPodOwner = &info
	// 			break
	// 		}
	// 	}
	// }

	// if serverPodOwner != nil {
	// 	clientPod := h.newNetperfPod(serverPodOwner.customResource, "netperf-client", []string{"netperf", "-H", pod.Status.PodIP})

	// 	err := sdk.Create(clientPod)
	// 	if err != nil && !errors.IsAlreadyExists(err) {
	// 		logrus.Errorf("Failed to create client pod : %v", err)
	// 		return err
	// 	}

	// 	serverPodOwner.clientPod = clientPod
	// 	logrus.Debug("New server pod started and registered for netperf: %s", serverPodOwner.customResource.Name)
	// 	return nil
	// }
	return nil
}
