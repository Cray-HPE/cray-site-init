//
//  MIT License
//
//  (C) Copyright 2022 Hewlett Packard Enterprise Development LP
//
//  Permission is hereby granted, free of charge, to any person obtaining a
//  copy of this software and associated documentation files (the "Software"),
//  to deal in the Software without restriction, including without limitation
//  the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the
//  Software is furnished to do so, subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included
//  in all copies or substantial portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
//  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
//  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
//  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
//  OTHER DEALINGS IN THE SOFTWARE.

package kubernetes

import (
	"context"
	"fmt"
	"log"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/drain"
)

// NewKubernetesClient - Creates a new kubernetes client.
func NewKubernetesClient(kubeconfig string) (utilsClient *UtilsClient, err error) {
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
		helper: &drain.Helper{
			Ctx:                 context.Background(),
			Client:              clientSet,
			Force:               false,
			IgnoreAllDaemonSets: true,
			DeleteEmptyDirData:  true,
			DisableEviction:     false,
			GracePeriodSeconds:  30,
			Out:                 log.Default().Writer(),
			ErrOut:              log.Default().Writer(),
		},
	}

	// Log progress when draining.
	utilsClient.helper.OnPodDeletedOrEvicted = func(pod *corev1.Pod, usingEviction bool) {
		utilsClient.Logger.debugf("Pod %s (in %s namespace) deleted (using eviction: %t) while draining %s.",
			pod.Name, pod.Namespace, usingEviction, pod.Spec.NodeName)
	}

	return
}

// ChangeNCNCordonState - Cordons or uncordons the given NCN.
func (utilsClient *UtilsClient) ChangeNCNCordonState(ncn string, cordoned bool) error {
	// Get the node.
	node, err := utilsClient.clientSet.CoreV1().Nodes().Get(context.Background(), ncn, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to (un)cordon %s: %w", ncn, err)
	}

	err = drain.RunCordonOrUncordon(utilsClient.helper, node, cordoned)
	if err != nil {
		return fmt.Errorf("failed to (un)cordon %s: %w", ncn, err)
	}

	return nil
}

// CordonNCN - Cordons the given NCN.
func (utilsClient *UtilsClient) CordonNCN(ncn string) error {
	return utilsClient.ChangeNCNCordonState(ncn, true)
}

// UnCordonNCN - Uncordons the given NCN.
func (utilsClient *UtilsClient) UnCordonNCN(ncn string) error {
	return utilsClient.ChangeNCNCordonState(ncn, false)
}

// DrainNCN - Draining an NCN is really 3 individual steps:
//   1) Cordon the node.
//   2) Identify any pods that might be on that node in violation of any pod distribution budgets and move them.
//   3) Drain the node.
func (utilsClient *UtilsClient) DrainNCN(ncn string) error {
	err := utilsClient.ChangeNCNCordonState(ncn, true)
	if err != nil {
		return fmt.Errorf("failed to drain NCN: %w", err)
	}

	utilsClient.Logger.debugf("%s cordoned.", ncn)

	// Now identify any pods running on this NCN that also have a pod distribution budget.
	budgets, err := utilsClient.clientSet.PolicyV1beta1().PodDisruptionBudgets("").List(context.Background(),
		v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to drain NCN: %w", err)
	}

	for _, budget := range budgets.Items {
		labels := budget.Spec.Selector.MatchLabels

		for key, value := range labels {
			// Find all the pods on this node with this budget's labels.
			pods, err := utilsClient.clientSet.CoreV1().Pods("").List(context.Background(),
				v1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", key, value)})
			if err != nil {
				return fmt.Errorf("failed to drain NCN: %w", err)
			}

			for _, pod := range pods.Items {
				if pod.Spec.NodeName == ncn {
					err := utilsClient.clientSet.CoreV1().Pods(pod.Namespace).Delete(context.Background(), pod.Name,
						v1.DeleteOptions{})
					if err != nil {
						return fmt.Errorf("failed to drain NCN: %w", err)
					}

					utilsClient.Logger.debugf("Deleted pod %s in %s namespace to satisfy pod disruption policy.",
						pod.Name, pod.Namespace)
				}
			}
		}
	}

	utilsClient.Logger.debugf("Pod disruption budgets satisfied.")

	// Now finally we can drain the node.
	attempts := 3
	sleep, _ := time.ParseDuration("2s")
	for i := 0; i < attempts; i++ {
		if i > 0 {
			utilsClient.Logger.warningf("Retrying after error: %w", err)
			time.Sleep(sleep)
		}
		err := utilsClient.runNodeDrain(utilsClient.helper, ncn)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("failed to drain NCN: %w", err)

}

// runNodeDrain is basically a direct copy of a function by the same name in the drain package except the way that it
// logs warnings. To stay consistent with the rest of the package it's reimplemented and pointed to the logger for the
// client.
func (utilsClient *UtilsClient) runNodeDrain(drainer *drain.Helper, nodeName string) error {
	list, errs := drainer.GetPodsForDeletion(nodeName)
	if errs != nil {
		return utilerrors.NewAggregate(errs)
	}
	if warnings := list.Warnings(); warnings != "" {
		utilsClient.Logger.warningf("%s", warnings)
	}

	if err := drainer.DeleteOrEvictPods(list.Pods()); err != nil {
		// Maybe warn about non-deleted pods here
		return err
	}

	return nil
}

// DeleteNCN - Deletes the NCN from Kubernetes.
func (utilsClient *UtilsClient) DeleteNCN(ncn string) error {
	// Start by making sure the NCN is drained.
	drainErr := utilsClient.DrainNCN(ncn)
	if drainErr != nil {
		return fmt.Errorf("failed to delete NCN: %w", drainErr)
	}

	// With the NCN fully drained of all pods now it can be removed from Kubernetes.
	deleteErr := utilsClient.clientSet.CoreV1().Nodes().Delete(context.Background(), ncn, v1.DeleteOptions{})
	if deleteErr != nil {
		return fmt.Errorf("failed to delete NCN: %w", deleteErr)
	}

	return nil
}

// IsMember - Returns true if the given NCN is a member of Kubernetes.
func (utilsClient *UtilsClient) IsMember(ncn string) (bool, error) {
	node, err := utilsClient.clientSet.CoreV1().Nodes().Get(context.Background(), ncn, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}

		return false, fmt.Errorf("failed to get node: %w", err)
	}

	// Make sure the node is ready.
	ready := false
	for _, condition := range node.Status.Conditions {
		if condition.Type == "Ready" && condition.Status == "True" {
			ready = true
			break
		}
	}

	return ready, nil
}

// GetNodes - Returns a map of all the nodes.
func (utilsClient *UtilsClient) GetNodes() (nodeMap NodeMap, err error) {
	nodes, listErr := utilsClient.clientSet.CoreV1().Nodes().List(context.Background(), v1.ListOptions{})
	if listErr != nil {
		err = fmt.Errorf("failed to list nodes: %w", listErr)
		return
	}

	nodeMap = make(NodeMap)

	for _, node := range nodes.Items {
		nodeMap[node.Name] = node
	}

	return
}

// IsMaster - Returns true if the given node is a master.
func IsMaster(node corev1.Node) bool {
	for key := range node.Labels {
		if key == "node-role.kubernetes.io/master" {
			return true
		}
	}

	return false
}
