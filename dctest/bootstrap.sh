#!/bin/sh -xe

git clone https://github.com/cybozu-go/neco.git /neco
cd /neco
git checkout poc-run-placemat-on-k8s
cd dctest
make setup SUDO=""
make placemat SUDO=""
make test SUITE=bootstrap SUDO=""
sleep infinity

