#!/bin/sh -e

squid -z
squid $@
touch /var/log/squid/access.log
touch /var/log/squid/store.log
touch /var/log/squid/cache.log
tail -F /var/log/squid/*.log
