# cctl

CLI tool for Kubernetes cluster management. This tool lets you create, scale, backup and restore your air-gapped, on-premise Kubernetes cluster.

## Installation
```
go get -u github.com/platform9/cctl
```
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

Ensure the correct version of the `nodeadm` and `etcdadm` binaries are placed in the `/opt/bin` directory of all nodes that will make up your cluster. 

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

