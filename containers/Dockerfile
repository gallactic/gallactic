
FROM alpine:3.8

ENV WORKING_DIR /gallactic


RUN apk add --no-cache bash libgcc


ENV GOPATH /go
ENV PATH "$PATH:/go/bin"
RUN apk add --no-cache curl git go build-base rust cargo && \
    mkdir -p /go/src/github.com/gallactic/gallactic/ && \
    cd /go/src/github.com/gallactic/gallactic/ && \
    git clone https://github.com/gallactic/gallactic/ . && \
    git checkout develop && \
    make tools deps build && \
    cp ./build/gallactic /usr/bin && \
    apk del curl git go build-base rust cargo && \
    rm -rf /go /root/.cache /root/.cargo

EXPOSE 46656
EXPOSE 50051
EXPOSE 1337

VOLUME $WORKING_DIR
WORKDIR $WORKING_DIR
ENTRYPOINT ["gallactic"]
