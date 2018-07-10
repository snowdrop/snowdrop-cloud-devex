FROM registry.access.redhat.com/redhat-openjdk-18/openjdk18-openshift

ENV JAVA_APP_DIR=/deployments
ENV JAVA_APP_JAR=*-exec.jar

EXPOSE 8080

USER root
RUN mkdir -p /tmp/src/target

RUN chgrp -R 0 /tmp/src/ && \
    chmod -R g+rw /tmp/src/