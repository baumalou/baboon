FROM docker.workshop21.ch/boilerplate/build/rook-build:latest


ENV LANG en_US.UTF-8
ENV GOVERSION 1.9.2
ENV GOROOT /go
ENV GOPATH /go/src

RUN apt-get update -y && apt-get install wget git -y &&  \
    cd / && wget https://storage.googleapis.com/golang/go${GOVERSION}.linux-amd64.tar.gz && \
    tar zxf go${GOVERSION}.linux-amd64.tar.gz && rm go${GOVERSION}.linux-amd64.tar.gz && \
    ln -s /go/bin/go /usr/bin/ && \
    ls /go

RUN mkdir -p /app
ADD rookctl /usr/local/bin/rookctl
RUN chmod +x /usr/local/bin/rookctl
ADD .git-credentials /
ADD main /app/main
RUN chmod +x /main.sh
ADD go-wrapper /usr/local/bin/
RUN chmod +x /usr/local/bin/go-wrapper
RUN chmod -R +rwx /app
WORKDIR ["/app/"]
ENTRYPOINT [ "/app/main" ]

 