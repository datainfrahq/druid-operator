# Quick Reference Guide

## Common Commands

### Operator Installation
```bash
# Add Helm repo
helm repo add datainfra https://charts.datainfra.io
helm repo update

# Install operator
helm install druid-operator datainfra/druid-operator -n druid-operator-system --create-namespace

# Uninstall
helm uninstall druid-operator -n druid-operator-system
```

### Working with Druid CRs
```bash
# Create Druid cluster
kubectl apply -f examples/tiny-cluster.yaml

# List Druid clusters
kubectl get druids
kubectl get druid

# Describe Druid cluster
kubectl describe druid tiny-cluster

# Get Druid cluster YAML
kubectl get druid tiny-cluster -o yaml

# Delete Druid cluster
kubectl delete druid tiny-cluster
```

### Debugging
```bash
# Check operator logs
kubectl logs -n druid-operator-system -l app.kubernetes.io/name=druid-operator -f

# Check Druid pod logs
kubectl logs druid-tiny-cluster-brokers-0

# Check events
kubectl get events --sort-by='.lastTimestamp'

# Check all resources for a Druid cluster
kubectl get all -l druid_cr=tiny-cluster
```

### Development
```bash
# Generate CRD manifests
make manifests

# Generate deep copy code
make generate

# Build operator
make build

# Run tests
make test

# Build Docker image
make docker-build IMG=my-registry/druid-operator:tag

# Push Docker image
make docker-push IMG=my-registry/druid-operator:tag
```

---

## Key Files Reference

| File | Purpose |
|------|---------|
| `main.go` | Entry point, starts manager and controllers |
| `apis/druid/v1alpha1/druid_types.go` | Druid CR definition |
| `apis/druid/v1alpha1/druidingestion_types.go` | DruidIngestion CR definition |
| `controllers/druid/druid_controller.go` | Druid reconciler |
| `controllers/druid/handler.go` | Resource creation logic |
| `controllers/ingestion/reconciler.go` | Ingestion logic |
| `config/crd/bases/*.yaml` | Generated CRD manifests |
| `chart/values.yaml` | Helm chart default values |

---

## Druid Node Types

| Node Type | Purpose | Stateful | Default Kind |
|-----------|---------|----------|--------------|
| `coordinator` | Manages data distribution | Yes | StatefulSet |
| `overlord` | Manages ingestion tasks | Yes | StatefulSet |
| `broker` | Query routing | No | StatefulSet |
| `router` | API gateway | No | StatefulSet |
| `historical` | Serves historical data | Yes | StatefulSet |
| `middleManager` | Runs ingestion tasks | Yes | StatefulSet |
| `indexer` | Alternative to middleManager | Yes | StatefulSet |

---

## CR Spec Quick Reference

```yaml
apiVersion: druid.apache.org/v1alpha1
kind: Druid
metadata:
  name: my-cluster
spec:
  # Required
  image: apache/druid:25.0.0
  common.runtime.properties: |
    druid.zk.service.host=zk:2181
  nodes:
    brokers:
      nodeType: broker
      druid.port: 8088
      replicas: 1
      runtime.properties: |
        druid.service=druid/broker
      nodeConfigMountPath: /opt/druid/conf/druid/cluster/query/broker
  
  # Optional
  startScript: /druid.sh
  commonConfigMountPath: /opt/druid/conf/druid/cluster/_common
  rollingDeploy: true
  forceDeleteStsPodOnError: true
  deleteOrphanPvc: true
  
  # JVM and logging
  jvm.options: |-
    -server
    -Xmx1g
  log4j.config: |-
    <?xml version="1.0"?>
    ...
  
  # Pod configuration
  podLabels:
    app: druid
  podAnnotations:
    key: value
  securityContext:
    runAsUser: 1000
  
  # Volumes
  volumes:
    - name: data
      emptyDir: {}
  volumeMounts:
    - name: data
      mountPath: /druid/data
```

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `WATCH_NAMESPACE` | "" | Namespace(s) to watch (empty = all) |
| `DENY_LIST` | "default,kube-system" | Namespaces to ignore |
| `RECONCILE_WAIT` | "10s" | Time between reconciliations |
| `MAX_CONCURRENT_RECONCILES` | "1" | Max parallel reconciliations |

---

## Useful Links

- [Apache Druid Documentation](https://druid.apache.org/docs/latest/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Kubebuilder Book](https://book.kubebuilder.io/)
- [Controller-Runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime)
- [Go Documentation](https://go.dev/doc/)
