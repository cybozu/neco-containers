#!/usr/bin/bash -xeu

set -o pipefail

VIRTUALIZATION_SUPPORT=$(grep -E -q 'vmx|svm' /proc/cpuinfo && echo yes || echo no)
echo ${VIRTUALIZATION_SUPPORT}
if [ "${VIRTUALIZATION_SUPPORT}" != "yes" ]; then
  echo "CPU does not support the virtualization feature."
  exit 1
fi
sudo apt-get install qemu-kvm libvirt-daemon-system libvirt-clients bridge-utils
kvm-ok
sudo adduser `id -un` libvirt
sudo adduser `id -un` kvm
virsh list --all
sudo ls -la /var/run/libvirt/libvirt-sock
sudo chmod 777 /var/run/libvirt/libvirt-sock
sudo ls -la /var/run/libvirt/libvirt-sock
ls -l /dev/kvm
sudo rmmod kvm_amd
sudo rmmod kvm
sudo modprobe -a kvm
sudo modprobe -a kvm_amd
