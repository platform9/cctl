package machine

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/ghodss/yaml"
	"github.com/platform9/ssh-provider/provisionedmachine"
	kubeadmv1 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1alpha1"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

type NodeadmConfiguration struct {
	VIPConfiguration    VIPConfiguration              `json:"vipConfiguration,omitEmpty"`
	MasterConfiguration kubeadmv1.MasterConfiguration `json:"masterConfiguration"`
}

type VIPConfiguration struct {
	// The virtual IP.
	IP net.IP `json:"ip"`
	// The virtual router ID. Must be in the range [0, 254]. Must be unique within
	// a single L2 network domain.
	RouterID int `json:"routerID"`
	// Network interface chosen to create the virtual IP. If it is not specified,
	// the interface of the default gateway is chosen.
	NetworkInterface string `json:"networkInterface,omitEmpty"`
}

func (sa *SSHActuator) NewNodeadmConfiguration(pm *provisionedmachine.ProvisionedMachine, cluster *clusterv1.Cluster, machine *clusterv1.Machine) (*NodeadmConfiguration, error) {
	masterConfiguration, err := sa.NewMasterConfiguration(cluster, machine)
	if err != nil {
		return nil, fmt.Errorf("error creating nodeadm configuration: %s", err)
	}

	cpc, err := sa.clusterproviderconfig(cluster.Spec.ProviderConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating nodeadm configuration: %s", err)
	}

	cfg := &NodeadmConfiguration{}
	cfg.VIPConfiguration.IP = cpc.VIPConfiguration.IP
	cfg.VIPConfiguration.RouterID = cpc.VIPConfiguration.RouterID
	cfg.VIPConfiguration.NetworkInterface = pm.VIPNetworkInterface
	cfg.MasterConfiguration = *masterConfiguration.DeepCopy()
	return cfg, nil
}

func MarshalToYAMLWithFixedKubeProxyFeatureGates(nodeadmConfiguration *NodeadmConfiguration) ([]byte, error) {
	j, err := json.Marshal(nodeadmConfiguration)
	if err != nil {
		return nil, fmt.Errorf("error marshalling nodeadm configuration: %s", err)
	}
	p, err := gabs.ParseJSON(j)
	fgString, ok := p.Path("masterConfiguration.kubeProxy.config.featureGates").Data().(string)
	if !ok {
		return nil, fmt.Errorf("error marshalling nodeadm configuration: error parsing masterConfiguration.kubeProxy.config.featureGates: %s", err)
	}
	p.ObjectP("masterConfiguration.kubeProxy.config.featureGates")
	if strings.Contains(fgString, ",") {
		for _, gate := range strings.Split(fgString, ",") {
			p.SetP(true, fmt.Sprintf("masterConfiguration.kubeProxy.config.featureGates.%s", gate))
		}
	}
	y, err := yaml.JSONToYAML(p.Bytes())
	if err != nil {
		return nil, fmt.Errorf("error marshalling nodeadm configuration: %s", err)
	}
	return y, nil
}
