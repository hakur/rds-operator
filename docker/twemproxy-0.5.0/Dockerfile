FROM ubuntu:20.04 as builder
COPY sources.list /etc/apt/sources.list
RUN apt update
RUN apt install -y make gcc git autoconf 
WORKDIR /
RUN git clone https://github.com/twitter/twemproxy
WORKDIR /twemproxy
RUN git checkout 0.5.0
RUN apt install -y libtool
RUN autoreconf -fvi
RUN ls && ./configure --enable-debug=full --prefix=/target
RUN make -j $(nproc)
RUN make install

FROM ubuntu:20.04
COPY --from=builder /target /target
COPY cmd.sh /cmd.sh
RUN mv /target/sbin/* /sbin/ \
    && mv /target/lib/* /lib \
    && rm -rf /target \
    && chmod +x /cmd.sh
CMD /cmd.sh