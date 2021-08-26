notice: mysql mgr mode at least need three nodes, because it use paxos protocol for mysql replication log consistence，log replay on slave nodes is async

注意：mysql mgr 模式至少需要三个节点，因为它使用paxos协议来保证mysql复制日志的一致性，日志在slave节点上的回放过程是异步的

### Role setting 角色设置
* mysql-1 candidate node （候选人节点）
* mysql-2 candidate node （候选人节点）
* mysql-3 candidate node （候选人节点）
* mysql-n candidate node ... （候选人节点）

### Run process
* #### how mysql nodes found cluster peers
    there is a mysql system variables group_replication_group_seeds, this defines mysql paxos peers

    mysql有一个系统变量 group_replication_group_seeds 定义了mysql paxos节点

* #### cluster run 集群运行
    when mysql-0 is running，it will try to find is there has master node in cluster.

    if there is no master node in cluster, then mysql-0 boot it self as master node，mysql-1 and mysql-2 join mysql-0 as slave node when them start running.

    if there is a master node in cluster, mysql-1 or mysql-2 must is master node, for avoid cluster data consistence, mysql-0 join cluster as slave node.

    当mysql-0运行后，它将尝试查找集群内是否有master节点。

    如果集群内没有master节点，mysql-0将自举为master节点，mysql-1和mysql-2在启动后将以slave身份加入mysql-0.

    如果集群内存在master节点，那master节点必将是mysql-1或mysql-2中的其中一个节点，为了保证集群数据一致性，mysql-0将以slave身份加入cluster

* ### master down （master宕机）
    when mysql-0 down, mysql-1 will be master node, when mysql-1 down, mysql-2 will be master node, when mysql-2 down , mysql-0 will be master

    当 mysql-0宕机，mysql-1将会成为master节点，当mysql-1宕机，mysql-2将成为master节点，当mysql-2宕机，mysql-0将成为master节点