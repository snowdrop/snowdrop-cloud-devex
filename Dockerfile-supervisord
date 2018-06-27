#FROM debian:latest
FROM centos:7
ADD echo.sh /tmp/echo.sh
ADD supervisor.conf /tmp/supervisor.conf
COPY --from=ochinchina/supervisord:latest /usr/local/bin/supervisord /usr/local/bin/supervisord
ENTRYPOINT ["/usr/local/bin/supervisord","-c","/tmp/supervisor.conf"]