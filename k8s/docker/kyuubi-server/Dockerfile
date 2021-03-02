#
# Copyright 2017 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

ARG java_image_tag=8-jre-slim

FROM openjdk:${java_image_tag}

RUN apt-get update \
    && apt-get install -y openssl curl tini \
    && rm -rf /var/lib/apt/lists/* \
	&& mkdir -p /opt/kyuubi/jars && mkdir -p /opt/spark-3.0.1-hadoop2/jars
#RUN  mkdir -p /opt/kyuubi/jars && mkdir -p /opt/spark-3.0.1-hadoop2/jars

#COPY dist/bin/tini-v0.19.0     /opt/tini
COPY external/jars/kyuubi/*     /opt/kyuubi/jars/
COPY external/jars/spark/*     /opt/spark-3.0.1-hadoop2/jars/
COPY k8s/docker/kyuubi-server/entrypoint.sh       /opt/entrypoint.sh

# KYUUBI_CONF_DIR should be set in deployment.yaml, and so to use configmap as conf dir
# KYUUBI_JAVA_OPTS should be set in deployment.yaml

# RUN chmod -R 755 /opt/kyuubi

ENV KYUUBI_JAR_DIR  /opt/kyuubi/jars
ENV SPARK_JAR_DIR   /opt/spark-3.0.1-hadoop2/jars
# ENV KYUUBI_HOME   /opt/kyuubi

EXPOSE 10019
#ENTRYPOINT [ "/opt/entrypoint.sh" ]
ENTRYPOINT ["/usr/bin/tini", "--", "/opt/entrypoint.sh"]