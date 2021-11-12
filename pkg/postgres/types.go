package postgres

import (
	"k8s.io/client-go/kubernetes"
	"net/http"
)

// MaxLag - The maximum lag in MB to tolerate when determining health of a Postgres cluster.
const MaxLag = 100

// UtilsClient - Structure for Postgres client.
type UtilsClient struct {
	Logger postgresLogger

	clientSet *kubernetes.Clientset

	httpClient http.Client
}

// Structures for Patroni responses.

// PatroniMember - Structure that allows for unmarshalling the response from Patroni.
type PatroniMember struct {
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

// PatroniCluster - Structure that allows for unmarshalling the response from Patroni.
type PatroniCluster struct {
	Members []PatroniMember `json:"members"`
}
