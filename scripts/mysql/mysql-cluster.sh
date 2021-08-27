#!/bin/bash
MYSQL_CLUSTER_MODE=${MYSQL_CLUSTER_MODE:-SemiSync}
MYSQL_USER=${MYSQL_USER:-root}
MYSQL_PASSWORD=${MYSQL_PASSWORD:-$MYSQL_ROOT_PASSWORD}
MYSQL_NODES=${MYSQL_NODES}
MYSQL_PORT=${MYSQL_PORT-3306}
MYSQL_REPL_USER=${MYSQL_REPL_USER}
MYSQL_REPL_PASSWORD=${MYSQL_REPL_PASSWORD}

################## master infomation functions ########################

function FindMasterServer() {
    local masterAddr=""

    if [ "$MYSQL_CLUSTER_MODE" == "SemiSync" ];then
        masterAddr=$(findMasterServerSymiSync)
        if [ "$masterAddr" != "" ];then
            echo $masterAddr
        fi
    fi
    if [ "$MYSQL_CLUSTER_MODE" == "MGRSP" ];then
        masterAddr=$(findMasterServerMGRSP)
        if [ "$masterAddr" != "" ];then
            echo $masterAddr
        fi
    fi
}

function findMasterServerSymiSync() {
    local masterOn="OFF"
    for server in $MYSQL_NODES;do 
        mysqladmin -uroot -p${MYSQL_PASSWORD} -P$MYSQL_PORT -h$server ping > /dev/null 2>/dev/null
        if [ $? == 0 ] ;then
            masterOn=$(mysql -uroot -p${MYSQL_PASSWORD} -P$MYSQL_PORT -h$server -N -se "show  variables like 'rpl_semi_sync_master_enabled';" 2>/dev/null | awk '{print $2}')
            if [ "$masterOn" == "ON" ];then
                echo $server
                break
            fi
        fi
    done
}

function findMasterServerMGRSP() {
    for server in $MYSQL_NODES;do
        mysqladmin -uroot -p${MYSQL_PASSWORD} -P$MYSQL_PORT -h$server ping > /dev/null 2>/dev/null
        if [ $? == 0 ] ;then
            local serverUUID=$(mysql -u$MYSQL_USER -p$MYSQL_PASSWORD -P$MYSQL_PORT -h$server -N -se "show variables like 'server_uuid';" 2>/dev/null | awk '{print $2}')
            local primaryServerUUID=$(mysql -u$MYSQL_USER -p$MYSQL_PASSWORD -P$MYSQL_PORT -h$server -N -se "SELECT VARIABLE_VALUE FROM performance_schema.global_status WHERE VARIABLE_NAME= 'group_replication_primary_member' AND VARIABLE_VALUE='${serverUUID}';" 2>/dev/null )
            if [[ "$primaryServerUUID" == "$serverUUID" ]] && [[ "$serverUUID" != "" ]];then
                echo $server
                break
            fi
        fi
    done
}

################################# cluster functions #################################################

function StartCluster() {
    local code=0

    if [ "$MYSQL_CLUSTER_MODE" == "SemiSync" ];then
        code=$?
    fi

    if [ "$MYSQL_CLUSTER_MODE" == "MGRSP" ];then
        StartMGRSPCluster
        code=$?
    fi

    if [ $code != 0 ];then
        exit 0
    fi
}

function StartMGRSPCluster() {
    local masterAddr=$(FindMasterServer)
    local serverID=${HOSTNAME##*-}
    let serverID+=1

    local mgrOn=$(checkGroupReplicationIsRunning)
    if [ "$mgrOn" == "ON" ];then
        return 0
    fi

    if [[ "$masterAddr" == "" ]] && [[ "$serverID" -eq 1 ]];then
        mysqlMGRSPBootMembers
    else
        mysqlMGRSPJoinMembers 
    fi
}

function checkGroupReplicationIsRunning() {
    mgrOn=$(mysql -uroot -p${MYSQL_ROOT_PASSWORD} -N -se "select * from performance_schema.replication_applier_status;" 2>/dev/null | grep group_replication_applier | awk '{print $2}' )
    echo $mgrOn
}


function  mysqlMGRSPJoinMembers() {
    echolog "start join ..."
    mysql -uroot -p${MYSQL_ROOT_PASSWORD} -e "CHANGE MASTER TO MASTER_USER='${MYSQL_REPL_USER}',MASTER_PASSWORD='${MYSQL_REPL_PASSWORD}' FOR CHANNEL 'group_replication_recovery';" 2>/dev/null
    mysql -uroot -p${MYSQL_ROOT_PASSWORD} -e "START group_replication;" 2>/dev/null
}

function mysqlMGRSPBootMembers() {
    echolog "start bootstrap ..."
    mysql -uroot -p${MYSQL_ROOT_PASSWORD} -e "SET GLOBAL group_replication_bootstrap_group=ON;" 2>/dev/null
    mysql -uroot -p${MYSQL_ROOT_PASSWORD} -e "START group_replication;" 2>/dev/null
    mysql -uroot -p${MYSQL_ROOT_PASSWORD} -e "SET GLOBAL group_replication_bootstrap_group=OFF;" 2>/dev/null
}
