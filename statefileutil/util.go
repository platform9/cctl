package statefileutil

import (
	"io/ioutil"
	"os"

	"github.com/platform9/pf9-clusteradm/cmd"
	"gopkg.in/yaml.v2"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

const (
	STATE_FILE_PATH = "/tmp/cluster-state.yaml"
)

func checkFileExists() (bool, error) {
	_, err := os.Stat(STATE_FILE_PATH)
	if err == nil {
		return true, nil
	}
	return false, err
}

func ReadStateFile() (*cmd.ClusterState, error) {
	cs := new(cmd.ClusterState)
	ret, _ := checkFileExists()
	if ret == false {
		return cs, nil
	}
	d, err := ioutil.ReadFile(STATE_FILE_PATH)
	if err != nil {
		return nil, err
	}
	yaml.Unmarshal(d, cs)
	return cs, nil
}

func WriteStateFile(cs *cmd.ClusterState) error {
	if cs != nil {
		bytes, _ := yaml.Marshal(cs)
		ioutil.WriteFile(STATE_FILE_PATH, bytes, 0600)
	}
	return nil
}

func GetClusterSpec() (*clusterv1.Cluster, error) {
	cs, err := ReadStateFile()
	if cs != nil {
		return &cs.Cluster, nil
	}
	return nil, err
}

func GetMachinesSpec() ([]clusterv1.Machine, error) {
	cs, err := ReadStateFile()
	if cs != nil {
		return cs.Machines, nil
	}
	return nil, err
}
