FROM quay.io/cybozu/ubuntu-debug:20.04

# Avoid bird post-installation script error
# See https://bird.network.cz/pipermail/bird-users/2019-December/014075.html
COPY include-bird /etc/dpkg/dpkg.cfg.d/include-bird

COPY bootstrap.sh /

RUN apt-get update \
&& apt-get install -y software-properties-common \
&& add-apt-repository -y ppa:smoser/swtpm \
&& apt-get update \
&& apt-get -y install --no-install-recommends \
      git \
      build-essential \
      less \
      wget \
      systemd-container \
      lldpd \
      qemu \
      qemu-kvm \
      socat \
      picocom \
      swtpm \
      cloud-utils \
      bird2 \
      squid \
      chrony \
      dnsmasq \
      xauth \
      bash-completion \
      dbus \
      jq \
      libgpgme11 \
      freeipmi-tools \
      unzip \
      fakeroot \
      time \
      kmod \
      iptables \
&& rm -rf /var/lib/apt/lists/*

ENV GOPATH=/go

ENV PATH=/go/bin:/usr/local/go/bin:"$PATH"

RUN curl -sSLf https://dl.google.com/go/go1.15.8.linux-amd64.tar.gz | tar -C /usr/local -xzf -

RUN curl -sfL https://github.com/cybozu-go/placemat/releases/download/v2.0.2/placemat2_2.0.2_amd64.deb -o placemat2_2.0.2_amd64.deb \
&& dpkg -i ./placemat2_2.0.2_amd64.deb \
&& rm ./placemat2_2.0.2_amd64.deb

ENTRYPOINT ["/bootstrap.sh"]
