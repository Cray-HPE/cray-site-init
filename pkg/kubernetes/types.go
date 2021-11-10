package kubernetes

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/drain"
)

const MIN_MASTERS = 1
const MIN_WORKERS = 3

type UtilsClient struct {
	Logger KubernetesLogger

	clientSet *kubernetes.Clientset
	helper    *drain.Helper
}

type NodeMap map[string]corev1.Node
