FROM registry.access.redhat.com/redhat-openjdk-18/openjdk18-openshift
#FROM fabric8/java-alpine-openjdk8-jre:latest
ENV JAVA_APP_DIR=/deployments
EXPOSE 8080
COPY target /deployments/
