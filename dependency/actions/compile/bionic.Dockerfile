FROM ubuntu:18.04

RUN apt-get -y update
RUN apt-get -y install build-essential curl zlib1g zlib1g-dev libssl-dev libpcre3 libpcre3-dev python3

COPY entrypoint /entrypoint

ENTRYPOINT ["/entrypoint"]
