#!/bin/bash

function intro() {
	echo "Welcome to the KlusterKit interactive Demo !!!"
}

function compilation() {
	if [[ -f "cctl-osx" ]] ; then
		echo "Detected cctl binary, no need to compile it..."	
	else
		echo "First we'll compile the CCTL tool for OS X... "
		pushd ../
			make container-build
			cp cctl* ./hack
		popd
	fi
}


function vagrant_up() {
	echo "Ok, now lets spin up some infrastructure..."
	pushd ../
		vagrant up
	popd
	
	echo "Typically, you'll replace this step with a terraform or other cloud provisioning step"
	echo "Note that we will need to be able to SSH into these VMs for KlusterKit to work!"
}

function initialize() {
	VM="172.100.100.3"
	CCTL="./cctl-osx"

	set -x
	echo "credential"
	$CCTL create credential --state cctl-state.yaml --user vagrant --private-key ~/.ssh/id_rsa
	echo "cluster..."
	$CCTL create cluster --state cctl-state.yaml --pod-network 192.168.0.0/16 --service-network 192.169.0.0/24
	echo "machine..."
	$CCTL create machine --iface eth1 --state cctl-state.yaml --ip 172.100.100.3 --role master
	$CCTL create machine --iface eth1 --state cctl-state.yaml --ip 172.100.100.2 --role worker
}

intro
compilation
vagrant_up
initialize

