FROM ubuntu:20.04 as builder
COPY sources.list /etc/apt/sources.list
ADD redis-cluster-proxy-1.0-beta2.tar.gz /build
RUN apt update 
RUN apt install -y gcc make
RUN cd /build/redis-cluster-proxy-1.0-beta2 && make -j $(nproc)
RUN cd /build/redis-cluster-proxy-1.0-beta2 && make PREFIX=/target install 

FROM ubuntu:20.04
COPY sources.list /etc/apt/sources.list
COPY --from=builder /target/bin/redis-cluster-proxy /bin/redis-cluster-proxy
CMD redis-cluster-proxy -a ${REDIS_PASSWORD} -p 6379 --bind 0.0.0.0 ${REDIS_NODES} 