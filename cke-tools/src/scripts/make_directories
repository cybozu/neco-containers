#!/usr/bin/python3

from argparse import ArgumentParser
import os
import sys


def octint(v: str) ->int:
    return int(v, 8)


def main():
    p = ArgumentParser()
    p.add_argument('--mode', type=octint, default='755',
                   help='permission bits for directories')
    p.add_argument('dirs', metavar='DIR', nargs='+', help='directories to be created')

    ns = p.parse_args()
    print(ns.dirs)

    for d in ns.dirs:
        if not os.path.isabs(d):
            sys.exit('DIR must be an asbolute path')

    for d in ns.dirs:
        os.makedirs(d, mode=ns.mode, exist_ok=True)
        os.chmod(d, ns.mode)


if __name__ == '__main__':
    main()
