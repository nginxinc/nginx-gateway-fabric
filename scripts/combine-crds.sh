#!/bin/bash

mkdir -p build/out

CRD_FILE=build/out/crds.yaml
echo "# NGINX Gateway API CustomResourceDefinitions" > ${CRD_FILE}

for file in `ls deploy/manifests/crds/*.yaml`; do
    echo "#" >> ${CRD_FILE}
    echo "# $file" >> ${CRD_FILE}
    echo "#" >> ${CRD_FILE}
    cat $file >> ${CRD_FILE}
done
