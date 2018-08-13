# The builder will be used to generate the executable (named 'bootstrap')
# which upon invocation will create the necessary 'supervisor.conf' file
FROM golang:1.9.4 as builder

ENV GOPATH /go

#copy the source file
WORKDIR /generator/

# This is source code of the http server
COPY server.go .

# generate executable that will be used by the entrypoint shell script to generate
# the supervisor configuration
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o server server.go

# Copy the application into a thin image
FROM busybox

ARG GENERATOR_DIR=/opt/generator

RUN mkdir -p ${GENERATOR_DIR}/bin

COPY --from=builder /supervisord/server ${SUPERVISORD_DIR}/bin/

RUN chgrp -R 0 ${GENERATOR_DIR} && \
    chmod -R g+rwX ${GENERATOR_DIR} && \
    chmod 775 ${SUPERVISORD_DIR}/bin/server

WORKDIR ${GENERATOR_DIR}
ENTRYPOINT ["/opt/generator/bin/server"]
