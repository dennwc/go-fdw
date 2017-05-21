FROM ubuntu:zesty

ENV PG_VERS=9.6
ENV GO_VERS=1.8.1

RUN apt update && \
    apt install -y postgresql-${PG_VERS} postgresql-server-dev-${PG_VERS} libpq-dev wget build-essential && \
    wget -q https://godeb.s3.amazonaws.com/godeb-amd64.tar.gz && \
    tar -xzf godeb-amd64.tar.gz && \
    ./godeb install ${GO_VERS} && \
    rm godeb* && \
    apt clean

WORKDIR /build/

ADD go_fdw* ./
ADD Makefile ./

VOLUME /gopath
ENV GOPATH=/gopath

VOLUME /out

ADD fdw.go ./

CMD sh -c 'make clean && make go && make && make install && cp go_fdw.so go_fdw.control go_fdw--1.0.sql /out'