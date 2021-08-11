FROM docker.io/proxysql/proxysql:2.2.0
ADD sources.list /etc/apt/sources.list
RUN apt update
RUN apt -y install mysql-client netcat vim procps
COPY mgrsp_switch.sh /mgrsp_switch.sh
COPY entrypoint.sh /entrypoint.sh
COPY load_servers.sh /load_servers.sh
RUN chmod +x /mgrsp_switch.sh  && chmod +x /entrypoint.sh && chmod + /load_servers.sh
ENTRYPOINT ["/entrypoint.sh"]