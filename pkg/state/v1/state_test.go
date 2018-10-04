package v1_test

import (
	"testing"

	spv1 "github.com/platform9/ssh-provider/pkg/apis/sshprovider/v1alpha1"
	spclientfake "github.com/platform9/ssh-provider/pkg/client/clientset_generated/clientset/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeclientfake "k8s.io/client-go/kubernetes/fake"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	clusterclientfake "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset/fake"

	state "github.com/platform9/cctl/pkg/state/v1"
)

const (
	testFilename  = "/tmp/state.yaml"
	testNamespace = "test"
)

func TestPullFromAPIs(t *testing.T) {
	kubeClient := kubeclientfake.NewSimpleClientset()
	clusterClient := clusterclientfake.NewSimpleClientset()
	spClient := spclientfake.NewSimpleClientset()

	s := state.NewFromFile(testFilename, kubeClient, clusterClient, spClient)

	cluster := clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "cluster.k8s.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testCluster",
			Namespace: testNamespace,
		},
		Spec:   clusterv1.ClusterSpec{},
		Status: clusterv1.ClusterStatus{},
	}
	machine := clusterv1.Machine{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Machine",
			APIVersion: "cluster.k8s.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testMachine",
			Namespace: testNamespace,
		},
		Spec:   clusterv1.MachineSpec{},
		Status: clusterv1.MachineStatus{},
	}
	secret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testSecret",
			Namespace: testNamespace,
		},
		Data: make(map[string][]byte),
	}
	pm := spv1.ProvisionedMachine{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ProvisionedMachine",
			APIVersion: "sshprovider.platform9.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testProvisionedMachine",
			Namespace: testNamespace,
		},
		Spec:   spv1.ProvisionedMachineSpec{},
		Status: spv1.ProvisionedMachineStatus{},
	}

	// Post to APIs
	kubeClient.CoreV1().Secrets(testNamespace).Create(&secret)
	clusterClient.ClusterV1alpha1().Clusters(testNamespace).Create(&cluster)
	clusterClient.ClusterV1alpha1().Machines(testNamespace).Create(&machine)
	spClient.SshproviderV1alpha1().ProvisionedMachines(testNamespace).Create(&pm)

	// Pull from APIs to state
	if err := s.PullFromAPIs(); err != nil {
		t.Fatal(err)
	}

	// Delete from APIs
	if err := kubeClient.CoreV1().Secrets(testNamespace).Delete(secret.Name, &metav1.DeleteOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := clusterClient.ClusterV1alpha1().Clusters(testNamespace).Delete(cluster.Name, &metav1.DeleteOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := clusterClient.ClusterV1alpha1().Machines(testNamespace).Delete(machine.Name, &metav1.DeleteOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := spClient.SshproviderV1alpha1().ProvisionedMachines(testNamespace).Delete(pm.Name, &metav1.DeleteOptions{}); err != nil {
		t.Fatal(err)
	}

	// Restore from state
	if err := s.PushToAPIs(); err != nil {
		t.Fatal(err)
	}

	// Verify objects restored
	if _, err := kubeClient.CoreV1().Secrets(testNamespace).Get(secret.Name, metav1.GetOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := clusterClient.ClusterV1alpha1().Clusters(testNamespace).Get(cluster.Name, metav1.GetOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := clusterClient.ClusterV1alpha1().Machines(testNamespace).Get(machine.Name, metav1.GetOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := spClient.SshproviderV1alpha1().ProvisionedMachines(testNamespace).Get(pm.Name, metav1.GetOptions{}); err != nil {
		t.Fatal(err)
	}
}
