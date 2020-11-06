## cloudkit core
```
createdb cloud_kit_dev
```

### Temp notes on spinning up a cloudkit server host
```
sudo apt-get update && sudo apt install net-tools qemu-kvm libvirt-clients libvirt-daemon-system bridge-utils virt-manager libguestfs-tools cloud-image-utils -y
```

edit the libvirtd config: vim /etc/libvirt/libvirtd.conf
```
listen_tls = 0
listen_tcp = 1
auth_tcp = "none"
tls_no_verify_certificate = 1
```

start libvirt daemon in background and listen on grpc endpoint
```
systemctl stop libvirtd
libvirtd -d -l
```

fetch and build ubuntu VM
```
wget https://cloud-images.ubuntu.com/bionic/current/bionic-server-cloudimg-amd64.img
qemu-img resize bionic-server-cloudimg-amd64.img 10G
qemu-img convert -f qcow2 bionic-server-cloudimg-amd64.img /var/lib/libvirt/images/ubuntu-bionic.img
```

set up cloud config
```
touch cloud.txt && vim cloud.txt

#cloud-config
password: ubuntu
chpasswd: { expire: False }
ssh_pwauth: True
hostname: ubuntu-bionic

cloud-localds /var/lib/libvirt/images/ubuntu-bionic.iso cloud.txt
```

check machine info, mac, ip, etc
```
virsh net-dhcp-leases default
virsh domifaddr {vm_name}
virsh domiflist {vm_name}
virsh console {vm_name}
virsh net-dumpxml default | egrep 'range|host\ mac'
```

Hop:
```
ssh -t root@167.172.219.248 ssh ubuntu@192.168.122.80
```
