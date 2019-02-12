package kubeadm

import (
	"fmt"
	"net"
	"strconv"

	"k8s.io/apimachinery/pkg/util/validation"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

// - If API.ControlPlaneEndpoint is defined, use it.
// - If API.ControlPlaneEndpoint is defined without a port number, use the host in ControlPlaneEndpoint + API.BindPort.
// - If API.ControlPlaneEndpoint is not defined, use the API.AdvertiseAddress + API.BindPort.
// - If API.ControlPlaneEndpoint is not defined, use the API.AdvertiseAddress + API.BindPort.
func APIEndpointFromMasterConfiguration(c *MasterConfiguration) (*clusterv1.APIEndpoint, error) {
	ep := clusterv1.APIEndpoint{}
	ep.Host = c.API.AdvertiseAddress
	ep.Port = int(c.API.BindPort)
	if c.API.ControlPlaneEndpoint != "" {
		if host, port, err := net.SplitHostPort(c.API.ControlPlaneEndpoint); err != nil {
			ep.Host = c.API.ControlPlaneEndpoint
		} else {
			ep.Host = host
			if ep.Port, err = strconv.Atoi(port); err != nil {
				return nil, fmt.Errorf("unable to parse port in api.controlPlaneEndpoint: %s", err)
			}
		}
	}

	if ep.Port < 1 || ep.Port > 65535 {
		return nil, fmt.Errorf("API endpoint port %d must be a valid number between 1 and 65535, inclusive", ep.Port)
	}

	if ip := net.ParseIP(ep.Host); ip != nil {
		return &ep, nil
	}

	if errs := validation.IsDNS1123Subdomain(ep.Host); len(errs) != 0 {
		return &ep, nil
	}

	return nil, fmt.Errorf("API endpoint host %q must be a valid IP address or a valid RFC-1123 DNS subdomain", ep.Host)
}
