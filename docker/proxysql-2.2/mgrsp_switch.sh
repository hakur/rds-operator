#!/bin/bash
# 注意如果proxysql的后端是sqlite3，那么是不支持事务的，因此这个脚本里面都是不带事务的happy pass
READ_GROUP_ID=2
WRITE_GROUP_ID=1
MYSQL_PORT=3306
PROXYSQL_ADMIN_PORT=6032
PROXYSQL_ADMIN_USER=$1
PROXYSQL_ADMIN_PASSWORD=$2
MYSQL_NDOE_MAX_CONN=${MYSQL_MAX_CONN:-300}
MYSQL_READ_HOSTS=$3

function echolog() {
  echo "msrsp_switch.sh：[$(date)] $1"
}

function proxysqlCmd(){
  mysql -u${PROXYSQL_ADMIN_USER} -p${PROXYSQL_ADMIN_PASSWORD} -P${PROXYSQL_ADMIN_PORT} -h127.0.0.1 -N -se "$1"
}

function moveServerToWriteGroup() {
  server=$1
  proxysqlCmd "delete from mysql_servers where hostgroup_id=${WRITE_GROUP_ID};"
  proxysqlCmd "delete from mysql_servers where hostgroup_id=${READ_GROUP_ID} and hostname='${server}';"
  if [ $? != 0 ];then
    echo "false"
  fi

  proxysqlCmd "insert into mysql_servers (hostgroup_id, hostname, port, max_connections) values(${WRITE_GROUP_ID}, '${server}', ${MYSQL_PORT}, ${MYSQL_NDOE_MAX_CONN});"
  if [ $? != 0 ];then
    echo "false"
  fi

  echo "true"
}

function moveServerToReadGroup() {
  server=$1
  proxysqlCmd "delete from mysql_servers where hostgroup_id=${WRITE_GROUP_ID} and hostname='${server}';"
  if [ $? != 0 ];then
    echo "false"
  fi
  proxysqlCmd "insert into mysql_servers (hostgroup_id, hostname, port, max_connections) values(${READ_GROUP_ID}, '${server}', ${MYSQL_PORT}, ${MYSQL_NDOE_MAX_CONN});"
  if [ $? != 0 ];then
    echo "false"
  fi

  echo "true"
}

function findBestPrimaryServer() {
  # local readServers=$(proxysqlCmd "select hostname from mysql_servers where hostgroup_id=${READ_GROUP_ID} and status='ONLINE';")
  local readServers=$MYSQL_READ_HOSTS
  for server in $readServers;do
    if [ "$(checkWriteNodeIsOk $server)" == "true" ];then
      local serverUUID=$(mysql -u$MYSQL_USER -p$MYSQL_PASSWORD -P$MYSQL_PORT -h$server -N -se "show variables like 'server_uuid';" | awk '{print $2}')
      local primaryServerUUID=$(mysql -u$MYSQL_USER -p$MYSQL_PASSWORD -P$MYSQL_PORT -h$server -N -se "SELECT VARIABLE_VALUE FROM performance_schema.global_status WHERE VARIABLE_NAME= 'group_replication_primary_member' AND VARIABLE_VALUE='${serverUUID}';")
      if [ "$primaryServerUUID" == "$serverUUID" ] && [ "$serverUUID" != "" ];then
        echo $server
      fi
    fi
  done
}

function getCurrentWriteNodeHost() {
  local writeNodeAddress=$(proxysqlCmd "select hostname from mysql_servers where hostgroup_id=${WRITE_GROUP_ID} and status='ONLINE';")
  local writeNodeAddress=$(echo $writeNodeAddress | awk '{print $1}')
  echo $writeNodeAddress
}

function checkWriteNodeIsOk() {
  local mysqlHost=$1

  nc -z $mysqlHost $MYSQL_PORT
  if [ $? == 0 ];then
    local serverUUID=$(mysql -u$MYSQL_USER -p$MYSQL_PASSWORD -P$MYSQL_PORT -h$mysqlHost -N -se "show variables like 'server_uuid';" | awk '{print $2}')

    local primaryServerUUID=$(mysql -u$MYSQL_USER -p$MYSQL_PASSWORD -P$MYSQL_PORT -h$mysqlHost -N -se "SELECT VARIABLE_VALUE FROM performance_schema.global_status WHERE VARIABLE_NAME= 'group_replication_primary_member' AND VARIABLE_VALUE='${serverUUID}';")

    if [ "$primaryServerUUID" == "$serverUUID" ] && [ "$primaryServerUUID" != "" ];then
      echo "true"
    else 
      echo "false"
    fi
  else 
    echo "false"
  fi
}

mysqlCredentials=$(proxysqlCmd "SELECT variable_value FROM global_variables WHERE variable_name IN ('mysql-monitor_username','mysql-monitor_password') ORDER BY variable_name DESC;")
MYSQL_USER=$(echo $mysqlCredentials|awk '{print $1}')
MYSQL_PASSWORD=$(echo $mysqlCredentials|awk '{print $2}')

function main() {
  # 从proxysql中搜寻当前的写组记录
  local writeNodeHost=$(getCurrentWriteNodeHost)
  # 从读组中找寻当前的boostrap node
  local currentPrimary=$(findBestPrimaryServer)
  if [ "$(checkWriteNodeIsOk $writeNodeHost)" == "false" ];then
    if [ "$currentPrimary" == "" ];then
      echolog "give up switch master, current master not found !!!"
      return 1
    fi
    
    # 移除异常的写入节点
    
    if [ "$(moveServerToReadGroup $writeNodeHost)" == "true" ];then
      proxysqlCmd "LOAD MYSQL SERVERS TO RUNTIME; SAVE MYSQL SERVERS TO DISK;"
      if [ $? != 0 ];then
        echolog "move old write master [$writeNodeHost] to read group error: proxysql flush runtime_mysql_servers and save to disk failed"
      else 
        echolog "move old write master [$writeNodeHost] to read group successfully"
      fi
      
      if [ "$(moveServerToWriteGroup $currentPrimary)" == "true" ];then
        proxysqlCmd "LOAD MYSQL SERVERS TO RUNTIME; SAVE MYSQL SERVERS TO DISK;"
        if [ $? != 0 ];then
            echolog "move current writer master [$currentPrimary] to write group failed: flush runtime_mysql_servers and save to disk failed"
        fi
        echolog "move current write master [$currentPrimary] to write group successfully"
      else
        echolog "move current write master [$currentPrimary] to write group failed"
        return 3
      fi
    else
      echolog "move old writer master [$writeNodeHost] to read group failed"
      return 4
    fi
  fi
}

main