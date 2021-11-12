package etcd

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"k8s.io/client-go/kubernetes"
	"net/http"
)

// EtcdSecretName - Name of the secret containing the etcd keys.
const EtcdSecretName = "kube-etcdbackup-etcd"

// UtilsClient - Structure for etcd client.
type UtilsClient struct {
	client     *clientv3.Client
	cluster    clientv3.Cluster
	httpClient http.Client

	clientSet *kubernetes.Clientset
	Endpoints []string
}

// Health - Structure to support unmarshalling health payload from etcd.
type Health struct {
	Health string `json:"health"`
}
