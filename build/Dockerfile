FROM golang AS gobuilder

WORKDIR /opt/app
RUN go env -w GOPROXY="https://goproxy.cn,direct"
COPY go.mod go.sum /opt/app/
RUN go mod download && go mod verify
COPY cmd /opt/app/cmd/
RUN go build -v -o startServer cmd/*



FROM centos:7

ENV PATH=/opt/jdk/bin:/usr/local/go/bin:$PATH
RUN yum install -y vim openssl && yum clean all && rm -rf /var/cache/yum
RUN useradd developer
RUN mkdir -p /opt/server/conf /opt/jdk /opt/zookeeper /mnt/data/{zookeeper,kafka} /mnt/logs/{zookeeper,kafka} && chown -R developer:developer /opt/server /opt/jdk /opt/zookeeper /mnt/data /mnt/logs
COPY --chown=developer:developer jdk1.8.0_341 /opt/jdk
ADD --chown=developer:developer apache-zookeeper-3.6.3-bin /opt/zookeeper
ADD --chown=developer:developer kafka_2.12-2.8.1 /opt/kafka
COPY --chown=developer:developer conf/ /opt/server/conf/
COPY --chown=developer:developer --chmod=755 start.sh /opt/server/
USER developer
RUN cp /opt/zookeeper/conf/zoo_sample.cfg /opt/zookeeper/conf/zoo.cfg
COPY --from=gobuilder --chown=developer:developer /opt/app/startServer /opt/server
WORKDIR /opt/server
# CMD ["/bin/bash", "-c", "set -e && /opt/server/startServer"]
# CMD ["/opt/server/startServer"]
CMD ["/opt/server/start.sh"]

