#! /bin/sh -ex

export HTTPS_PROXY="http://squid.internet-egress.svc:3128"

DATE=$(curl -s "https://api.github.com/repos/cybozu-go/neco/releases/latest" | jq -r ".tag_name" | sed -e "s/release-//")
FILE="neco-operation-cli-linux_${DATE}_amd64.deb"

CLI_PATH="${HOME}/neco-operation-cli/usr/bin"
if [ ! -f "${HOME}/.profile" ]  || [ ! "$(cat ${HOME}/.profile | grep ${CLI_PATH})" ]; then
    echo "export PATH=\${PATH}:${HOME}/neco-operation-cli/usr/bin" >> ${HOME}/.profile
fi

if [ ! -f ${HOME}/deb/${FILE} ]; then
    echo "Downloading and extracting Neco CLI tools..."
    curl -sLf -O https://github.com/cybozu-go/neco/releases/download/release-${DATE}/${FILE}
    mkdir -p ${HOME}/neco-operation-cli
    dpkg -x ${FILE} ${HOME}/neco-operation-cli
    mkdir -p ${HOME}/deb
    mv ${FILE} ${HOME}/deb/
fi
