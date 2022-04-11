package kubernetes

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/drain"
)

// MinMasters - Minimum number of master nodes the cluster can survive with.
const MinMasters = 1

// MinWorkers - Minimum number of workers nodes the cluster can survive with.
const MinWorkers = 2

// UtilsClient - Structure for kubernetes client.
type UtilsClient struct {
	Logger kubernetesLogger

	clientSet *kubernetes.Clientset
	helper    *drain.Helper
}

// NodeMap - Data type to hold a map between NCN name and Kubernetes node pointer.
type NodeMap map[string]corev1.Node
