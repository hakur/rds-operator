### this repo is also an kubebuilder project exmple

### brefore use release yaml of mysql. must modify storage class name and whitelist field

### crd operator contains below resources
* mysql.rds.hakurei.cn/v1alpha1
    * mysql version
        - [ ] 5.7.34
            - [x] prometheus operator pod monitor
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
    * - [x] prometheus operator service monitor
    * redis version
        - [x] 6.2.5
            - [x] redis cluster with predixy

* proxysql.rds.hakurei.cn/v1alpha1
    * version: 2.x , current 2.2.x 2.3.x supported
    * limits
        * not support mysql cluster mode changed, such as from semi sync replication to group replication

* mongo.rds.hakurei.cn/v1alpha1 (in plan)
        
* mutating adminssion webhook (in plan)

### develop
use [kt-connect](https://github.com/alibaba/kt-connect) for development 

first of all:
```sh
# install crd defines
make gen -j $(nproc)
make install
```

wsl2:
```sh
    ktctl connect --method=sock5 # open terminal window, notice : when pod created or deleted, must restart ktctl, otherwise you will see many context exceeded
    make dev -j $(nproc) # open new terminal window
    make release BRANCH=v0.0.1 PUSH=true -j $(nproc) # example make release 
```

skaffold (recommend with local kubernetes):
```sh
    make skaffold -j $(nproc)
```
