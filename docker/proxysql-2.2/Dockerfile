FROM docker.io/proxysql/proxysql:2.2.0
ADD sources.list /etc/apt/sources.list
RUN apt update
RUN apt -y install mysql-client netcat vim procps