# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure(2) do |config|
  $num_instances = 2
  (1..$num_instances).each do |i|
    config.vm.define vm_name = "centos-7-%02d.vagrant.test"  % [i] do |config|

    config.vm.box = "centos/7"

    ip = "172.100.100.#{i+2}"

    config.vm.network :private_network, ip: ip
      config.vm.provider "virtualbox" do |vb|
        config.vm.synced_folder "./hack/", "/vagrant", type: "rsync", rsync__args: ["--verbose", "--archive", "--delete", "-z"]
        vb.memory = "1028"
        vb.cpus = "1"
      end
      config.vm.provision "shell" do |s|
        ssh_pub_key = File.readlines("#{Dir.home}/.ssh/id_rsa.pub").first.strip
        s.inline = <<-SHELL

          echo #{ssh_pub_key} >> /home/vagrant/.ssh/authorized_keys
          sudo mkdir -p /opt/bin
          sudo chmod -R 777 /opt/bin
          sudo mkdir -p /var/cache/ssh-provider/nodeadm/v0.2.1 
          sudo mkdir -p /var/cache/ssh-provider/etcdadm/v0.1.1

          sudo chmod -R 777 /var/cache/ssh-provider/

          cp /vagrant/nodeadm-linux /var/cache/ssh-provider/nodeadm/v0.2.1/nodeadm
          cp /vagrant/etcdadm-linux /var/cache/ssh-provider/etcdadm/v0.1.1/etcdadm

          sudo yum install -y yum-utils device-mapper-persistent-data lvm2 ; sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo ; sudo yum install -y docker-ce ; sudo usermod -aG docker vagrant
  
          echo "Starting docker daemon..."
          sudo systemctl start docker

        SHELL
      end
     end
  end
end
