# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
    config.vm.define "node1", primary: true do |node1|
        node1.vm.box = "opscode-f19-64"
        node1.vm.box_url = "https://opscode-vm-bento.s3.amazonaws.com/vagrant/opscode-fedora-19_provisionerless.box"
        #node1.vm.synced_folder "~/code/eureka", "/eureka-repo"
        node1.vm.provision :shell, :inline => "hostname 'node1.localdomain'; echo 'node1.localdomain' > /etc/hostname"
        node1.vm.provision :shell, :path => "provision_eureka.sh"
        node1.vm.network :private_network, ip: "172.16.0.11"
        node1.vm.provider :virtualbox do |vb|
            vb.gui = false
            vb.customize ["modifyvm", :id, "--memory", "2500"]
            vb.customize ["modifyvm", :id, "--cpus", "4"]
            # This allows symlinks to be created within the /vagrant root directory
            vb.customize ["setextradata", :id, "VBoxInternal2/SharedFoldersEnableSymlinksCreate/v-root", "1"]
        end
    end
    config.vm.define "node2" do |node2|
        node2.vm.box = "opscode-f19-64"
        node2.vm.box_url = "https://opscode-vm-bento.s3.amazonaws.com/vagrant/opscode-fedora-19_provisionerless.box"
        #node2.vm.synced_folder "~/code/eureka", "/eureka-repo"
        node2.vm.provision :shell, :inline => "hostname 'node2.localdomain'; echo 'node2.localdomain' > /etc/hostname"
        node2.vm.provision :shell, :path => "provision_eureka.sh"
        node2.vm.network :private_network, ip: "172.16.0.22"
        node2.vm.provider :virtualbox do |vb|
            vb.gui = false
            vb.customize ["modifyvm", :id, "--memory", "2500"]
            vb.customize ["modifyvm", :id, "--cpus", "4"]
            # This allows symlinks to be created within the /vagrant root directory
            vb.customize ["setextradata", :id, "VBoxInternal2/SharedFoldersEnableSymlinksCreate/v-root", "1"]
        end
    end

end
