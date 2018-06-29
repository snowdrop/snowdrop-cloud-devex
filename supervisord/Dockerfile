# Instructions to build the code boostrapping the supervisord and responsible to populate the supervisord.conf file
FROM golang:1.9.4 as builder

ENV GOPATH /go

#copy the source file
WORKDIR /supervisord/
COPY conf/ conf/
COPY main.go .

#compile the project to generate bootstrap
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bootstrap main.go

# Copy the application into a thin image
FROM busybox

ARG VERSION=0.5

ENV SUPERVISORD_DIR /opt/supervisord

RUN mkdir -p ${SUPERVISORD_DIR}/conf ${SUPERVISORD_DIR}/bin

COPY --from=builder /supervisord/bootstrap ${SUPERVISORD_DIR}/bin/
COPY --from=builder /supervisord/conf/ ${SUPERVISORD_DIR}/

RUN echo "VERSION :" ${VERSION}

#add the go supervisord application
ADD https://github.com/ochinchina/supervisord/releases/download/v${VERSION}/supervisord_${VERSION}_linux_amd64 ${SUPERVISORD_DIR}/bin/supervisord

RUN chgrp -R 0 ${SUPERVISORD_DIR} && \
    chmod -R g+rwX ${SUPERVISORD_DIR} && \
    chmod -R 666 ${SUPERVISORD_DIR}/conf/* && \
    chmod 775 ${SUPERVISORD_DIR}/bin/supervisord && \
    chmod 775 ${SUPERVISORD_DIR}/bin/bootstrap && \
    chmod 775 ${SUPERVISORD_DIR}/conf/echo.sh

WORKDIR $SUPERVISORD_DIR
ENTRYPOINT ["/opt/supervisord/bin/bootstrap"]
