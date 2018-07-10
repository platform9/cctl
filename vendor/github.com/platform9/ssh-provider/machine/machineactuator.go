/*
Copyright 2018 Platform 9 Systems, Inc.
*/

package machine

import (
	"encoding/json"
	"fmt"
	"log"

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

	provisionedMachineConfigMaps []*corev1.ConfigMap
	sshCredentials               *corev1.Secret
	etcdCA                       *corev1.Secret
	apiServerCA                  *corev1.Secret
	frontProxyCA                 *corev1.Secret
	serviceAccountKey            *corev1.Secret
	clusterToken                 *corev1.Secret
}

func NewActuator(provisionedMachineConfigMaps []*corev1.ConfigMap,
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
		sshProviderCodec:             codec,
		provisionedMachineConfigMaps: provisionedMachineConfigMaps,
		sshCredentials:               sshCredentials,
		etcdCA:                       etcdCA,
		apiServerCA:                  apiServerCA,
		frontProxyCA:                 frontProxyCA,
		serviceAccountKey:            serviceAccountKey,
		clusterToken:                 clusterToken,
	}, nil
}

func (sa *SSHActuator) Create(cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	cm, err := sa.ReserveProvisionedMachine(machine)
	if err != nil {
		return fmt.Errorf("error creating machine: error reserving provisioned machine %q: %s", machine.Name, err)
	}

	client, err := sshClient(cm, sa.sshCredentials, sa.InsecureIgnoreHostKey)
	if err != nil {
		return fmt.Errorf("error creating machine %q: failed to create SSH client: %s", machine.Name, err)
	}
	defer client.Close()

	pm, err := provisionedmachine.NewFromConfigMap(cm)
	if err != nil {
		return fmt.Errorf("error creating machine: error parsing ProvisionedMachine from ConfigMap %q: %s", cm.Name, err)
	}
	if clusterutil.IsMaster(machine) {
		if err := sa.createMaster(pm, cluster, machine, client); err != nil {
			return fmt.Errorf("error creating machine %q: %s", machine.Name, err)
		}
	} else {
		if err := sa.createNode(cluster, machine, client); err != nil {
			return fmt.Errorf("error creating machine %q: %s", machine.Name, err)
		}
	}
	return nil
}

func chooseEtcdEndpoint(members []sshconfigv1.EtcdMember) string {
	return members[0].ClientURLs[0]
}

func (sa *SSHActuator) createMaster(pm *provisionedmachine.ProvisionedMachine, cluster *clusterv1.Cluster, machine *clusterv1.Machine, client *ssh.Client) error {
	var err error

	nodeadmConfiguration, err := sa.NewNodeadmConfiguration(pm, cluster, machine)
	if err != nil {
		return err
	}

	mcb, err := MarshalToYAMLWithFixedKubeProxyFeatureGates(nodeadmConfiguration)
	if err != nil {
		return err
	}

	sftp, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("error creating SFTP client: %s", err)
	}
	defer sftp.Close()

	f, err := sftp.Create("/tmp/nodeadm.yaml")
	if err != nil {
		return fmt.Errorf("error creating kubeadm.yaml: %s", err)
	}
	if _, err := f.Write(mcb); err != nil {
		return fmt.Errorf("error writing kubeadm.yaml: %s", err)
	}
	sa.writeCAs(sftp)

	var session *ssh.Session
	var out []byte
	session, err = client.NewSession()
	defer session.Close()
	if err != nil {
		return fmt.Errorf("error creating new SSH session for machine %q: %s", machine.Name, err)
	}
	out, err = session.CombinedOutput("echo writing ca cert and key")
	if err != nil {
		return fmt.Errorf("error invoking ssh command %s", err)
	}
	log.Println(string(out))

	session, err = client.NewSession()
	defer session.Close()
	if err != nil {
		return fmt.Errorf("error creating new SSH session for machine %q: %s", machine.Name, err)
	}
	clusterProviderStatus := sshconfigv1.SSHClusterProviderStatus{}
	if cluster.Status.ProviderStatus.Value != nil {
		err = sa.sshProviderCodec.DecodeFromProviderStatus(cluster.Status.ProviderStatus, &clusterProviderStatus)
	}
	if err != nil {
		return fmt.Errorf("error decoding cluster provider status %v\n", err)
	}
	if len(clusterProviderStatus.EtcdMembers) > 0 {
		member := chooseEtcdEndpoint(clusterProviderStatus.EtcdMembers)
		out, err = session.CombinedOutput(fmt.Sprintf("/opt/bin/etcdadm join %s", member))
	} else {
		out, err = session.CombinedOutput("/opt/bin/etcdadm init")
	}
	if err != nil {
		return fmt.Errorf("error invoking ssh command %s", err)
	}
	log.Println(string(out))

	session, err = client.NewSession()
	defer session.Close()
	if err != nil {
		return fmt.Errorf("error creating new SSH session for machine %q: %s", machine.Name, err)
	}
	out, err = session.CombinedOutput("/opt/bin/etcdadm info")
	if err != nil {
		return fmt.Errorf("error invoking ssh command %s", err)
	}
	etcdMember := sshconfigv1.EtcdMember{}
	err = json.Unmarshal(out, &etcdMember)
	if err != nil {
		return fmt.Errorf("error reading etcdadm info: %s", err)
	}
	mps := &sshconfigv1.SSHMachineProviderStatus{
		EtcdMember: &etcdMember,
	}
	sa.sshProviderCodec.DecodeFromProviderStatus(machine.Status.ProviderStatus, mps)
	ps, err := sa.sshProviderCodec.EncodeToProviderStatus(mps)
	if err != nil {
		return fmt.Errorf("error encoding machine provider status: %s", err)
	}
	machine.Status.ProviderStatus = *ps

	session, err = client.NewSession()
	defer session.Close()
	if err != nil {
		return fmt.Errorf("error creating new SSH session for machine %q: %s", machine.Name, err)
	}
	out, err = session.CombinedOutput("/opt/bin/nodeadm init --cfg /tmp/nodeadm.yaml")
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

func (sa *SSHActuator) createNode(cluster *clusterv1.Cluster, machine *clusterv1.Machine, client *ssh.Client) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("error creating new SSH session for machine %q: %s", machine.Name, err)
	}
	cmd := fmt.Sprintf("/opt/bin/nodeadm join --master %s --token %s --cahash %s",
		getAPIEndPoint(cluster),
		string(sa.clusterToken.Data["token"]),
		string(sa.clusterToken.Data["cahash"]))
	log.Printf("Nodeadm join command = %s", cmd)
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
