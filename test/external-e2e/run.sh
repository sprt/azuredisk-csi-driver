#!/bin/bash

# Copyright 2021 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Ensure that we have the desired version of the ginkgo test runner.

set -xe

PROJECT_ROOT=$(git rev-parse --show-toplevel)
DRIVER="test"

install_ginkgo () {
    apt update -y
    apt install -y golang-ginkgo-dev
}

setup_e2e_binaries() {
    mkdir /tmp/csi-azuredisk

    # download k8s external e2e binary for kubernetes
    curl -sL https://storage.googleapis.com/kubernetes-release/release/v1.22.0/kubernetes-test-linux-amd64.tar.gz --output e2e-tests.tar.gz
    tar -xvf e2e-tests.tar.gz && rm e2e-tests.tar.gz

    # enable fsGroupPolicy (only available from k8s 1.20)
    export EXTRA_HELM_OPTIONS="--set feature.enableFSGroupPolicy=true --set image.csiProvisioner.tag=v3.0.0 --set snapshot.apiVersion=ga"
    # test on alternative driver name
    EXTRA_HELM_OPTIONS=$EXTRA_HELM_OPTIONS" --set driver.name=$DRIVER.csi.azure.com --set controller.name=csi-$DRIVER-controller --set linux.dsName=csi-$DRIVER-node --set windows.dsName=csi-$DRIVER-node-win"
    # install the azuredisk-csi-driver driver
    make e2e-bootstrap
    sed -i "s/csi-azuredisk-controller/csi-$DRIVER-controller/g" deploy/example/metrics/csi-azuredisk-controller-svc.yaml
    make create-metrics-svc
}

print_logs() {
    sed -i "s/disk.csi.azure.com/$DRIVER.csi.azure.com/g" deploy/example/storageclass-azuredisk-csi.yaml
    bash ./hack/verify-examples.sh linux azurepubliccloud ephemeral $DRIVER
    echo "print out driver logs ..."
    bash ./test/utils/azuredisk_log.sh $DRIVER
}


install_ginkgo
setup_e2e_binaries
trap print_logs EXIT

ginkgo -p --progress --v -focus="External.Storage.*$DRIVER.csi.azure.com" \
       -skip='\[Disruptive\]|\[Slow\]|should check snapshot fields, check restore correctly works after modifying source data, check deletion|should resize volume when PVC is edited while pod is using it' kubernetes/test/bin/e2e.test -- \
       -storage.testdriver=$PROJECT_ROOT/test/external-e2e/manifest/testdriver.yaml \
       --kubeconfig=$KUBECONFIG
