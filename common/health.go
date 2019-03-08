/*
Copyright 2019 The cctl authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"fmt"
	"strings"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"
)

func getKubeClient(kubeconfig string) (clientset.Interface, error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("Could not get build config from kubeconfig")
	}
	// create the clientset
	client, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Could not create client from kubeconfig: %v", err)
	}
	return client, nil
}

// MasterNodesReady checks whether all master Nodes in the cluster are in the Ready state
func MasterNodesReady(kubeconfig string) error {
	client, err := getKubeClient(kubeconfig)
	if err != nil {
		return fmt.Errorf("unable to create kubeclient: %v", err)
	}
	selector := labels.SelectorFromSet(labels.Set(map[string]string{
		constants.LabelNodeRoleMaster: "",
	}))
	masters, err := client.CoreV1().Nodes().List(metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return fmt.Errorf("couldn't list masters in cluster: %v", err)
	}

	if len(masters.Items) == 0 {
		return fmt.Errorf("failed to find any nodes with master role")
	}

	notReadyMasters := getNotReadyNodes(masters.Items)
	if len(notReadyMasters) != 0 {
		return fmt.Errorf("there are NotReady masters in the cluster: %v", notReadyMasters)
	}
	return nil
}

// ControlPlaneReady checks whether all master pods in the cluster are in the Ready state
func ControlPlaneReady(kubeconfig string) error {
	client, err := getKubeClient(kubeconfig)
	if err != nil {
		return fmt.Errorf("unable to create kube client: %v", err)
	}

	pods, err := client.CoreV1().Pods(KubeSystemNamespace).List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("unable to get pods for kube-system: %v", err)
	}

	notReadyMasterPods := getNotReadyMasterPods(pods.Items)
	if len(notReadyMasterPods) != 0 {
		return fmt.Errorf("there are NotReady control plane pods in the cluster: %v", notReadyMasterPods)
	}
	return nil
}

func getNotReadyMasterPods(pods []v1.Pod) []string {
	notReadyMasterPods := []string{}
	for _, pod := range pods {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == v1.PodReady && condition.Status != v1.ConditionTrue {
				for _, prefix := range MasterComponents {
					if strings.HasPrefix(pod.ObjectMeta.Name, prefix) {
						notReadyMasterPods = append(notReadyMasterPods, pod.ObjectMeta.Name)
					}
				}
			}
		}
	}
	return notReadyMasterPods
}

func getNotReadyNodes(nodes []v1.Node) []string {
	notReadyNodes := []string{}
	for _, node := range nodes {
		for _, condition := range node.Status.Conditions {
			if condition.Type == v1.NodeReady && condition.Status != v1.ConditionTrue {
				notReadyNodes = append(notReadyNodes, node.ObjectMeta.Name)
			}
		}
	}
	return notReadyNodes
}
