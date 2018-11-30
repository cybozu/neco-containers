#!/bin/sh -e

chown proxy:proxy /dev/stdout
squid -N -z
exec squid -N $@
