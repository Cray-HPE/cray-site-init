package etcd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"net"
	"net/http"
	"time"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewETCDClient - Creates a new etcd client.
func NewETCDClient(endpoints []string, kubeconfig string) (utilsClient *UtilsClient, err error) {
	// Use Kubernetes for all certificate related activities.
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

	// Get the certs to connect to the cluster.
	caCertData, tlsCertData, tlsKeyData, getErr := getETCDCertsData(clientSet)
	if getErr != nil {
		err = fmt.Errorf("failed to get ETCD certs data: %w", getErr)
		return
	}

	var tlsConfig *tls.Config

	if caCertData != nil && tlsCertData != nil && tlsKeyData != nil {
		cert, loadErr := tls.X509KeyPair(tlsCertData, tlsKeyData)
		if loadErr != nil {
			return
		}

		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(caCertData)

		tlsConfig = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            pool,
			InsecureSkipVerify: true,
		}
	}

	client, clientErr := clientv3.New(clientv3.Config{
		Endpoints:        endpoints,
		DialTimeout:      5 * time.Second,
		TLS:              tlsConfig,
		AutoSyncInterval: 100 * time.Millisecond,
	})
	if clientErr != nil {
		err = fmt.Errorf("failed to create new client: %w", clientErr)
		return
	}

	// Setup the client now so we know we can just use it in whatever function without thinking about it.
	transport := http.Transport{
		TLSClientConfig: tlsConfig,
	}

	utilsClient = &UtilsClient{
		client:  client,
		cluster: clientv3.NewCluster(client),
		httpClient: http.Client{
			Transport: &transport,
			Timeout:   time.Second * 5,
		},
		clientSet: clientSet,
		Endpoints: endpoints,
	}

	return
}

// CloseETCDClient - Closes the current etcd client.
func (utilsClient *UtilsClient) CloseETCDClient() error {
	return utilsClient.client.Close()
}

// IsMember - Returns true if the given NCN (of the format ncn-mXXX) is a member of the etcd cluster.
func (utilsClient *UtilsClient) IsMember(ncn string) (bool, error) {
	members, err := utilsClient.getMembers()
	if err != nil {
		return false, err
	}

	for _, member := range members {
		if member.Name == ncn {
			return true, nil
		}
	}

	return false, nil
}

func (utilsClient *UtilsClient) getMembers() (members []*etcdserverpb.Member, err error) {
	memberListResponse, listErr := utilsClient.cluster.MemberList(context.Background())
	if err != nil {
		err = fmt.Errorf("failed to get etcd member list: %w", listErr)
		return
	}

	members = memberListResponse.Members

	return
}

// RemoveMember - Removes the given NCN from the etcd cluster.
func (utilsClient *UtilsClient) RemoveMember(ncn string) (bool, error) {
	members, err := utilsClient.getMembers()
	if err != nil {
		return false, fmt.Errorf("failed to remove member: %w", err)
	}
	var targetMember *etcdserverpb.Member
	for _, member := range members {
		if member.Name == ncn {
			targetMember = member
			break
		}
	}
	if targetMember == nil {
		// No op: the given NCN is not a member of etcd cluster, nothing to remove
		return true, nil
	}

	// Do not allow the removal if the cluster isn't healthy!
	if err := utilsClient.ClusterIsHealthy(); err != nil {
		return false, fmt.Errorf("cluster is not healthy, can not remove member: %w", err)
	}

	// Now we can proceed with the removal.
	_, err = utilsClient.cluster.MemberRemove(context.Background(), targetMember.ID)
	if err != nil {
		return false, fmt.Errorf("failed to remove member: %w", err)
	}

	return true, nil
}

// AddMember - Adds an NCN to the etcd cluster.
func (utilsClient *UtilsClient) AddMember(ncn string) (bool, error) {
	// Sanity check, make sure it's not already a member.
	isMember, err := utilsClient.IsMember(ncn)
	if err != nil {
		return false, fmt.Errorf("failed to add member because could not verify current membership status: %w",
			err)
	}
	if isMember {
		return false, fmt.Errorf("failed to add member becuase it is already a member")
	}

	// The only thing we need to add a new etcd member are the peer URLs which is just a single entry in a string
	// array of the format `https://IP_ADDRESS:2380`.
	// To reduce complexity we'll get the IP address directly from DNS.
	ncnIPs, err := net.LookupIP(fmt.Sprintf("%s.nmn", ncn))
	if err != nil {
		return false, fmt.Errorf("failed to lookup IP for NCN: %w", err)
	}
	if len(ncnIPs) > 1 {
		return false, fmt.Errorf("found more than one IP for NCN in DNS lookup")
	}

	nmnIP := ncnIPs[0]
	peerURLs := []string{fmt.Sprintf("https://%s:2380", nmnIP)}

	// Now we can add this member.
	_, err = utilsClient.cluster.MemberAdd(context.Background(), peerURLs)
	if err != nil {
		return false, fmt.Errorf("failed to add member: %w", err)
	}

	return true, nil
}

// ClusterIsHealthy - Returns true if the etcd cluster is healthy.
func (utilsClient *UtilsClient) ClusterIsHealthy() error {
	// Unfortunately there isn't anything built into the Go package to do this for us, so we have to hit each endpoint
	// and figure it out for ourselves.
	for _, endpoint := range utilsClient.Endpoints {
		response, err := utilsClient.httpClient.Get(fmt.Sprintf("https://%s/health", endpoint))
		if err != nil {
			return fmt.Errorf("failed to check health of %s: %w", endpoint, err)
		}

		var health Health
		err = json.NewDecoder(response.Body).Decode(&health)
		if err != nil {
			return fmt.Errorf("failed to decode response from %s: %w", endpoint, err)
		}

		// Why...WHY do they not just use the JSON bool type!?
		if health.Health != "true" {
			return fmt.Errorf("endpoint %s is not healthy", endpoint)
		}
	}

	return nil
}

func getETCDCertsData(clientSet *kubernetes.Clientset) (caCertData []byte, tlsCertData []byte, tlsKeyData []byte,
	err error) {
	secret, getErr := clientSet.CoreV1().Secrets("kube-system").Get(context.Background(),
		EtcdSecretName, v1.GetOptions{})
	if getErr != nil {
		err = fmt.Errorf("failed to get secret %s: %w", secret, getErr)
		return
	}

	var ok bool

	caCertData, ok = secret.Data["ca.crt"]
	if !ok {
		err = fmt.Errorf("failed to get CA cert")
	}

	tlsCertData, ok = secret.Data["tls.crt"]
	if !ok {
		err = fmt.Errorf("failed to get TLS cert")
	}

	tlsKeyData, ok = secret.Data["tls.key"]
	if !ok {
		err = fmt.Errorf("failed to get TLS key")
	}

	return
}
