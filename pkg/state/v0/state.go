package v0

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"

	spv1 "github.com/platform9/ssh-provider/pkg/apis/sshprovider/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	spclient "github.com/platform9/ssh-provider/pkg/client/clientset_generated/clientset"
	"k8s.io/client-go/kubernetes"
	clusterclient "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
)

const (
	// FileMode defines the file mode used to create the state file.
	FileMode = 0600
)

// State holds all the objects that make up cctl state.State Contains unexported
// fields.
type State struct {
	Filename      string                  `json:"-"`
	KubeClient    kubernetes.Interface    `json:"-"`
	ClusterClient clusterclient.Interface `json:"-"`
	SPClient      spclient.Interface      `json:"-"`

	SecretList             corev1.SecretList           `json:"secretList,omitempty"`
	ClusterList            clusterv1.ClusterList       `json:"clusterList,omitempty"`
	MachineList            clusterv1.MachineList       `json:"machineList,omitempty"`
	ProvisionedMachineList spv1.ProvisionedMachineList `json:"provisionedMachineList,omitempty"`
}

// NewWithFile returns the state ready to sync objects between the APIs and the
// file.
func NewWithFile(filename string, kubeClient kubernetes.Interface, clusterClient clusterclient.Interface, spClient spclient.Interface) *State {
	s := State{
		Filename:      filename,
		KubeClient:    kubeClient,
		ClusterClient: clusterClient,
		SPClient:      spClient,

		SecretList:             corev1.SecretList{},
		ClusterList:            clusterv1.ClusterList{},
		MachineList:            clusterv1.MachineList{},
		ProvisionedMachineList: spv1.ProvisionedMachineList{},
	}
	return &s
}

func (s *State) read() error {
	file, err := os.OpenFile(s.Filename, os.O_RDONLY|os.O_CREATE, FileMode)
	if err != nil {
		return fmt.Errorf("unable to open %q: %v", s.Filename, err)
	}
	defer file.Close()
	stateBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("unable to read from %q: %v", s.Filename, err)
	}
	if err := yaml.Unmarshal(stateBytes, s); err != nil {
		return fmt.Errorf("unable to unmarshal state from YAML: %v", err)
	}
	return nil
}

func (s *State) write() error {
	file, err := os.OpenFile(s.Filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, FileMode)
	if err != nil {
		return fmt.Errorf("unable to open %q: %v", s.Filename, err)
	}
	defer file.Close()
	stateBytes, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("unable to marshal state to YAML: %v", err)
	}
	_, err = file.Write(stateBytes)
	if err != nil {
		return fmt.Errorf("unable to write to %q: %v", s.Filename, err)
	}
	return nil
}

// PullFromAPIs reads objects in the state file and creates them using the APIs.
// If the file does not exist, it will be created.
func (s *State) PushToAPIs() error {
	if err := s.read(); err != nil {
		return err
	}
	for _, secret := range s.SecretList.Items {
		if _, err := s.KubeClient.CoreV1().Secrets(secret.Namespace).Create(&secret); err != nil {
			return err
		}
	}
	for _, cluster := range s.ClusterList.Items {
		if _, err := s.ClusterClient.ClusterV1alpha1().Clusters(cluster.Namespace).Create(&cluster); err != nil {
			return err
		}
	}
	for _, machine := range s.MachineList.Items {
		if _, err := s.ClusterClient.ClusterV1alpha1().Machines(machine.Namespace).Create(&machine); err != nil {
			return err
		}
	}
	for _, pm := range s.ProvisionedMachineList.Items {
		if _, err := s.SPClient.SshproviderV1alpha1().ProvisionedMachines(pm.Namespace).Create(&pm); err != nil {
			return err
		}
	}
	return nil
}

// PullFromAPIs stores API objects in the state file. If the file does not
// exist, it will be created.
func (s *State) PullFromAPIs() error {
	secretList, err := s.KubeClient.CoreV1().Secrets(corev1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	s.SecretList = *secretList
	clusterList, err := s.ClusterClient.ClusterV1alpha1().Clusters(corev1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	s.ClusterList = *clusterList
	machineList, err := s.ClusterClient.ClusterV1alpha1().Machines(corev1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	s.MachineList = *machineList
	pmList, err := s.SPClient.SshproviderV1alpha1().ProvisionedMachines(corev1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	s.ProvisionedMachineList = *pmList
	return s.write()
}
