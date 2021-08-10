FROM docker.io/proxysql/proxysql:2.2.0
ADD sources.list /etc/apt/sources.list
RUN apt update
RUN apt -y install mysql-client netcat vim procps
COPY mgrsp_switch.sh /mgrsp_switch.sh
RUN chmod +x /mgrsp_switch.sh 