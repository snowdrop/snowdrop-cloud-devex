#FROM registry.access.redhat.com/redhat-openjdk-18/openjdk18-openshift
FROM fabric8/s2i-java:latest
ENV JAVA_APP_DIR=/deployments
EXPOSE 8080
COPY target /deployments/