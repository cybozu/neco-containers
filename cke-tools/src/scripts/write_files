#!/usr/bin/python3

from argparse import ArgumentParser
import os
import shutil
import sys
import tarfile


def octint(v: str) ->int:
    return int(v, 8)


def main():
    p = ArgumentParser(description='Read TAR from stdin and write out the contents')
    p.add_argument('--mode', type=octint, default='644',
                   help='permission bits for the file')
    p.add_argument('prefix', metavar='PREFIX', help='root directory to extract contents')

    ns = p.parse_args()

    tar = tarfile.open(mode='r|', fileobj=sys.stdin.buffer)
    for t in tar:
        if not t.isreg():
            continue
        p = os.path.normpath(ns.prefix + "/" + t.name)
        pd = os.path.dirname(p)
        os.makedirs(pd, 0o755, exist_ok=True)
        with open(p, "wb") as f:
            shutil.copyfileobj(tar.extractfile(t), f)
            os.chmod(p, ns.mode)
            os.fsync(f.fileno())
    tar.close()


if __name__ == '__main__':
    main()
