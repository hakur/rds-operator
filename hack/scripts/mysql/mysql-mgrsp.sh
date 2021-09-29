#!/bin/bash
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


function findMasterServerMGRSP() {
    for server in $MYSQL_NODES;do
        mysqladmin -u$MYSQL_USER  -P$MYSQL_PORT -h$server ping 2>/dev/null > /dev/null 
        if [ $? == 0 ] ;then
            local serverUUID=$(mysql -u$MYSQL_USER  -P$MYSQL_PORT -h$server -N -se "show variables like 'server_uuid';"  | awk '{print $2}')
            local primaryServerUUID=$(mysql -u$MYSQL_USER  -P$MYSQL_PORT -h$server -N -se "SELECT VARIABLE_VALUE FROM performance_schema.global_status WHERE VARIABLE_NAME= 'group_replication_primary_member' AND VARIABLE_VALUE='${serverUUID}';"  )
            if [[ "$primaryServerUUID" == "$serverUUID" ]] && [[ "$serverUUID" != "" ]];then
                echo $server
                break
            fi
        fi
    done
}


function checkGroupReplicationIsRunning() {
    mgrOn=$(mysql -u$MYSQL_USER  -N -se "select * from performance_schema.replication_applier_status;"  | grep group_replication_applier | awk '{print $2}' )
    echo $mgrOn
}


function  mysqlMGRSPJoinMembers() {
    echolog "start join ..."
    mysql -u$MYSQL_USER  -e "CHANGE MASTER TO MASTER_USER='${MYSQL_REPL_USER}',MASTER_PASSWORD='${MYSQL_REPL_PASSWORD}' FOR CHANNEL 'group_replication_recovery';" 
    mysql -u$MYSQL_USER  -e "START group_replication;" 
}

function mysqlMGRSPBootMembers() {
    echolog "start bootstrap ..."
    mysql -u$MYSQL_USER  -e "SET GLOBAL group_replication_bootstrap_group=ON;" 
    mysql -u$MYSQL_USER  -e "START group_replication;" 
    mysql -u$MYSQL_USER  -e "SET GLOBAL group_replication_bootstrap_group=OFF;" 
}
