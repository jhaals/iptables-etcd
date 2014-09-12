# Ubuntu should be replaced by a smaller image
FROM ubuntu
MAINTAINER Johan Haals <johan.haals@gmail.com>

RUN apt-get update
RUN apt-get install -y golang

ADD . /source

RUN cd /source && go build -o /bin/iptables-etcd

CMD ["iptables-etcd"]