#!/bin/bash
###
curdir=`pwd`

IMG_BUILDER=/mnt/hgfs/linuxtools/img-builder/img-linux-amd64

######################################
echo "build kyuubi start ..."
cd ./kyuubi
#mvn clean install -DskipTests

cp -f ./kyuubi-main/target/kyuubi-main-1.1.0-SNAPSHOT.jar ${curdir}/external/jars/kyuubi/
cp -f ./kyuubi-common/target/kyuubi-common-1.1.0-SNAPSHOT.jar ${curdir}/external/jars/kyuubi/
cp -f ./externals/kyuubi-spark-sql-engine/target/kyuubi-spark-sql-engine-1.1.0-SNAPSHOT.jar ${curdir}/external/jars/kyuubi/

cp -f ./externals/kyuubi-spark-sql-engine/target/kyuubi-spark-sql-engine-1.1.0-SNAPSHOT.jar ${curdir}/external/jars/sql-engine/

cp -f ./externals/kyuubi-download/target/spark-3.0.1-bin-hadoop3.2/jars/* ${curdir}/external/jars/spark/

cd $curdir
echo "build kyuubi over"

######################################
echo "build spark sql dialect start ..."

cd ./spark-sql-dialect

#mvn clean install -DskipTests
cp -f ./target/original-core-1.0-SNAPSHOT.jar   ${curdir}/external/jars/kyuubi/

cd $curdir
echo "build spark sql dialect over"

######################################

echo "build sparkappctrl start ..."
cd ./sparkop-ctrl/sparkappctrl
go build
mv ./sparkappctrl ../../bin/
cd $curdir
echo "build sparkappctrl over"

######################################

echo "build traefikkit start ..."
cd ./sparkop-ctrl/traefikkit
go build
mv ./traefikkit ../../bin/
cd $curdir
echo "build traefikkit over"

######################################

# build sparkop-ctrl images
# ./img-linux-amd64 build -t sparkop-ctrl:v1.0 -f k8s/docker/sparkop-ctrl/Dockerfile .
echo "build sparkop-ctrl image start ..."
${IMG_BUILDER} build -t sparkop-ctrl:v1.0 -f k8s/docker/sparkop-ctrl/Dockerfile .
echo "build sparkop-ctrl image over"

######################################

echo "build kyuubi-server image start ..."
${IMG_BUILDER} build -t kyuubi-server:v1.0 -f k8s/docker/kyuubi-server/Dockerfile .
echo "build kyuubi-server image over"


