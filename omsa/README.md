[![Docker Repository on Quay](https://quay.io/repository/cybozu/omsa/status?token=aecbaf01-41ea-4e2c-9b6d-b6dd9ad533f5 "Docker Repository on Quay")](https://quay.io/repository/cybozu/omsa)

OMSA container
==============

This directory provides a Dockerfile to build a Docker container that runs [Dell EMC OMSA(Opan Manage Server Administrator)](https://www.dell.com/support/contents/us/en/04/article/product-support/self-support-knowledgebase/enterprise-resource-center/systemsmanagement/omsa?lwp=rt)

Usage
-----

### Generate omsa.json if template exists

```console
$ sudo rkt run \
    --volume=host,kind=host,source=/ \
    --mount=volume=host,target=/host \
    quay.io/cybozu/omsa:v18.08.00-1 \
    --exec install-tools
```

### Run as daemon

```console
$ sudo rkt run \
  --insecure-options=all \
  --volume modules,kind=host,source=/lib/modules/$(uname -r),readOnly=true \
  --mount volume=modules,target=/lib/modules/$(uname -r) \
  --volume dev,kind=host,source=/dev \
  --mount volume=dev,target=/dev \
  --mount volume=neco,target=/etc/neco"
  --volume neco,kind=host,source=/etc/neco,readOnly=true \
  quay.io/cybozu/omsa:v18.08.00-1 \
  --name omsa
```

### Run setup-hw

```console
$ sudo rkt enter --app omsa UUID setup-hw
```

### Run omreport

```console
$ sudo rkt enter --app omsa UUID omreport about
$ sudo rkt enter --app omsa UUID omreport bios
```
