Vagrant.configure("2") do |config|
    config.vm.define "archlinux"

    config.vm.box = "archlinux/archlinux"
    config.vm.box_version = "20210601.24453"

    config.vm.provision "shell", inline: <<-SHELL
        # Packages
        pacman -Syy

        # wget
        pacman -S --noconfirm wget

        # Golang
        cd /tmp
        wget https://golang.org/dl/go1.16.5.linux-amd64.tar.gz
        rm -rf /usr/local/go
        tar -C /usr/local -xzf go1.16.5.linux-amd64.tar.gz
        echo -e "\nexport PATH=\$PATH:/usr/local/go/bin\n" > /home/vagrant/.bashrc
        rm -rf go1.16.5.linux-amd64.tar.gz
    SHELL

    config.vm.synced_folder ".", "/vagrant"
end
