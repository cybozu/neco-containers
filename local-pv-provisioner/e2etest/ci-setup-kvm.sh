#!/usr/bin/bash -xeu

set -o pipefail

sudo apt-get install qemu-kvm libvirt-daemon-system libvirt-clients bridge-utils
kvm-ok
sudo adduser `id -un` libvirt
sudo adduser `id -un` kvm
virsh list --all
sudo ls -la /var/run/libvirt/libvirt-sock
sudo chmod 777 /var/run/libvirt/libvirt-sock
sudo ls -la /var/run/libvirt/libvirt-sock
ls -l /dev/kvm
