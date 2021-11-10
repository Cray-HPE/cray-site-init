package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"math/rand"
	"net/http"
	"time"
)

func NewPostgresClient(kubeconfig string) (utilsClient *UtilsClient, err error) {
	// Build config from kubeconfig file.
	config, configErr := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if configErr != nil {
		err = fmt.Errorf("failed to build config from kubeconfig file: %w", configErr)
		return
	}

	// Create the clientset.
	clientSet, clientsetErr := kubernetes.NewForConfig(config)
	if clientsetErr != nil {
		err = fmt.Errorf("failed to create clientset: %w", clientsetErr)
		return
	}

	utilsClient = &UtilsClient{
		clientSet: clientSet,
		httpClient: http.Client{
			Timeout: time.Second * 5,
		},
	}

	return
}

func (utilsClient *UtilsClient) GetAllClusters() (clusters map[string]Cluster, err error) {
	clusters = make(map[string]Cluster)

	pods, listErr := utilsClient.clientSet.CoreV1().Pods("").List(context.Background(),
		v1.ListOptions{
			LabelSelector: "application=spilo",
		})
	if listErr != nil {
		err = fmt.Errorf("failed to list spilo pods: %w", listErr)
		return
	}

	// This is super wasteful, but, I didn't really feel like pulling in all the CRD structures for Zolando just to do
	// a health check.
	for _, pod := range pods.Items {
		clusterName, labelExists := pod.Labels["cluster-name"]
		if !labelExists {
			err = fmt.Errorf("pod does not contain expected cluster-name label: %s", pod.Name)
			return
		}

		endpoint := fmt.Sprintf("http://%s:8008/cluster", pod.Status.PodIP)
		response, getErr := utilsClient.httpClient.Get(endpoint)
		if getErr != nil {
			err = fmt.Errorf("failed to get cluster status from %s: %w", pod.Name, getErr)
			return
		}

		var cluster Cluster
		decodeErr := json.NewDecoder(response.Body).Decode(&cluster)
		if decodeErr != nil {
			err = fmt.Errorf("failed to decode response from %s: %w", endpoint, decodeErr)
		}

		clusters[clusterName] = cluster
	}

	return
}

func HealthCheckCluster(cluster Cluster) error {
	timeline := cluster.Members[0].Timeline
	for _, member := range cluster.Members {
		// Make sure all members are on the same timeline.
		if member.Timeline != timeline {
			return fmt.Errorf("timeline mismatch: %s (%d vs. %d)", member.Name, member.Timeline, timeline)
		}

		// Check for bad lag. But first because this can be an int or a string figure that out.
		switch v := member.Lag.(type) {
		case string:
			return fmt.Errorf("unacceptable lag: %s (currently: %s)", member.Name, v)
		case int:
			if v > MAX_LAG {
				return fmt.Errorf("unacceptable lag: %s (currently %d)", member.Name, v)
			}
		}

		// Check for running state.
		if member.State != "running" {
			return fmt.Errorf("instance not running: %s (currently: %s)", member.Name, member.State)
		}
	}

	return nil
}

func (utilsClient *UtilsClient) HealthCheckAllClusters() error {
	var errors []error

	clusters, err := utilsClient.GetAllClusters()
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to get Postgres clusters: %w", err))
	}

	for clusterName, cluster := range clusters {
		err = HealthCheckCluster(cluster)
		if err != nil {
			errors = append(errors, fmt.Errorf("%s is not healthy: %w", clusterName, err))
		}
	}

	if len(errors) == 0 {
		// All clusters must be healthy if we get to this point.
		for clusterName, _ := range clusters {
			utilsClient.Logger.Debugf("%s is healthy.", clusterName)
		}
	}

	return utilerrors.NewAggregate(errors)
}

func (utilsClient *UtilsClient) FailoverPostgresLeaders(ncn string) error {
	// Find all the Postgres pods running on the given NCN.
	postgresPods, err := utilsClient.clientSet.CoreV1().Pods("").List(context.Background(),
		v1.ListOptions{
			LabelSelector: "application=spilo",
			FieldSelector: fmt.Sprintf("spec.nodeName=%s", ncn),
		})
	if err != nil {
		return fmt.Errorf("failed to get Postgres pods on NCN: %w", err)
	}

	clusters, err := utilsClient.GetAllClusters()
	if err != nil {
		return fmt.Errorf("failed to get Postgres clusters: %w", err)
	}

	for clusterName, cluster := range clusters {
		err = HealthCheckCluster(cluster)
		if err != nil {
			return fmt.Errorf("%s is not healthy: %w", clusterName, err)
		}

		// Find the current leader.
		var currentLeader Member
		var candidates []string
		for _, member := range cluster.Members {
			if member.Role == "leader" {
				currentLeader = member
			} else {
				candidates = append(candidates, member.Name)
			}
		}

		// Is the leader running on the target NCN?
		isRunningLeader := false
		for _, pod := range postgresPods.Items {
			if pod.Name == currentLeader.Name {
				isRunningLeader = true
				break
			}
		}

		// Do we need to failover?
		if !isRunningLeader {
			utilsClient.Logger.Debugf("Not failing over %s because leader pod (%s) is not on %s.",
				clusterName, currentLeader.Name, ncn)
			continue
		}

		// We're healthy, move it, move it! Just to be fun, pick a random standby candidate. Democracy!
		randomIndex := rand.Intn(len(candidates))
		pick := candidates[randomIndex]

		body := struct {
			Candidate string `json:"candidate"`
		}{Candidate: pick}

		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON body: %w", err)
		}

		response, err := utilsClient.httpClient.Post(fmt.Sprintf("http://%s:8008/failover", currentLeader.Host),
			"application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			return fmt.Errorf("failed to failover cluster %s from %s to %s",
				clusterName, currentLeader.Name, pick)
		} else if response.StatusCode != http.StatusOK {
			return fmt.Errorf("failover cluster %s from %s to %s gave unexpected status code: %d",
				clusterName, currentLeader.Name, pick, response.StatusCode)
		} else {
			utilsClient.Logger.Debugf("%s failed over from %s to %s.",
				clusterName, currentLeader.Name, pick)
		}
	}

	return nil
}
