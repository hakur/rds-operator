#!/bin/bash

CONFIG_FILE=${CONFIG_FILE:-/etc/nutcracker.yml}
LISTEN_PORT=${LISTEN_PORT:-6379}

# create config file dir  if not eixsts
configFileDir=$(echo $CONFIG_FILE | sed "s/$(basename $CONFIG_FILE)//")
if [[ ! -d $configFileDir ]];then
    mkdir -p $configFileDir
fi
# write Redis Servers list to config file,values are from REDIS_NODES env
function writeRedisServers() {
    local serverIndex=0
    for i in $REDIS_NODES;do 
        let serverIndex+=1
        echo "   - ${i}:1 server-${serverIndex}" >> $CONFIG_FILE
    done
}

cat > $CONFIG_FILE << EOF
beta:
  listen: 0.0.0.0:6379
  hash: fnv1a_64
  hash_tag: "{}"
  distribution: ketama
  auto_eject_hosts: true
  timeout: 400
  redis: true
  redis_auth: ${REDIS_PASSWORD}
  servers:
EOF

writeRedisServers

echo "server config file content:"
cat $CONFIG_FILE

nutcracker -c $CONFIG_FILE