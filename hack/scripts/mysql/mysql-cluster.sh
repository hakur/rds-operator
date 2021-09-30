#!/bin/bash

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

################################# cluster functions #################################################

function StartCluster() {
    local code=0

    if [ "$MYSQL_CLUSTER_MODE" == "SemiSync" ];then
        StartSemiSyncCluster
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
