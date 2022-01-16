### crd operator contains below resources
* mysql.rds.hakurei.cn/v1alpha1
    * mysql version
        - [ ] 5.7.34
            - [x] MGR single primary
            - [ ] MGR multi primary
            - [x] Semi sync replication
        - [ ] 8.0
            - [ ] MGR single primary
            - [ ] MGR multi primary
            - [ ] Semi sync replication
* mysqlbackup.rds.hakurei.cn/v1alpha1
    - [x] logical backup dump sql to s3 server
    - [ ] physical backup

* redis.rds.hakurei.cn/v1alpha1
    * redis version
        - [x] 6.2.5
            - [x] redis cluster with predixy

* proxysql.rds.hakurei.cn/v1alpha1
    * version: 2.x , current 2.2.x 2.3.x supported
    * limits
        * not support mysql cluster mode changed, such as from semi sync replication to group replication

* mongo.rds.hakurei.cn/v1alpha1 (in plan)
        
* mutating adminssion webhook (in plan)

* prometheus operator service monitor (in plan)

### develop
use [kt-connect](https://github.com/alibaba/kt-connect) for development 

wsl2:
```
    ktctl connect --method=sock5 # open terminal window, notice : when pod created or deleted, must restart ktctl, otherwise you will see many context exceeded
    make dev # open new terminal window
    make release BRANCH=v0.0.1 PUSH=true # example make release 
```