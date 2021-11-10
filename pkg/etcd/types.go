package etcd

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"k8s.io/client-go/kubernetes"
	"net/http"
)

const ETCD_SECRET_NAME = "kube-etcdbackup-etcd"

type UtilsClient struct {
	client     *clientv3.Client
	cluster    clientv3.Cluster
	httpClient http.Client

	clientSet *kubernetes.Clientset
	Endpoints []string
}

type Health struct {
	Health string `json:"health"`
}
