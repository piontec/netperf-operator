package operator

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/piontec/netperf-operator/pkg/apis/app/kube"
	"github.com/piontec/netperf-operator/pkg/apis/app/v1alpha1"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type netperfType string

const (
	netperfTypeServer netperfType = "server"
	netperfTypeClient netperfType = "client"
	netperfImage                  = "tailoredcloud/netperf:v2.7"
)

type Netperfer interface {
	HandleNetperf(*v1alpha1.Netperf, bool) error
	HandlePod(*v1.Pod, bool) error
}

type Netperf struct {
	provider kube.Provider
}

func NewNetperf(provider kube.Provider) Netperfer {
	return &Netperf{
		provider: provider,
	}
}

func (n *Netperf) HandleNetperf(o *v1alpha1.Netperf, deleted bool) error {
	logrus.Debugf("New Netperf event, name: %s, deleted: %v, status: %v", o.Name, deleted, o.Status.Status)
	if deleted {
		return n.deleteNetperfPods(o)
	}
	return n.handleNetperfUpdateEvent(o)
}

func (n *Netperf) HandlePod(pod *v1.Pod, deleted bool) error {
	if pod.ObjectMeta.OwnerReferences[0].Kind != "Netperf" {
		return nil
	}
	if pod.ObjectMeta.OwnerReferences[0].UID == "" {
		logrus.Warnf("Pod %s/%s has owner of type Netperf, but UID is unknown", pod.Namespace, pod.Name)
	}
	logrus.Debugf("New pod event: %s/%s, deleted status: %v", pod.Namespace, pod.Name, deleted)
	return n.handlePodUpdateEvent(pod)
}

func (n *Netperf) deleteNetperfPods(cr *v1alpha1.Netperf) error {
	logrus.Debugf("Netperf object %s/%s is being deleted", cr.Namespace, cr.Name)
	return nil
}

func (n *Netperf) handleNetperfUpdateEvent(cr *v1alpha1.Netperf) error {
	switch cr.Status.Status {
	case v1alpha1.NetperfPhaseInitial:
		return n.startServerPod(cr)
	case v1alpha1.NetperfPhaseServer:
		return n.startServerPod(cr)
	default:
		logrus.Debugf("Nothing needed to do for update event on Netperf %s in state %s",
			cr.Name, cr.Status.Status)
		return nil
	}
}

func (n *Netperf) startServerPod(cr *v1alpha1.Netperf) error {
	serverPod := n.newNetperfPod(cr, netperfTypeServer, v1.RestartPolicyAlways, []string{})

	err := n.provider.Create(serverPod)
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

	if err := n.registerNetperfServer(cr, serverPod); err != nil {
		return err
	}
	return nil
}

func (n *Netperf) getNetperfPodAffinity(cr *v1alpha1.Netperf, npType netperfType) *v1.Affinity {
	if (npType == netperfTypeClient && cr.Spec.ClientNode == "") ||
		(npType == netperfTypeServer && cr.Spec.ServerNode == "") {
		return nil
	}

	nodeName := ""
	if npType == netperfTypeClient {
		nodeName = cr.Spec.ClientNode
	} else if npType == netperfTypeServer {
		nodeName = cr.Spec.ServerNode
	} else {
		logrus.Errorf("Unexpected netperf pod type %s. This should never happen", npType)
	}
	return &v1.Affinity{
		NodeAffinity: &v1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
				NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{
							{
								Key:      "kubernetes.io/hostname",
								Operator: "In",
								Values:   []string{nodeName},
							},
						},
					},
				},
			},
		},
	}
}

func (n *Netperf) getNetperfPodName(cr *v1alpha1.Netperf, npType netperfType) string {
	var name string
	guidString := fmt.Sprint(cr.UID)
	suffix := strings.Split(guidString, "-")[4]
	switch npType {
	case netperfTypeClient:
		name = "netperf-client-" + suffix
	case netperfTypeServer:
		name = "netperf-server-" + suffix
	}
	return name
}

func (n *Netperf) newNetperfPod(cr *v1alpha1.Netperf, npType netperfType, restartPolicy v1.RestartPolicy, command []string) *v1.Pod {
	name := n.getNetperfPodName(cr, npType)
	affinity := n.getNetperfPodAffinity(cr, npType)
	labels := map[string]string{
		"app":          "netperf-operator",
		"netperf-type": fmt.Sprint(npType),
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
					Image:   netperfImage,
					Command: command,
				},
			},
			RestartPolicy: restartPolicy,
			Affinity:      affinity,
		},
	}
	return pod
}

func (n *Netperf) registerNetperfServer(cr *v1alpha1.Netperf, serverPod *v1.Pod) error {
	c := cr.DeepCopy()
	c.Status.Status = v1alpha1.NetperfPhaseServer
	c.Status.ServerPod = serverPod.Name
	return n.provider.Update(c)
}

func (n *Netperf) getNetperfByName(name, namespace string) (*v1alpha1.Netperf, error) {
	cr := &v1alpha1.Netperf{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Netperf",
			APIVersion: "app.example.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	if err := n.provider.Get(cr); err != nil {
		return nil, err
	}
	return cr, nil
}

func (n *Netperf) handlePodUpdateEvent(pod *v1.Pod) error {
	logrus.Debugf("Trying to bind the pod %s with CR", pod.Name)
	var cr *v1alpha1.Netperf
	var err error
	if cr, err = n.getNetperfByName(pod.OwnerReferences[0].Name, pod.Namespace); err != nil {
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
		return n.handleClientPodEvent(cr, pod)
	}
	if isServerPod {
		logrus.Debugf("This is server pod event for CR %s: pod phase: %s, pod IP: %s",
			cr.Name, pod.Status.Phase, pod.Status.PodIP)
		return n.handleServerPodEvent(cr, pod)
	}

	return nil
}

func (n *Netperf) handleClientPodEvent(cr *v1alpha1.Netperf, pod *v1.Pod) error {
	if pod.Status.Phase == v1.PodRunning {
		logrus.Debugf("Client pod is running")
		return nil
	}

	if pod.Status.Phase == v1.PodSucceeded && cr.Status.Status != v1alpha1.NetperfPhaseDone {
		logrus.Debugf("Test completed, parsing results")
		res := n.getLogFromClientPod(pod)
		throughput, convErr := n.parseNetperfResult(res)
		if convErr != nil {
			n.updateNetperfStatus(cr, v1alpha1.NetperfPhaseError)
			return fmt.Errorf("error trying to convert test result to float: %v", convErr)
		}

		serverPod, err := n.getPodByName(cr.Status.ServerPod, cr.Namespace)
		if err != nil {
			n.updateNetperfStatus(cr, v1alpha1.NetperfPhaseError)
			logrus.Errorf("Error fetching pod %v by name: %v. Won't delete Netperf.", cr.Status.ServerPod, err)
			return err
		}
		logrus.Debug("Test completed, deleting resources")
		if err = n.provider.Delete(pod); err != nil {
			n.updateNetperfStatus(cr, v1alpha1.NetperfPhaseError)
			logrus.Debugf("Error deleting client pod %v: %v", pod.Name, err)
			return err
		}
		if err = n.provider.Delete(serverPod); err != nil {
			n.updateNetperfStatus(cr, v1alpha1.NetperfPhaseError)
			logrus.Debugf("Error deleting server pod %v: %v", serverPod.Name, err)
			return err
		}
		netperf := cr.DeepCopy()
		netperf.Status.SpeedBitsPerSec = throughput
		netperf.Status.Status = v1alpha1.NetperfPhaseDone
		return n.provider.Update(netperf)
	}

	return nil
}

func (n *Netperf) updateNetperfStatus(resource *v1alpha1.Netperf, status string) error {
	netperf := resource.DeepCopy()
	netperf.Status.Status = v1alpha1.NetperfPhaseDone
	return n.provider.Update(netperf)
}

func (n *Netperf) parseNetperfResult(result string) (float64, error) {
	lines := strings.Split(result, "\n")
	if len(lines) < 7 {
		return 0, fmt.Errorf("Bad netperf command output")
	}
	entries := strings.Fields(lines[6])
	return strconv.ParseFloat(entries[4], 64)
}

func (n *Netperf) getPodByName(name, namespace string) (*v1.Pod, error) {
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
	err := n.provider.Get(pod)
	return pod, err
}

func (n *Netperf) getLogFromClientPod(pod *v1.Pod) string {
	client := n.provider.GetKubeClient()
	logOptions := &v1.PodLogOptions{}
	req := client.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, logOptions)
	rc, err := req.Stream()
	if err != nil {
		logrus.Errorf("Client error: %v", err)
	}
	defer rc.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(rc)
	return buf.String()
}

func (n *Netperf) handleServerPodEvent(cr *v1alpha1.Netperf, pod *v1.Pod) error {
	if pod.Status.Phase != v1.PodRunning {
		logrus.Debugf("Server pod is not running yet")
		return nil
	}

	if cr.Status.ClientPod != "" {
		logrus.Debugf("It seems client pod as already created, skipping creation")
		return nil
	}

	logrus.Debugf("Creating client pod for netperf: %v", cr.Name)
	clientPod := n.newNetperfPod(cr, netperfTypeClient, v1.RestartPolicyOnFailure, []string{"netperf", "-H", pod.Status.PodIP})
	err := n.provider.Create(clientPod)
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
	n.provider.Update(c)
	logrus.Debugf("Custom resource %s updated with client pod info: %s", cr.Name, clientPod.Name)

	return nil
}
