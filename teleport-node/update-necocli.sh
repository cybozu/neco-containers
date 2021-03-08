#! /bin/sh -ex

export HTTPS_PROXY="http://squid.internet-egress.svc:3128"

DATE=$(curl -s "https://api.github.com/repos/cybozu-go/neco/releases/latest" | grep -Po '"tag_name": "\K.*?(?=")' | sed -e "s/release-//")
FILE="neco-operation-cli-linux_${DATE}_amd64.deb"

ADD_PATH="${HOME}/neco-operation-cli/usr/bin"
if [ ! -f "${HOME}/.profile" ]  || [ ! "$(cat ${HOME}/.profile | grep ${ADD_PATH})" ]; then
    echo "export PATH=${PATH}:${HOME}/neco-operation-cli/usr/bin" >> ${HOME}/.profile
fi

if [ ! -f $FILE ]; then
    curl -sLf -O https://github.com/cybozu-go/neco/releases/download/release-${DATE}/neco-operation-cli-linux_${DATE}_amd64.deb
    mkdir -p ${HOME}/neco-operation-cli
    dpkg -x $FILE ${HOME}/neco-operation-cli
fi
