# grafana_plugins_init container

FROM quay.io/cybozu/ubuntu:20.04

ARG SRCREPO=grafana-operator/grafana_plugins_init
ARG REVISION=586394f88c1d6f3b2a83ba1c2e67c7f056e564fd

RUN apt-get update && \
    apt-get install -y --no-install-recommends python3 && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* && \
    curl -o /plugins.py https://raw.githubusercontent.com/${SRCREPO}/${REVISION}/plugins.py

USER 10000:10000

CMD [ "python3", "/plugins.py"]
