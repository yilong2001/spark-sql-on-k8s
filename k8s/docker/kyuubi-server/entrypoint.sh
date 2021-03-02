#!/bin/bash

# echo commands to the terminal output
set -ex

CLASS="org.apache.kyuubi.server.KyuubiServer"
RUNNER="${JAVA_HOME}/bin/java"

# KYUUBI_CONF_DIR should be set in deployment.yaml
# KYUUBI_JAVA_OPTS should be set in deployment.yaml

# Check whether there is a passwd entry for the container UID
myuid=$(id -u)
mygid=$(id -g)
# turn off -e for getent because it will return error code in anonymous uid case
set +e
uidentry=$(getent passwd $myuid)
set -e

echo $myuid
echo $mygid
echo $uidentry

# If there is no passwd entry for the container UID, attempt to create one
if [[ -z "$uidentry" ]] ; then
    if [[ -w /etc/passwd ]] ; then
        echo "$myuid:x:$myuid:$mygid:anonymous" >> /etc/passwd
    else
        echo "Container ENTRYPOINT failed to add passwd entry for anonymous UID"
    fi
fi

KYUUBI_CLASSPATH="${KYUUBI_JAR_DIR}/*:${SPARK_JAR_DIR}/*"
if ! [ -z ${KYUUBI_CONF_DIR} ]; then
  KYUUBI_CLASSPATH="${KYUUBI_CLASSPATH}:${KYUUBI_CONF_DIR}";
fi

cmd="${RUNNER} ${KYUUBI_JAVA_OPTS} -cp ${KYUUBI_CLASSPATH} $CLASS"

$cmd "$@"
#