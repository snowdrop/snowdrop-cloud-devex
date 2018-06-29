# Instructions to install API using the installer
# Build and test the controller-manager
FROM golang:1.9.4 AS builder

# Copy in the go src
WORKDIR /go/src/github.com/cmoulliard/k8s-supervisor

# Build and test the API code
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o app ./app.go

# Copy the application into a thin image
FROM busybox

ENV SUPERVISORD_DIR /opt/supervisord

RUN mkdir -p ${SUPERVISORD_DIR}/conf ${SUPERVISORD_DIR}/bin

ADD conf/echo.sh ${SUPERVISORD_DIR}/conf/
ADD https://github.com/ochinchina/supervisord/releases/download/v0.5/supervisord_0.5_linux_amd64 ${SUPERVISORD_DIR}/bin/supervisord

RUN chgrp -R 0 ${SUPERVISORD_DIR} && \
    chmod -R g+rwX ${SUPERVISORD_DIR} && \
    chmod -R 666 ${SUPERVISORD_DIR}/conf/* && \
    chmod 775 ${SUPERVISORD_DIR}/bin/supervisord && \
    chmod 775 ${SUPERVISORD_DIR}/conf/echo.sh

WORKDIR /root/
COPY --from=builder /go/src/github.com/cmoulliard/k8s-supervisor/app .
ENTRYPOINT ["./app"]
CMD ["--install-crds=false"]
