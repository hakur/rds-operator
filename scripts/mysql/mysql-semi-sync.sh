#!/bin/bash

# SEMI_SYNC_FIXED_MASTERS master address ,splited by space,max only support two master adresses
SEMI_SYNC_FIXED_MASTERS=${SEMI_SYNC_FIXED_MASTERS} 

function StartSemiSyncCluster() {
    if [[ "$HOSTNAME" == *"$SEMI_SYNC_FIXED_MASTERS"* ]];then
        semiSyncBootMembers
    else
        local masterAddr=$(FindMasterServer)
        semiSyncJoinMaster $masterAddr "1"
    fi
}

function semiSyncJoinMaster() {
    local master=$1
    local superReadOnly=$2

    if [[ "$master" == "" ]] || [[ "$superReadOnly" == "" ]];then
        echolog "master host address or superReadOnly is empty,exit 1 now"
        exit 1
    fi

    local slaveOn=$(mysql -u$MYSQL_USER  -N -se "show variables like 'rpl_semi_sync_master_enabled';"  | awk '{print $2}')
    local myMaster=$(mysql -u$MYSQL_USER  -N -se "show slave status \G"  | grep Master_Host | awk '{print $2}')
    # if my master is not current cluster master
    if [ "$slaveOn" != "ON" ];then
        mysql -u$MYSQL_USER  -N -se "SET GLOBAL rpl_semi_sync_slave_enabled=ON"
        mysql -u$MYSQL_USER  -N -se "SET GLOBAL super_read_only=$superReadOnly"
    fi

    if [ "$myMaster" != "$master" ];then
        mysql -u$MYSQL_USER  -N -se "STOP SLAVE"
        mysql -u$MYSQL_USER  -N -se "CHANGE MASTER TO MASTER_HOST='${master}',MASTER_USER='${MYSQL_REPL_USER}',MASTER_PASSWORD='${MYSQL_REPL_PASSWORD}',MASTER_AUTO_POSITION=1;"
        mysql -u$MYSQL_USER  -N -se "START SLAVE"
    fi
}

function semiSyncBootMembers() {
    # if master module not enabled
    local masterOn=$(mysql -u$MYSQL_USER  -N -se "show variables like 'rpl_semi_sync_master_enabled';"  | awk '{print $2}')
    if [ "$masterOn" != "ON" ];then
        mysql -u$MYSQL_USER  -N -se "SET GLOBAL rpl_semi_sync_master_enabled=ON"
        mysql -u$MYSQL_USER  -N -se "SET GLOBAL super_read_only=0"
    fi

    local master=""
    for i in $SEMI_SYNC_FIXED_MASTERS;do
        if [ "$i" != "$HOSTNAME" ];then
            master=$i" "
        fi
    done

    if [ "$master" != "" ];then
        semiSyncJoinMaster $master 0
    fi
}

function findMasterServerSymiSync() {
    local masterOn="OFF"
    for server in $MYSQL_NODES;do 
        mysqladmin -u$MYSQL_USER  -P$MYSQL_PORT -h $server ping 2>/dev/null > /dev/null 
        if [ $? == 0 ] ;then
            masterOn=$(mysql -u$MYSQL_USER  -P$MYSQL_PORT -h $server -N -se "show  variables like 'rpl_semi_sync_master_enabled';"  | awk '{print $2}')
            if [ "$masterOn" == "ON" ];then
                echo $server
                break
            fi
        fi
    done
}