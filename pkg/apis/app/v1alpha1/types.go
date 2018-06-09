package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NetperfPhaseInitial = ""
	NetperfPhaseServer  = "Created server pod"
	NetperfPhaseTest    = "Started test"
	NetperfPhaseDone    = "Done"
	NetperfPhaseError   = "Test finished with error"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type NetperfList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Netperf `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Netperf struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              NetperfSpec   `json:"spec"`
	Status            NetperfStatus `json:"status,omitempty"`
}

type NetperfSpec struct {
	ServerNode string `json:"serverNode"`
	ClientNode string `json:"clientNode"`
}
type NetperfStatus struct {
	Status          string  `json:"status"`
	ServerPod       string  `json:"serverPod"`
	ClientPod       string  `json:"clientPod"`
	SpeedBitsPerSec float64 `json:"speedBitsPerSec"`
}
