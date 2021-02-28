#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
#FROM busybox:latest 
FROM debian:buster-slim

ARG spark_uid=185

# docker build commond
# docker build -t sparkop:1.0 -f k8s/docker/Dockerfile .

# Add Tini
RUN mkdir -p /opt/sparkop && \
    mkdir -p /opt/bin && \
    mkdir -p /opt/jars

RUN apt-get update \
    && apt-get install -y wget \
    && rm -rf /var/lib/apt/lists/*

RUN wget https://github.com/nats-io/nats-streaming-server/releases/download/v0.20.0/nats-streaming-server-v0.20.0-linux-amd64.tar.gz https://github.com/krallin/tini/releases/download/v0.19.0/tini-amd64 \
 && tar -xzf nats-streaming-server-v0.20.0-linux-amd64.tar.gz \
 && mv nats-streaming-server-v0.20.0-linux-amd64/nats-streaming-server /opt/bin \
 && mv tini-amd64 /opt/bin \
 && rm -rf nats-streaming-server-v0.20.0-linux-amd64 \
 && rm -r nats-streaming-server-v0.20.0-linux-amd64.tar.gz \
 && chmod 777 /opt/bin/nats-streaming-server \
 && chmod 777 /opt/bin/tini-amd64

COPY bin/sparkappctrl /opt/sparkop/sparkappctrl
COPY bin/traefikkit /opt/sparkop/traefikkit
COPY external/jars/*  /opt/jars/

#RUN chmod 777 /opt/bin/nats-streaming-server  && chmod 777 /opt/bin/tini-amd64 && chmod 777 /opt/sparkop/sparkappctrl && chmod 777 /opt/sparkop/traefikkit

ENV SPARK_SQL_ENGINE_JAR /opt/jars/kyuubi-spark-sql-engine-1.1.0-SNAPSHOT.jar
ENV SPARK_SQL_DRIVER_JAR_MYSQL /opt/jars/mysql-connector-java-5.1.48.jar
ENV SPARK_SQL_DRIVER_JAR_PG /opt/jars/postgresql-42.2.18.jar
#ENV SPARK_SQL_IMAGE_NAME gcr.io/spark-operator/spark:v3.0.1-1216

#WORKDIR /opt/sparkop
#RUN chmod 777 /opt/bin/tini && chmod 777 /opt/sparkop/sparkappctrl && chmod 777 /opt/sparkop/traefikkit

#RUN useradd -u ${spark_uid} -ms /bin/bash spark

# Specify the User that the actual main process will run as
#USER ${spark_uid}

EXPOSE 8085
ENTRYPOINT ["/opt/bin/tini-amd64", "--"]

