#!/bin/bash
###
curdir=`pwd`

echo "build sparkappctrl start ..."
cd ./sparkop-ctrl/sparkappctrl
go build
mv ./sparkappctrl ../../bin/
cd $curdir
echo "build sparkappctrl over"

echo "build traefikkit start ..."
cd ./sparkop-ctrl/traefikkit
go build
mv ./traefikkit ../../bin/
cd $curdir
echo "build traefikkit over"

# build sparkop-ctrl images
# ./img-linux-amd64 build -t sparkop-ctrl:v1.0 -f k8s/docker/sparkop-ctrl/Dockerfile .

