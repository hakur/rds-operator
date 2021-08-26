notice: redis cluster mode at least need three master nodethee slave node, total six nodes.

注意：redis cluster模式至少需要三个master节点，三个slave节点，总共六个节点

### Role setting 角色设置
* redis-0 candidate (候选人节点)
* redis-1 candidate (候选人节点)
* redis-2 candidate (候选人节点)
* redis-3 candidate (候选人节点)
* redis-4 candidate (候选人节点)
* redis-5 candidate (候选人节点)
* redis-3n... candidate (候选人节点)

### predixy
* #### vs redislabs redis cluster proxy
    current redislabs product redis cluster proxy version is v1.0-beta2, it dose not suppor client must auth function, it has a password config option but only for automatically auth backend redis server nodes, so any redis client can connect and do set/get on redis cluster proxy without password, it's a nude proxy middleware. it's a strange design.

    so I switch to predixy, it support redis cluster mode and other mode, such as sentinel cluster mode. it also support client must auth mode, and read/write key space acl control.

    当前redislabs的产品redis cluster proxy 版本是v1.0-beta2，它不支持客户端必须认证这个功能，它有一个password配置项但只是为了自动认证后端的redis服务器节点，所以任何redis客户端都不用需要密码就能在redis cluster proxy上执行set/get操作，它就是个裸奔的代理中间件，真是个奇怪的设计。

    所以我切换到了predixy，它支持redis cluster模式及其他模式，比如哨兵集群模式。它也支持客户端必须认证这个功能，及基于key空间的read/write权限控制。