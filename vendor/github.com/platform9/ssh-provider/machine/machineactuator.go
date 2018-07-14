/*
Copyright 2018 Platform 9 Systems, Inc.
*/

package machine

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ghodss/yaml"

	"github.com/pkg/sftp"
	"github.com/platform9/ssh-provider/provisionedmachine"

	"path/filepath"

	"golang.org/x/crypto/ssh"

	sshconfigv1 "github.com/platform9/ssh-provider/sshproviderconfig/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	clusterutil "sigs.k8s.io/cluster-api/pkg/util"
)

type SSHActuator struct {
	InsecureIgnoreHostKey bool
	sshProviderCodec      *sshconfigv1.SSHProviderCodec

	provisionedMachineConfigMap *corev1.ConfigMap
	sshCredentials              *corev1.Secret
	etcdCA                      *corev1.Secret
	apiServerCA                 *corev1.Secret
	frontProxyCA                *corev1.Secret
	serviceAccountKey           *corev1.Secret
	clusterToken                *corev1.Secret
}

func NewActuator(provisionedMachineConfigMap *corev1.ConfigMap,
	sshCredentials *corev1.Secret,
	etcdCA *corev1.Secret,
	apiServerCA *corev1.Secret,
	frontProxyCA *corev1.Secret,
	serviceAccountKey *corev1.Secret,
	clusterToken *corev1.Secret) (*SSHActuator, error) {
	codec, err := sshconfigv1.NewCodec()
	if err != nil {
		return nil, err
	}
	return &SSHActuator{
		sshProviderCodec:            codec,
		provisionedMachineConfigMap: provisionedMachineConfigMap,
		sshCredentials:              sshCredentials,
		etcdCA:                      etcdCA,
		apiServerCA:                 apiServerCA,
		frontProxyCA:                frontProxyCA,
		serviceAccountKey:           serviceAccountKey,
		clusterToken:                clusterToken,
	}, nil
}

func (sa *SSHActuator) Create(cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	client, err := sshClient(sa.provisionedMachineConfigMap, sa.sshCredentials, sa.InsecureIgnoreHostKey)
	if err != nil {
		return fmt.Errorf("error creating machine %q: failed to create SSH client: %s", machine.Name, err)
	}
	defer client.Close()

	pm, err := provisionedmachine.NewFromConfigMap(sa.provisionedMachineConfigMap)
	if err != nil {
		return fmt.Errorf("error creating machine: error parsing ProvisionedMachine from ConfigMap %q: %s", sa.provisionedMachineConfigMap.Name, err)
	}
	if clusterutil.IsMaster(machine) {
		if err := sa.createMaster(cluster, machine, pm, client); err != nil {
			return fmt.Errorf("error creating machine %q: %s", machine.Name, err)
		}
	} else {
		if err := sa.createNode(cluster, machine, pm, client); err != nil {
			return fmt.Errorf("error creating machine %q: %s", machine.Name, err)
		}
	}
	return nil
}

func chooseEtcdEndpoint(members []sshconfigv1.EtcdMember) string {
	return members[0].ClientURLs[0]
}

func (sa *SSHActuator) createMaster(cluster *clusterv1.Cluster, machine *clusterv1.Machine, pm *provisionedmachine.ProvisionedMachine, client *ssh.Client) error {
	var err error

	nodeadmInitConfiguration, err := sa.NodeadmInitConfigurationForMachine(pm, cluster, machine)
	if err != nil {
		return fmt.Errorf("error creating nodeadm configuration: %v", err)
	}

	sftp, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("error creating SFTP client: %s", err)
	}
	defer sftp.Close()

	nodeadmInitConfigurationBytes, err := yaml.Marshal(nodeadmInitConfiguration)
	if err != nil {
		return fmt.Errorf("error marshalling nodeadm configuration to yaml: %v", err)
	}
	f, err := sftp.Create("/tmp/nodeadm.yaml")
	if err != nil {
		return fmt.Errorf("error creating kubeadm.yaml: %s", err)
	}
	if _, err := f.Write(nodeadmInitConfigurationBytes); err != nil {
		return fmt.Errorf("error writing kubeadm.yaml: %s", err)
	}
	sa.writeCAs(sftp)

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("error creating new SSH session for machine %q: %s", machine.Name, err)
	}
	defer session.Close()
	clusterProviderStatus := sshconfigv1.SSHClusterProviderStatus{}
	if cluster.Status.ProviderStatus.Value != nil {
		err = sa.sshProviderCodec.DecodeFromProviderStatus(cluster.Status.ProviderStatus, &clusterProviderStatus)
	}
	if err != nil {
		return fmt.Errorf("error decoding cluster provider status %v", err)
	}

	var cmd string
	var out []byte
	if len(clusterProviderStatus.EtcdMembers) > 0 {
		member := chooseEtcdEndpoint(clusterProviderStatus.EtcdMembers)
		cmd = fmt.Sprintf("/opt/bin/etcdadm join %s", member)
	} else {
		cmd = "/opt/bin/etcdadm init"
	}
	session, err = client.NewSession()
	if err != nil {
		return fmt.Errorf("error creating new SSH session for machine %q: %s", machine.Name, err)
	}
	defer session.Close()
	log.Printf("Running %q on machine %s. This may take a few minutes.", cmd, machine.Name)
	out, err = session.CombinedOutput(cmd)
	if err != nil {
		return fmt.Errorf("error invoking ssh command %s", err)
	}
	log.Println(string(out))

	session, err = client.NewSession()
	if err != nil {
		return fmt.Errorf("error creating new SSH session for machine %q: %s", machine.Name, err)
	}
	defer session.Close()
	out, err = session.CombinedOutput("/opt/bin/etcdadm info")
	if err != nil {
		return fmt.Errorf("error invoking ssh command %s", err)
	}
	etcdMember := sshconfigv1.EtcdMember{}
	err = json.Unmarshal(out, &etcdMember)
	if err != nil {
		return fmt.Errorf("error reading etcdadm info: %s", err)
	}
	mps := &sshconfigv1.SSHMachineProviderStatus{}
	sa.sshProviderCodec.DecodeFromProviderStatus(machine.Status.ProviderStatus, mps)
	if err != nil {
		return fmt.Errorf("error decoding machine provider status: %s", err)
	}
	mps.EtcdMember = &etcdMember
	ps, err := sa.sshProviderCodec.EncodeToProviderStatus(mps)
	if err != nil {
		return fmt.Errorf("error encoding machine provider status: %s", err)
	}
	machine.Status.ProviderStatus = *ps

	session, err = client.NewSession()
	if err != nil {
		return fmt.Errorf("error creating new SSH session for machine %q: %s", machine.Name, err)
	}
	defer session.Close()
	cmd = "/opt/bin/nodeadm init --cfg /tmp/nodeadm.yaml"
	log.Printf("Running %q on machine %s. This may take a few minutes.", cmd, machine.Name)
	out, err = session.CombinedOutput(cmd)
	if err != nil {
		return fmt.Errorf("error invoking ssh command %s", err)
	}
	log.Println(string(out))
	return nil
}

func writeCA(sftp *sftp.Client, data []byte, fileName string) error {
	f, err := sftp.Create(fileName)
	if err != nil {
		return fmt.Errorf("error creating file: %s, %v", fileName, err)
	}
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("error writing file: %s, %v", fileName, err)
	}
	return nil
}

func (sa *SSHActuator) writeCAs(sftp *sftp.Client) error {
	basePath := "/etc/kubernetes/pki"
	//removeDir should ideally be unnecessary, will be removed in future
	sftp.RemoveDirectory(basePath)
	err := sftp.MkdirAll(basePath)
	if err != nil {
		return fmt.Errorf("error create remote directory %s, %v", basePath, err)
	}
	err = writeCA(sftp, sa.frontProxyCA.Data["tls.key"], filepath.Join(basePath, "front-proxy-ca.key"))
	if err != nil {
		return err
	}
	err = writeCA(sftp, sa.frontProxyCA.Data["tls.crt"], filepath.Join(basePath, "front-proxy-ca.crt"))
	if err != nil {
		return err
	}
	err = writeCA(sftp, sa.apiServerCA.Data["tls.key"], filepath.Join(basePath, "ca.key"))
	if err != nil {
		return err
	}
	err = writeCA(sftp, sa.apiServerCA.Data["tls.crt"], filepath.Join(basePath, "ca.crt"))
	if err != nil {
		return err
	}
	err = writeCA(sftp, sa.serviceAccountKey.Data["publickey"], filepath.Join(basePath, "sa.pub"))
	if err != nil {
		return err
	}
	err = writeCA(sftp, sa.serviceAccountKey.Data["privatekey"], filepath.Join(basePath, "sa.key"))
	if err != nil {
		return err
	}

	basePath = "/etc/etcd/pki"
	//removeDir should ideally be unnecessary, will be removed in future
	sftp.RemoveDirectory(basePath)
	sftp.MkdirAll(basePath)
	if err != nil {
		return fmt.Errorf("error create remote directory %s, %v", basePath, err)
	}
	err = writeCA(sftp, sa.etcdCA.Data["tls.key"], filepath.Join(basePath, "ca.key"))
	if err != nil {
		return err
	}
	err = writeCA(sftp, sa.etcdCA.Data["tls.crt"], filepath.Join(basePath, "ca.crt"))
	return err
}

func (sa *SSHActuator) createNode(cluster *clusterv1.Cluster, machine *clusterv1.Machine, pm *provisionedmachine.ProvisionedMachine, client *ssh.Client) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("error creating new SSH session for machine %q: %s", machine.Name, err)
	}
	defer session.Close()

	nodeadmJoinConfiguration, err := sa.NodeadmJoinConfigurationForMachine(pm, cluster, machine)
	if err != nil {
		return fmt.Errorf("error creating nodeadm configuration: %v", err)
	}

	sftp, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("error creating SFTP client: %s", err)
	}
	defer sftp.Close()

	nodeadmJoinConfigurationBytes, err := yaml.Marshal(nodeadmJoinConfiguration)
	if err != nil {
		return fmt.Errorf("error marshalling nodeadm configuration to yaml: %v", err)
	}
	f, err := sftp.Create("/tmp/nodeadm.yaml")
	if err != nil {
		return fmt.Errorf("error creating kubeadm.yaml: %s", err)
	}
	if _, err := f.Write(nodeadmJoinConfigurationBytes); err != nil {
		return fmt.Errorf("error writing kubeadm.yaml: %s", err)
	}

	cmd := fmt.Sprintf("/opt/bin/nodeadm join --cfg %s --master %s --token %s --cahash %s",
		"/tmp/nodeadm.yaml",
		getAPIEndPoint(cluster),
		string(sa.clusterToken.Data["token"]),
		string(sa.clusterToken.Data["cahash"]))
	log.Printf("Running %q on machine %s. This may take a few minutes.", cmd, machine.Name)
	out, err := session.CombinedOutput(cmd)
	if err != nil {
		return fmt.Errorf("error invoking ssh command %s", err)
	}
	log.Println(string(out))
	return nil
}

func getAPIEndPoint(cluster *clusterv1.Cluster) string {
	return fmt.Sprintf("%s:%d", cluster.Status.APIEndpoints[0].Host, cluster.Status.APIEndpoints[0].Port)
}

func (sa *SSHActuator) Delete(cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	client, err := sshClient(sa.provisionedMachineConfigMap, sa.sshCredentials, sa.InsecureIgnoreHostKey)
	if err != nil {
		return fmt.Errorf("error deleting machine %q: failed to create SSH client: %s", machine.Name, err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("error creating new SSH session for machine %q: %s", machine.Name, err)
	}
	defer session.Close()
	cmd := "/opt/bin/nodeadm reset"
	log.Printf("Running %q on machine %s. This may take a few minutes.", cmd, machine.Name)
	out, err := session.CombinedOutput(cmd)
	if err != nil {
		return fmt.Errorf("error invoking ssh command %q: %v", cmd, err)
	}
	log.Println(string(out))

	if clusterutil.IsMaster(machine) {
		session, err := client.NewSession()
		if err != nil {
			return fmt.Errorf("error creating new SSH session for machine %q: %s", machine.Name, err)
		}
		defer session.Close()
		cmd := "/opt/bin/etcdadm reset"
		log.Printf("Running %q on machine %s. This may take a few minutes.", cmd, machine.Name)
		out, err := session.CombinedOutput(cmd)
		if err != nil {
			return fmt.Errorf("error invoking ssh command %q: %v", cmd, err)
		}
		log.Println(string(out))
	}
	return nil
}

func (sa *SSHActuator) Update(cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	return nil
}

func (sa *SSHActuator) Exists(cluster *clusterv1.Cluster, machine *clusterv1.Machine) (bool, error) {
	return false, nil
}

func (sa *SSHActuator) machineproviderconfig(providerConfig clusterv1.ProviderConfig) (*sshconfigv1.SSHMachineProviderConfig, error) {
	var config sshconfigv1.SSHMachineProviderConfig
	err := sa.sshProviderCodec.DecodeFromProviderConfig(providerConfig, &config)
	if err != nil {
		return nil, fmt.Errorf("error decoding SSHMachineProviderConfig from ProviderConfig: %s", err)
	}
	return &config, nil
}

func (sa *SSHActuator) clusterproviderconfig(providerConfig clusterv1.ProviderConfig) (*sshconfigv1.SSHClusterProviderConfig, error) {
	var config sshconfigv1.SSHClusterProviderConfig
	err := sa.sshProviderCodec.DecodeFromProviderConfig(providerConfig, &config)
	if err != nil {
		return nil, fmt.Errorf("error decoding SSHClusterProviderConfig from ProviderConfig: %s", err)
	}
	return &config, nil
}
