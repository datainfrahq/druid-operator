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
sleep 10
for m in $(kubectl get pod -n minio-operator -o name)
do
  kubectl wait -n minio-operator "$m" --for=condition=Ready --timeout=5m
done
sleep 10
kubectl wait pod -n ${NAMESPACE} -l app=minio --for=condition=Ready --timeout=5m
# wait for minio pods
kubectl rollout status sts $MINIO_STS_NAME -n ${NAMESPACE} --timeout=5m
# output pods
kubectl get pods -n ${NAMESPACE}
# apply druid cr
kubectl apply -f e2e/configs/druid-cr.yaml -n ${NAMESPACE}
sleep 10
for d in $(kubectl get pods -n ${NAMESPACE} -l app=druid -l druid_cr=tiny-cluster -o name)
do
  kubectl wait -n ${NAMESPACE} "$d" --for=condition=Ready --timeout=5m
done
# wait for druid pods
for s in $(kubectl get sts -n ${NAMESPACE} -l app=${NAMESPACE} -l druid_cr=tiny-cluster -o name)
do
  kubectl rollout status "$s" -n ${NAMESPACE}  --timeout=5m
done

# Running test job with an example dataset 
make deploy-testjob

# Delete old druid
kubectl delete -f e2e/configs/druid-cr.yaml -n ${NAMESPACE}
for d in $(kubectl get pods -n ${NAMESPACE} -l app=druid -l druid_cr=tiny-cluster -o name)
do
  kubectl wait -n ${NAMESPACE} "$d" --for=delete --timeout=5m
done

# Start testing use-cases
# Test: `ExtraCommonConfig`
sed -e "s/NAMESPACE/${NAMESPACE}/g" e2e/configs/extra-common-config.yaml | kubectl apply -n ${NAMESPACE} -f -
sleep 10
# Wait for ZooKeeper
for z in $(kubectl get pods -n ${NAMESPACE} -l zk_cluster=extra-common-config-zk -o name)
do
  kubectl wait -n ${NAMESPACE} "$z" --for=condition=Ready --timeout=5m
done
# Wait for Druid
for d in $(kubectl get pods -n ${NAMESPACE} -l app=druid -l druid_cr=extra-common-config -o name)
do
  kubectl wait -n ${NAMESPACE} "$d" --for=condition=Ready --timeout=5m
done
# wait for druid pods
for s in $(kubectl get sts -n ${NAMESPACE} -l app=${NAMESPACE} -l druid_cr=extra-common-config -o name)
do
  kubectl rollout status "$s" -n ${NAMESPACE}  --timeout=5m
done

extraDataTXT=$(kubectl get configmap -n $NAMESPACE extra-common-config-druid-common-config -o 'jsonpath={.data.test\.txt}')
if [[ "${extraDataTXT}" != "This Is Test" ]]
then
  echo "Bad value for key: test.txt"
  echo "Test: ExtraCommonConfig => FAILED\!"
fi

extraDataYAML=$(kubectl get configmap -n $NAMESPACE extra-common-config-druid-common-config -o 'jsonpath={.data.test\.yaml}')
if [[ "${extraDataYAML}" != "YAML" ]]
then
  echo "Bad value for key: test.yaml"
  echo "Test: ExtraCommonConfig => FAILED\!"
fi

echo "Test: ExtraCommonConfig => SUCCESS\!"
kind delete cluster

