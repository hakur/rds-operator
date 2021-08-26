### Role setting 角色设置
* master: mysql-0 mysql-1
* slaves: mysql-2 mysql-3 mysql-n ...

### Run process 运行流程
* #### master servers join each other as master 互为主从
    cluster use gtid replication mode.
    
    mysql-0 boot as master node, then join mysql-1 as slave node. 

    mysql-1 boot as slave node, then join mysql-1 as slave node.

    mysql-0 and mysql-1 copy data from each other, so cluster data mostly lossless

    集群使用gtid复制模式。

    mysql-0以master节点身份启动，然后以slave身份加入mysql-1。

    mysql-1以master节点身份启动，然后以slave身份加入mysql-0。

    mysql-0和mysql-1互相复制对方的数据，因此集群数据大多数时候都是无损的。


* #### cluster run 集群启动
    after mysql-0 and mysql-1 boot as master node and join each other as slave node.

    mysql-2 join master-1 as slave node, mysql-3 join master-1 as slave node.

    当mysql-0和mysql-1都以master身份启动并互相以slave身份加入对方形成互为主从之后。

    mysql-2以slave身份加入master-1，mysql-3以slave身份加入mysql-1

* #### master down 主节点宕机
    when mysql-0 down, all slave nodes use mysql-1 as master node for replication.

    when mysql-1 down, all slave nodes use mysql-0 as master node for replication, after mysql-1 respawn, all slaves will swtich master to mysql-1.

    当mysql-0宕机后，所有的slave节点使用mysql-1作为master节点以复制数据。

    当mysql-1宕机后，所有的slave节点将使用mysql-0作为master节点以复制数据，在mysql-1复活之后，所有的slave节点将master切换至mysql-1。

* ### mysql-router routing （mysql-router路由）
    mysql clients connect mysql-router as single entry mysql server, mysql-router is a read/write split middleware.

    mysql客户端连接mysql-router作为mysql服务器单点入口，mysql-router是一个读写分离中间件。