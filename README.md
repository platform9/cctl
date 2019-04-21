# cctl

cctl is a cluster lifecycle management tool that adopts the Kubernetes community's Cluster API and uses nodeadm and etcdadm to easily deploy and maintain highly-available Kubernetes clusters in on-premises, even air-gapped environments.  

Along with [etcdadm](https://github.com/kubernetes-sigs/etcdadm) and [nodeadm](https://github.com/platform9/nodeadm), this tool makes up _klusterkit_, which lets you create, scale, backup and restore your air-gapped, on-premise Kubernetes cluster.

## Features
* Highly-available Kubernetes control plane and etcd
* Deploy & manage secure etcd clusters
* Works in air-gapped environments
* Rolling upgrade support with rollback capability
* Flannel (vxlan) CNI backend with plans to support other CNI backends
* Backup & recovery of etcd clusters from quorum loss
* Control plane protection from low memory/cpu situations

## Installation
```
go get -u github.com/platform9/cctl
```

Quick demo: cd hack/ and follow the directions for the Vagrant recipe!

## Usage
```
$GOPATH/bin/cctl [command]
```
### Available Commands: 
```
  backup      Create an archive with the current cctl state and an etcd snapshot from the cluster.
  bundle      Used to create cctl bundle
  create      Used to create resources
  delete      Used to delete resources
  deploy      Used to deploy app to the cluster
  get         Display one or more resources
  help        Help about any command
  migrate     Migrate the state file to the current version
  recover     Used to recover the cluster
  restore     Restore the cctl state and etcd snapshot from an archive.
  snapshot    Used to get a snapshot
  status      Used to get status of the cluster
  upgrade     Used to upgrade the cluster
  version     Print version information
```

## Getting Started 

If your setup has internet connectivity, follow these steps. For an airgapped environment, please see documentation [wiki](https://github.com/platform9/cctl/wiki).

On all nodes that will make up your Kubernetes cluster, ensure that:
- The docker container runtime is installed and the docker daemon is running.
- The `etcdadm` binary is in the `/var/cache/ssh-provider/etcdadm/<version>/` directory, and the `nodeadm` binary is in `/var/cache/ssh-provider/nodeadm/<version>/` directory. To find the versions required by the `cctl` release you use, see the [releases](https://github.com/platform9/cctl/releases) page.

First, create the credentials used for the cluster.
```
$GOPATH/bin/cctl create credential --user root --private-key ~/.ssh/id_rsa
```

Then, create a cluster object. Use `--help` to see a list of supported flags. 
```
$GOPATH/bin/cctl create cluster --pod-network 192.168.0.0/16 --service-network 192.169.0.0/24
```

Finally, create the first machine in your cluster.
```
$GOPATH/bin/cctl create machine --ip $MACHINE_IP --role master
```


#### For detailed documentation see [wiki](https://github.com/platform9/cctl/wiki)
