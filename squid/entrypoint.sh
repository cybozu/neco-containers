#!/bin/sh -e

squid -z
touch /var/log/squid/access.log
touch /var/log/squid/store.log
touch /var/log/squid/cache.log
chown -R proxy:proxy /var/log/squid
squid $@
tail -F /var/log/squid/*.log
