# bpftrace

![Regular Update](../regular_update.svg)

1. Check the bpftrace [release page](https://github.com/bpftrace/bpftrace/releases).
2. Check the bpftrace [Dockerfile](https://github.com/bpftrace/bpftrace/blob/master/docker/Dockerfile.ubuntu) of the latest release version.
3. Check the bcc [release page](https://github.com/iovisor/bcc/releases) and update `BCC_VERSION` and `BCC_SHA` in the `Makefile` to its latest version.
4. Check the libbpf [release page](https://github.com/libbpf/libbpf/releases) and update `LIBBPF_SHA` in the `Makefile` to its latest version.
5. Update `BPFTRACE_SHA` in the `Makefile`.
6. Update `TAG` file.
