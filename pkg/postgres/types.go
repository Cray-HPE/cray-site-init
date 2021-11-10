package postgres

import (
	"k8s.io/client-go/kubernetes"
	"net/http"
)

const MAX_LAG = 100

type UtilsClient struct {
	Logger PostgresLogger

	clientSet *kubernetes.Clientset

	httpClient http.Client
}

// Structures for Patroni responses.

type Member struct {
	Name     string `json:"name"`
	Role     string `json:"role"`
	State    string `json:"state"`
	APIURL   string `json:"apiurl"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Timeline int    `json:"timeline"`

	// I really do hate Zolando...this can either be an int or a string.
	Lag interface{} `json:"lag"`
}

type Cluster struct {
	Members []Member `json:"members"`
}
