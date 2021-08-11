#!/bin/bash

# /etc/proxysql.cnf.d directory is used for k8s emptydir share, it is not designed for merge config with /etc/proxy.cnf
CONFIG_FILE=${CONFIG_FILE:-/etc/proxysql.cnf.d/proxysql.cnf}
if [[ ! -f $CONFIG_FILE ]];then
    CONFIG_FILE=/etc/proxysql.cnf
fi
DATADIR=${DATADIR:-/var/lib/proxysql}

proxysql -f -D $DATADIR -c $CONFIG_FILE