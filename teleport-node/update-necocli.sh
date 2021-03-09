export KUBERNETES_SERVICE_PORT_HTTPS="443"
export KUBERNETES_SERVICE_PORT="443"
export KUBERNETES_PORT_443_TCP="tcp://kubernetes.default.svc:443"
export KUBERNETES_PORT_443_TCP_PROTO="tcp"
export KUBERNETES_PORT_443_TCP_ADDR="kubernetes.default.svc"
export KUBERNETES_SERVICE_HOST="kubernetes.default.svc"
export KUBERNETES_PORT="tcp://kubernetes.default.svc:443"
export KUBERNETES_PORT_443_TCP_PORT="443"
export PATH="${PATH}:${HOME}/neco-operation-cli/usr/bin"

export HTTPS_PROXY="http://squid.internet-egress.svc:3128"
DATE=$(curl -s "https://api.github.com/repos/cybozu-go/neco/releases/latest" | jq -r ".tag_name" | sed -e "s/release-//")
FILE="neco-operation-cli-linux_${DATE}_amd64.deb"

if [ ! -f ${HOME}/deb/${FILE} ]; then
    echo "Downloading and extracting Neco CLI tools..."
    curl -sLf -O https://github.com/cybozu-go/neco/releases/download/release-${DATE}/${FILE}
    mkdir -p ${HOME}/neco-operation-cli
    dpkg -x ${FILE} ${HOME}/neco-operation-cli
    mkdir -p ${HOME}/deb
    mv ${FILE} ${HOME}/deb/
fi

unset HTTPS_PROXY
