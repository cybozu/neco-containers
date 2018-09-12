FROM centos:centos7

ENV PATH $PATH:/opt/dell/srvadmin/bin:/opt/dell/srvadmin/sbin

ARG OMSA_VERSION=18.08.00

RUN rpm --import https://linux.dell.com/repo/hardware/DSU_${OMSA_VERSION}/public.key && \
rpm --import https://linux.dell.com/repo/hardware/DSU_${OMSA_VERSION}/public_gpg3.key

COPY dell-system-update.repo /etc/yum.repos.d/dell-system-update.repo
RUN sed "s/@@OMSA_VERSION@@/${OMSA_VERSION}/" -i /etc/yum.repos.d/dell-system-update.repo

RUN yum update -y && \
    yum -y install https://centos7.iuscommunity.org/ius-release.rpm && \
    yum -y install srvadmin-omcommon srvadmin-server-cli srvadmin-storageservices-cli srvadmin-idrac7 dmidecode sysvinit-tools python36u && \
    yum clean all

COPY setup-hw /usr/local/omsa/bin/setup-hw
COPY install-tools /usr/local/omsa/bin/install-tools

ENV PATH=/usr/local/omsa/bin:"$PATH"

ENV SYSTEMCTL_SKIP_REDIRECT=1

CMD /etc/init.d/instsvcdrv start && \
    /etc/init.d/dataeng start && \
    tail -f /opt/dell/srvadmin/var/log/openmanage/dcsys64.xml
