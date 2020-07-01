# tail

[tail](https://github.com/nxadm/tail) outputs the last part of files.

In Kubernetes, this container can be used as a sidecar container to output logs to stdout.

## Usage

To launch tail:

    $ docker run quay.io/cybozu/tail:1.4 -f /path/to/logfile
 
## Docker images

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/tail)
