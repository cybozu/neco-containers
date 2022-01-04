# grafana_plugins_init container

FROM quay.io/cybozu/ubuntu:20.04

ARG SRCREPO=grafana-operator/grafana_plugins_init
ARG GRAFANA_PLUGINS_INIT_VERSION=0.0.5

RUN apt-get update && \
    apt-get install -y --no-install-recommends python3 && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* && \
    curl -o /plugins.py https://raw.githubusercontent.com/${SRCREPO}/${GRAFANA_PLUGINS_INIT_VERSION}/plugins.py

USER 10000:10000

CMD [ "python3", "/plugins.py"]
