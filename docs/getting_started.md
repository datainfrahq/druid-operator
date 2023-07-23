# Installation

The operator installation is available as a [Helm chart](https://operatorhub.io/operator/druid-operator).  

The operator can be deployed in one of the following modes:
- namespace scope (default)
- cluster scope

### Cluster Scope Installation 
```bash
# Install Druid operator using Helm
helm -n druid-operator-system upgrade -i --create-namespace cluster-druid-operator ./chart

# ... or generate manifest.yaml to install using other means:
helm -n druid-operator-system template --create-namespace cluster-druid-operator ./chart > manifest.yaml
```

### Custom Namespaces Installation
```bash
# Install Druid operator using Helm
kubectl create ns mynamespace
helm -n druid-operator-system upgrade -i --create-namespace --set env.WATCH_NAMESPACE="mynamespace" namespaced-druid-operator ./chart

# you can use myvalues.yaml instead of --set
helm -n druid-operator-system upgrade -i --create-namespace -f myvalues.yaml namespaced-druid-operator ./chart

# ... or generate manifest.yaml to install using other means:
helm -n druid-operator-system template --set env.WATCH_NAMESPACE=""  namespaced-druid-operator ./chart --create-namespace > manifest.yaml
```

- Update settings, upgrade or rollback:

```bash
# To upgrade chart or apply changes in myvalues.yaml
helm -n druid-operator-system upgrade -f myvalues.yaml namespaced-druid-operator ./chart

# Rollback to previous revision
helm -n druid-operator-system rollback cluster-druid-operator
```

- Uninstall operator

```bash
# To avoid destroying existing clusters, helm will not uninstall its CRD. For 
# complete cleanup annotation needs to be removed first:
kubectl annotate crd druids.druid.apache.org helm.sh/resource-policy-

# This will uninstall operator
helm -n druid-operator-system  uninstall cluster-druid-operator
```

## Deploy a sample Druid cluster

- An example spec to deploy a tiny druid cluster is included. For full details on spec please see `apis/druid/v1alpha1/druid_types.go`

```bash
# deploy single node zookeeper
kubectl apply -f examples/tiny-cluster-zk.yaml

# deploy druid cluster spec
kubectl apply -f examples/tiny-cluster.yaml
```

Note that above tiny-cluster only works on a single node kubernetes cluster(e.g. typical k8s cluster setup for dev using kind or minikube) as it uses local disk as "deep storage".

## Debugging Problems

```bash
# get druid-operator pod name
kubectl get po | grep druid-operator

# check druid-operator pod logs
kubectl logs <druid-operator pod name>

# check the druid spec
kubectl describe druids tiny-cluster

# check if druid cluster is deployed
kubectl get svc | grep tiny
kubectl get cm | grep tiny
kubectl get sts | grep tiny
```
