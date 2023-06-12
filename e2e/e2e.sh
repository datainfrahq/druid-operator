#!/bin/bash
set -o errexit
set -x
# Get Kind
go install sigs.k8s.io/kind@v0.17.0
# minio statefulset name
MINIO_STS_NAME=myminio-ss-0
# druid namespace
NAMESPACE=druid
# fmt code
make fmt
# vet
make vet
# deploy kind
make kind
# build local docker druid operator image
make docker-build-local
# push to kind registry
make docker-push-local
# build local docker test image
make docker-build-local-test
# push to kind registry
make docker-push-local-test
# try to install the CRD with make
make install
# delete the crd
make uninstall
# install druid-operator
make helm-install-druid-operator
# install minio-operator and tenant
make helm-minio-install
# hack for minio pod to get started
sleep 60
# wait for minio pods
kubectl rollout status sts $MINIO_STS_NAME -n ${NAMESPACE}  --timeout=300s
# output pods
kubectl get pods -n ${NAMESPACE}
# apply druid cr
kubectl apply -f e2e/configs/druid-cr.yaml -n ${NAMESPACE}
# hack for druid pods
sleep 30
# wait for druid pods
declare -a sts=($( kubectl get sts -n ${NAMESPACE} -l app=${NAMESPACE} -o name| sort -r))
for s in ${sts[@]}; do
  echo $s
  kubectl rollout status $s -n ${NAMESPACE}  --timeout=300s
done

# Running test job with an example dataset 
make deploy-testjob

# Delete old druid
kubectl delete -f e2e/configs/druid-cr.yaml -n ${NAMESPACE}
sleep 30

# Start testing use-cases
# Test: `ExtraCommonConfig`
sed -e "s/NAMESPACE/${NAMESPACE}/g" e2e/configs/extra-common-config.yaml | kubectl apply -n ${NAMESPACE} -f -
# hack for druid pods
sleep 30
# wait for druid pods
declare -a sts=($( kubectl get sts -n ${NAMESPACE} -l app=${NAMESPACE} -l druid_cr=extra-common-config -o name| sort -r))
for s in ${sts[@]}; do
  echo $s
  kubectl rollout status $s -n ${NAMESPACE}  --timeout=300s
done

extraDataTXT=$(kubectl get configmap -n $NAMESPACE extra-common-config-druid-common-config -o 'jsonpath={.data.test\.txt}')
if [[ "${extraDataTXT}" != "This Is Test" ]]
then
  echo "Bad value for key: test.txt"
  echo "Test: ExtraCommonConfig => FAILED!"
fi

extraDataYAML=$(kubectl get configmap -n $NAMESPACE extra-common-config-druid-common-config -o 'jsonpath={.data.test\.yaml}')
if [[ "${extraDataYAML}" != "YAML" ]]
then
  echo "Bad value for key: test.yaml"
  echo "Test: ExtraCommonConfig => FAILED!"
fi

echo "Test: ExtraCommonConfig => SUCCESS!"
kind delete cluster
