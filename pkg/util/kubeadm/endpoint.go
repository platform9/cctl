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

package kubeadm

import (
	"fmt"
	"net"
	"strconv"

	"k8s.io/apimachinery/pkg/util/validation"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

// APIEndpointFromClusterConfiguration parses the API Endpoint (Host and Port)
// from the kubeadm ClusterConfiguration.
func APIEndpointFromClusterConfiguration(c *ClusterConfiguration) (*clusterv1.APIEndpoint, error) {
	ep := clusterv1.APIEndpoint{}

	if c.ControlPlaneEndpoint == "" {
		return nil, fmt.Errorf("controlPlaneEndpoint is not defined")
	}

	host, portStr, err := net.SplitHostPort(c.ControlPlaneEndpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to parse %q controlPlaneEndpoint: %s", c.ControlPlaneEndpoint, err)
	}

	if ep.Port, err = strconv.Atoi(portStr); err != nil {
		return nil, fmt.Errorf("unable to parse API endpoint port %q: %s", portStr, err)
	}
	if ep.Port < 1 || ep.Port > 65535 {
		return nil, fmt.Errorf("API endpoint port %d must be a number between 1 and 65535, inclusive", ep.Port)
	}

	ep.Host = host
	if ip := net.ParseIP(ep.Host); ip == nil {
		return nil, fmt.Errorf("API endpoint host %q must be a valid IP address", ep.Host)
	}
	if errs := validation.IsDNS1123Subdomain(ep.Host); len(errs) != 0 {
		return nil, fmt.Errorf("API endpoint host %q must be a valid RFC-1123 DNS subdomain", ep.Host)
	}

	return &ep, nil
}
