#!/bin/bash
CONN_MAX_RETRIES=${CONN_MAX_RETRIES:-20}
CONN_MAX_RETRY_SLEEP=${CONN_MAX_RETRY_SLEEP:-3}

for ((i=1; i<=${CONN_MAX_RETRIES}; i ++));do
    echo "try to connect redis cluster at [ ${REDIS_NODES} ]"
    redis-cluster-proxy -a ${REDIS_PASSWORD} -p 6379 --bind 0.0.0.0 ${REDIS_NODES}
    if [ $? != 0 ];then
        sleep $CONN_MAX_RETRY_SLEEP
    else
        echo "bye"
        exit 0
    fi
done