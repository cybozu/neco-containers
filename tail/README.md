# tail

[tail](https://github.com/nxadm/tail) outputs the last part of files.

In Kubernetes, this container can be used as a sidecar container to output logs to stdout.

## Usage

To launch tail:

    $ docker run quay.io/cybozu/tail:0 -f /path/to/file1.log /path/to/file2.log
 
## Docker images

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/tail)
