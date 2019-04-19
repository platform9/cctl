# What is cctl ?

cctl is a cluster lifecycle management tool, powered by the [Kubernetes Cluster API](https://github.com/kubernetes-sigs/cluster-api) alongside [etcdadm](https://github.com/kubernetes-sigs/etcdadm) and [nodeadm](https://github.com/platform9/nodeadm).  

We refer to the trifecta of these three projects, which enable completely SSH driven bootstrapping and installation of highly available K8s clusters in any air-gapped or cloud environment, as _klusterkit_.

## Features

* Highly-available Kubernetes control plane and etcd
* Deploy & manage secure etcd clusters
* Works in air-gapped environments
* Rolling upgrade support with rollback capability
* Flannel (vxlan) CNI backend with plans to support other CNI backends
* Backup & recovery of etcd clusters from quorum loss
* Control plane protection from low memory/cpu situations
* No dependency on heavyweight configuration or infrastructure tooling.

## Installation

For users: The easiest thing to do is use docker to build the linux cctl binary, which you'll run on a linux machine where you are bootstrapping your cluster.

- Docker: `make container-build` will install the `cctl` binary in your current directory.

For developers: 
- Go: `go get -u github.com/platform9/cctl` will install `cctl` to your `$GOPATH/bin` directory.
- Development: `make`.

## Quick Start !

If you have a few VMs laying around and an SSH key, you can try out CCTL right now !
```
# Create the credentials used for the cluster.  
# Obviously, make sure your SSH key can get into the target nodes!
cctl create credential --user root --private-key ~/.ssh/id_rsa

# Now create a cluster object. Use `--help` to see a list of supported flags. 
cctl create cluster --pod-network 192.168.0.0/16 --service-network 192.169.0.0/24

# Finally, bootstrap your cluster, via SSH.
cctl create machine --ip $MACHINE_IP --role master
```

## Usage

`cctl` itself is a meant to be a reusable, lightweight cluster management for the _klusterkit_ toolchain.  It can be utilized in other toolchain's for building your own Kubernetes management tooling.  

CCTL thus supports the following sub-commands:

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

#### For detailed documentation see [wiki](https://github.com/platform9/cctl/wiki)
