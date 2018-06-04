package stub

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/piontec/netperf-operator/pkg/apis/app/v1alpha1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	NetperfTypeServer = "server"
	NetperfTypeClient = "client"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.Netperf:
		logrus.Debugf("New Netperf event, name: %s, deleted: %v, status: %v", o.Name, event.Deleted, o.Status.Status)
		if event.Deleted {
			return h.deleteNetperfPods(o)
		}
		return h.handleNetperfUpdateEvent(o)
	case *v1.Pod:
		pod := event.Object.(*v1.Pod)
		if pod.ObjectMeta.OwnerReferences[0].Kind != "Netperf" {
			return nil
		}
		if pod.ObjectMeta.OwnerReferences[0].UID == "" {
			logrus.Warnf("Pod %s/%s has owner of type Netperf, but UID is unknown", pod.Namespace, pod.Name)
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
		return h.startServerPod(cr)
	default:
		logrus.Debugf("Nothing needed to do for update event on Netperf %s in state %s",
			cr.Name, cr.Status.Status)
		return nil
	}
}

func (h *Handler) startServerPod(cr *v1alpha1.Netperf) error {
	serverPod := h.newNetperfPod(cr, NetperfTypeServer, v1.RestartPolicyAlways, []string{})

	err := sdk.Create(serverPod)
	if err != nil && !errors.IsAlreadyExists(err) {
		logrus.Errorf("Failed to create server pod : %v", err)
		return err
	}
	if err != nil && errors.IsAlreadyExists(err) {
		logrus.Debugf("Server pod is already created for netperf: %v", cr.Name)
		if cr.Status.ServerPod == serverPod.Name {
			logrus.Debugf("Server pod %v already registered for netperf %v", serverPod.Name,
				cr.Name)
			return nil
		}
	} else {
		logrus.Debugf("New server pod started for netperf: %s", cr.Name)
	}

	if err := h.registerNetperfServer(cr, serverPod); err != nil {
		return err
	}
	return nil
}

func (h *Handler) getPodAffinity(cr *v1alpha1.Netperf, netperfType string) *v1.Affinity {
	if (netperfType == NetperfTypeClient && cr.Spec.ClientNode == "") ||
		(netperfType == NetperfTypeServer && cr.Spec.ServerNode == "") {
		return nil
	}

	nodeName := ""
	if netperfType == NetperfTypeClient {
		nodeName = cr.Spec.ClientNode
	} else if netperfType == NetperfTypeServer {
		nodeName = cr.Spec.ServerNode
	} else {
		logrus.Errorf("Unexpected netperf pod type %s. This should never happen", netperfType)
	}
	return &v1.Affinity{
		NodeAffinity: &v1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
				NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{
							{
								Key:      "TODO",
								Operator: "equals",
								Values:   []string{nodeName},
							},
						},
					},
				},
			},
		},
	}
}

func (h *Handler) getNetperfPodName(cr *v1alpha1.Netperf, netperfType string) string {
	var name string
	guidString := fmt.Sprint(cr.UID)
	suffix := strings.Split(guidString, "-")[4]
	switch netperfType {
	case NetperfTypeClient:
		name = "netperf-client-" + suffix
	case NetperfTypeServer:
		name = "netperf-server-" + suffix
	}
	return name
}

func (h *Handler) newNetperfPod(cr *v1alpha1.Netperf, netperfType string, restartPolicy v1.RestartPolicy, command []string) *v1.Pod {
	name := h.getNetperfPodName(cr, netperfType)
	labels := map[string]string{
		"app":          "netperf-operator",
		"netperf-type": netperfType,
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
			RestartPolicy: restartPolicy,
		},
	}
	return pod
}

func (h *Handler) registerNetperfServer(cr *v1alpha1.Netperf, serverPod *v1.Pod) error {
	c := cr.DeepCopy()
	c.Status.Status = v1alpha1.NetperfPhaseServer
	c.Status.ServerPod = serverPod.Name
	return sdk.Update(c)
}

func (h *Handler) handlePodUpdateEvent(pod *v1.Pod) error {
	logrus.Debugf("Trying to bind the pod %s with CR", pod.Name)
	cr := &v1alpha1.Netperf{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Netperf",
			APIVersion: "app.example.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.OwnerReferences[0].Name,
			Namespace: pod.Namespace,
		},
	}
	if err := sdk.Get(cr); err != nil {
		logrus.Errorf("error trying to fetch Netperf object %s/%s defined as owner of pod %s/%s: %v",
			pod.Namespace, pod.OwnerReferences[0].Name, pod.Namespace, pod.Name, err)
		return nil
	}

	isServerPod := pod.Name == cr.Status.ServerPod
	isClientPod := pod.Name == cr.Status.ClientPod
	if !isServerPod && !isClientPod {
		logrus.Errorf("pod with UID %s was not detected as server nor client pod of the CR",
			pod.UID)
		return nil
	}

	if isClientPod {
		logrus.Debugf("This is client pod event for CR %s: pod phase: %s, pod IP: %s",
			cr.Name, pod.Status.Phase, pod.Status.PodIP)
		return h.handleClientPodEvent(cr, pod)
	}
	if isServerPod {
		logrus.Debugf("This is server pod event for CR %s: pod phase: %s, pod IP: %s",
			cr.Name, pod.Status.Phase, pod.Status.PodIP)
		return h.handleServerPodEvent(cr, pod)
	}

	return nil
}

func (h *Handler) handleClientPodEvent(cr *v1alpha1.Netperf, pod *v1.Pod) error {
	if pod.Status.Phase == v1.PodRunning {
		logrus.Debugf("Client pod is running")
		return nil
	}

	if pod.Status.Phase == v1.PodSucceeded && cr.Status.Status != v1alpha1.NetperfPhaseDone {
		logrus.Debugf("Test completed, parsing results")
		res := h.getLogFromClientPod(pod)
		lines := strings.Split(res, "\n")
		entries := strings.Fields(lines[6])
		throughput, convErr := strconv.ParseFloat(entries[4], 64)
		if convErr != nil {
			return fmt.Errorf("error trying to convert result \"%s\" to float: %v", entries[4], convErr)
		}

		serverPod, err := h.getPodByName(cr.Status.ServerPod, cr.Namespace)
		if err != nil {
			logrus.Errorf("Error fetching pod %v by name: %v. Won't delete Netperf.", cr.Status.ServerPod, err)
			return err
		}
		logrus.Debug("Test completed, deleting resources")
		if err = sdk.Delete(pod); err != nil {
			logrus.Debugf("Error deleting client pod %v: %v", pod.Name, err)
			return err
		}
		if err = sdk.Delete(serverPod); err != nil {
			logrus.Debugf("Error deleting server pod %v: %v", serverPod.Name, err)
			return err
		}
		netperf := cr.DeepCopy()
		netperf.Status.SpeedBitsPerSec = throughput
		netperf.Status.Status = v1alpha1.NetperfPhaseDone
		return sdk.Update(netperf)
	}

	return nil
}

func (h *Handler) getPodByName(name, namespace string) (*v1.Pod, error) {
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := sdk.Get(pod)
	return pod, err
}

func (h *Handler) getLogFromClientPod(pod *v1.Pod) string {
	//FIXME: replace this with some call through operator SDK to use existing pooled connection
	var kubeconfig string
	if home := os.Getenv("HOME"); home != "" {
		kubeconfig = filepath.Join(home, ".kube/config")
	}
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		logrus.Errorf("Wrong config: %v", err)
	}
	client, err := corev1client.NewForConfig(cfg)
	if err != nil {
		logrus.Errorf("Client error: %v", err)
	}
	logOptions := &v1.PodLogOptions{}
	req := client.Pods(pod.Namespace).GetLogs(pod.Name, logOptions)
	rc, err := req.Stream()
	if err != nil {
		logrus.Errorf("Client error: %v", err)
	}
	defer rc.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(rc)
	s := buf.String()
	return s
}

func (h *Handler) handleServerPodEvent(cr *v1alpha1.Netperf, pod *v1.Pod) error {
	if pod.Status.Phase != v1.PodRunning {
		logrus.Debugf("Server pod is not running yet")
		return nil
	}

	if cr.Status.ClientPod != "" {
		logrus.Debugf("It seems client pod as already created, skipping creation")
		return nil
	}

	logrus.Debugf("Creating client pod for netperf: %v", cr.Name)
	clientPod := h.newNetperfPod(cr, NetperfTypeClient, v1.RestartPolicyOnFailure, []string{"netperf", "-H", pod.Status.PodIP})
	err := sdk.Create(clientPod)
	if err != nil && !errors.IsAlreadyExists(err) {
		logrus.Errorf("Failed to create client pod : %v", err)
		return err
	}
	if err != nil && errors.IsAlreadyExists(err) {
		logrus.Debugf("client pod already created for netperf: %s", cr.Name)
		if cr.Status.ClientPod == clientPod.Name {
			logrus.Debugf("Client pod %v already registered with netperf %v", clientPod.Name,
				cr.Name)
			return nil
		}
	} else {
		logrus.Debugf("New client pod started: %s/%s", clientPod.Namespace, clientPod.Name)
	}
	c := cr.DeepCopy()
	c.Status.Status = v1alpha1.NetperfPhaseTest
	c.Status.ClientPod = clientPod.Name
	sdk.Update(c)
	logrus.Debugf("Custom resource %s updated with client pod info: %s", cr.Name, clientPod.Name)

	return nil
}
