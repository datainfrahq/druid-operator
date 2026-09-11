# 10. Complete Flow Walkthrough

## End-to-End: From YAML to Running Druid Cluster

Let's trace what happens when you deploy a Druid cluster using this operator.

---

## Step 1: Install the Operator

```bash
# Add Helm repository
helm repo add datainfra https://charts.datainfra.io

# Install the operator
helm install druid-operator datainfra/druid-operator -n druid-operator-system --create-namespace
```

**What happens:**
1. Helm creates the `druid-operator-system` namespace
2. CRDs are installed (`Druid`, `DruidIngestion`)
3. RBAC resources are created (ServiceAccount, Role, RoleBinding)
4. Operator Deployment is created
5. Operator pod starts running

```
┌─────────────────────────────────────────────────────────────┐
│                 druid-operator-system namespace              │
│                                                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              druid-operator Deployment               │    │
│  │                                                      │    │
│  │  ┌────────────────────────────────────────────────┐ │    │
│  │  │           druid-operator Pod                   │ │    │
│  │  │                                                │ │    │
│  │  │  main.go starts:                               │ │    │
│  │  │  - Creates Manager                             │ │    │
│  │  │  - Registers DruidReconciler                   │ │    │
│  │  │  - Registers DruidIngestionReconciler          │ │    │
│  │  │  - Starts watching for CRs                     │ │    │
│  │  │                                                │ │    │
│  │  └────────────────────────────────────────────────┘ │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Step 2: Create a Druid CR

```bash
kubectl apply -f - <<EOF
apiVersion: druid.apache.org/v1alpha1
kind: Druid
metadata:
  name: tiny-cluster
  namespace: druid
spec:
  image: apache/druid:25.0.0
  startScript: /druid.sh
  
  common.runtime.properties: |
    druid.zk.service.host=zookeeper:2181
    druid.metadata.storage.type=derby
    druid.storage.type=local
  
  nodes:
    brokers:
      nodeType: broker
      druid.port: 8088
      replicas: 1
      runtime.properties: |
        druid.service=druid/broker
    
    historicals:
      nodeType: historical
      druid.port: 8088
      replicas: 1
      runtime.properties: |
        druid.service=druid/historical
EOF
```

**What happens:**
1. `kubectl` sends the YAML to Kubernetes API server
2. API server validates against CRD schema
3. CR is stored in etcd
4. API server notifies watchers (including our operator)

---

## Step 3: Operator Receives Event

The controller-runtime framework detects the new CR:

```go
// In druid_controller.go
func (r *DruidReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&druidv1alpha1.Druid{}).  // Watch Druid CRs
        WithEventFilter(GenericPredicates{}).
        Complete(r)
}
```

The `Reconcile` function is called:

```go
func (r *DruidReconciler) Reconcile(ctx context.Context, request reconcile.Request) (ctrl.Result, error) {
    // request.NamespacedName = "druid/tiny-cluster"
    
    // Fetch the Druid CR
    instance := &druidv1alpha1.Druid{}
    r.Get(ctx, request.NamespacedName, instance)
    
    // Deploy the cluster
    deployDruidCluster(ctx, r.Client, instance, emitEvent)
    
    // Requeue after 10 seconds
    return ctrl.Result{RequeueAfter: r.ReconcileWait}, nil
}
```

---

## Step 4: deployDruidCluster Executes

### 4.1 Validate Spec
```go
if err := verifyDruidSpec(m); err != nil {
    // Invalid spec - emit event and return
    return nil
}
```

### 4.2 Create Common ConfigMap
```go
// Creates ConfigMap with common.runtime.properties
commonConfig, _ := makeCommonConfigMap(ctx, sdk, m, ls)
sdkCreateOrUpdateAsNeeded(ctx, sdk, ...)
```

**Created resource:**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: druid-tiny-cluster-druid-common-config
  namespace: druid
  ownerReferences:
    - apiVersion: druid.apache.org/v1alpha1
      kind: Druid
      name: tiny-cluster
data:
  common.runtime.properties: |
    druid.zk.service.host=zookeeper:2181
    druid.metadata.storage.type=derby
    druid.storage.type=local
```

### 4.3 Process Each Node Type

For each node (brokers, historicals):

#### Create Node ConfigMap
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: druid-tiny-cluster-brokers-config
data:
  runtime.properties: |
    druid.service=druid/broker
  jvm.config: |
    -server
    -Xmx512M
```

#### Create Service
```yaml
apiVersion: v1
kind: Service
metadata:
  name: druid-tiny-cluster-brokers
spec:
  selector:
    app: druid
    component: broker
    druid_cr: tiny-cluster
  ports:
    - port: 8088
      targetPort: 8088
```

#### Create StatefulSet
```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: druid-tiny-cluster-brokers
spec:
  serviceName: druid-tiny-cluster-brokers
  replicas: 1
  selector:
    matchLabels:
      app: druid
      component: broker
  template:
    metadata:
      labels:
        app: druid
        component: broker
    spec:
      containers:
        - name: druid
          image: apache/druid:25.0.0
          command: ["/druid.sh", "broker"]
          ports:
            - containerPort: 8088
          volumeMounts:
            - name: common-config-volume
              mountPath: /opt/druid/conf/druid/cluster/_common
            - name: nodetype-config-volume
              mountPath: /opt/druid/conf/druid/cluster/query/broker
      volumes:
        - name: common-config-volume
          configMap:
            name: druid-tiny-cluster-druid-common-config
        - name: nodetype-config-volume
          configMap:
            name: druid-tiny-cluster-brokers-config
```

---

## Step 5: Kubernetes Creates Pods

Kubernetes StatefulSet controller:
1. Sees new StatefulSet
2. Creates Pod `druid-tiny-cluster-brokers-0`
3. Schedules Pod to a node
4. Kubelet pulls image and starts container

```
┌─────────────────────────────────────────────────────────────┐
│                      druid namespace                         │
│                                                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                    ConfigMaps                        │    │
│  │  - druid-tiny-cluster-druid-common-config           │    │
│  │  - druid-tiny-cluster-brokers-config                │    │
│  │  - druid-tiny-cluster-historicals-config            │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                     Services                         │    │
│  │  - druid-tiny-cluster-brokers                       │    │
│  │  - druid-tiny-cluster-historicals                   │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                   StatefulSets                       │    │
│  │                                                      │    │
│  │  druid-tiny-cluster-brokers (replicas: 1)           │    │
│  │  └── Pod: druid-tiny-cluster-brokers-0              │    │
│  │      └── Container: druid (apache/druid:25.0.0)     │    │
│  │                                                      │    │
│  │  druid-tiny-cluster-historicals (replicas: 1)       │    │
│  │  └── Pod: druid-tiny-cluster-historicals-0          │    │
│  │      └── Container: druid (apache/druid:25.0.0)     │    │
│  │                                                      │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Step 6: Status Update

After creating resources, the operator updates the CR status:

```go
updatedStatus := v1alpha1.DruidClusterStatus{
    StatefulSets: []string{"druid-tiny-cluster-brokers", "druid-tiny-cluster-historicals"},
    Services:     []string{"druid-tiny-cluster-brokers", "druid-tiny-cluster-historicals"},
    ConfigMaps:   []string{"druid-tiny-cluster-druid-common-config", ...},
    Pods:         []string{"druid-tiny-cluster-brokers-0", "druid-tiny-cluster-historicals-0"},
    DruidNodeStatus: DruidNodeTypeStatus{
        DruidNodeConditionStatus: "True",
        DruidNodeConditionType:   "DruidClusterReady",
    },
}
druidClusterStatusPatcher(ctx, sdk, updatedStatus, m, emitEvents)
```

You can see this:
```bash
kubectl get druid tiny-cluster -n druid -o yaml
```

---

## Step 7: Continuous Reconciliation

The operator requeues after 10 seconds:
```go
return ctrl.Result{RequeueAfter: r.ReconcileWait}, nil
```

On each reconciliation:
1. Check if CR still exists
2. Compare desired state with actual state
3. Create/update/delete resources as needed
4. Update status
5. Requeue

This ensures:
- Self-healing (recreate deleted resources)
- Configuration drift detection
- Continuous monitoring

---

## Step 8: Handling Updates

When you update the CR:
```bash
kubectl patch druid tiny-cluster -n druid --type merge -p '{"spec":{"nodes":{"brokers":{"replicas":2}}}}'
```

The operator:
1. Receives update event
2. Reads new spec (replicas: 2)
3. Updates StatefulSet
4. Kubernetes creates new pod `druid-tiny-cluster-brokers-1`

---

## Step 9: Handling Deletion

When you delete the CR:
```bash
kubectl delete druid tiny-cluster -n druid
```

The operator:
1. Sees `deletionTimestamp` is set
2. Runs finalizers (cleanup PVCs if configured)
3. Removes finalizer
4. Kubernetes garbage collects all owned resources

---

## Complete Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Kubernetes Cluster                               │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │                    druid-operator-system namespace                  │ │
│  │                                                                     │ │
│  │  ┌─────────────────────────────────────────────────────────────┐  │ │
│  │  │                  druid-operator Pod                          │  │ │
│  │  │                                                              │  │ │
│  │  │  ┌──────────────────┐    ┌──────────────────────────────┐   │  │ │
│  │  │  │ DruidReconciler  │    │ DruidIngestionReconciler     │   │  │ │
│  │  │  │                  │    │                              │   │  │ │
│  │  │  │ Watches: Druid   │    │ Watches: DruidIngestion      │   │  │ │
│  │  │  │ Creates: STS,    │    │ Calls: Druid REST API        │   │  │ │
│  │  │  │   Svc, CM, etc.  │    │                              │   │  │ │
│  │  │  └────────┬─────────┘    └──────────────────────────────┘   │  │ │
│  │  │           │                                                  │  │ │
│  │  └───────────┼──────────────────────────────────────────────────┘  │ │
│  │              │                                                      │ │
│  └──────────────┼──────────────────────────────────────────────────────┘ │
│                 │                                                        │
│                 │ Creates/Updates/Deletes                                │
│                 ▼                                                        │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │                        druid namespace                              │ │
│  │                                                                     │ │
│  │  ┌─────────────┐                                                   │ │
│  │  │  Druid CR   │  ◄── User creates this                            │ │
│  │  │ tiny-cluster│                                                   │ │
│  │  └──────┬──────┘                                                   │ │
│  │         │ Owner Reference                                          │ │
│  │         ▼                                                          │ │
│  │  ┌─────────────────────────────────────────────────────────────┐  │ │
│  │  │                    Created Resources                         │  │ │
│  │  │                                                              │  │ │
│  │  │  ConfigMaps:                                                 │  │ │
│  │  │  ├── druid-tiny-cluster-druid-common-config                  │  │ │
│  │  │  ├── druid-tiny-cluster-brokers-config                       │  │ │
│  │  │  └── druid-tiny-cluster-historicals-config                   │  │ │
│  │  │                                                              │  │ │
│  │  │  Services:                                                   │  │ │
│  │  │  ├── druid-tiny-cluster-brokers                              │  │ │
│  │  │  └── druid-tiny-cluster-historicals                          │  │ │
│  │  │                                                              │  │ │
│  │  │  StatefulSets:                                               │  │ │
│  │  │  ├── druid-tiny-cluster-brokers                              │  │ │
│  │  │  │   └── Pod: druid-tiny-cluster-brokers-0                   │  │ │
│  │  │  └── druid-tiny-cluster-historicals                          │  │ │
│  │  │      └── Pod: druid-tiny-cluster-historicals-0               │  │ │
│  │  │                                                              │  │ │
│  │  └─────────────────────────────────────────────────────────────┘  │ │
│  │                                                                     │ │
│  └─────────────────────────────────────────────────────────────────────┘ │
│                                                                          │
└──────────────────────────────────────────────────────────────────────────┘
```

---

## Summary

1. **User** creates a Druid CR (YAML file)
2. **API Server** validates and stores the CR
3. **Operator** receives event via Watch
4. **Reconciler** reads CR spec (desired state)
5. **Reconciler** creates/updates Kubernetes resources
6. **Kubernetes** schedules and runs Druid pods
7. **Operator** updates CR status (actual state)
8. **Operator** requeues for continuous monitoring
9. **Repeat** every 10 seconds

This is the **operator pattern** in action - encoding operational knowledge into software that continuously ensures your desired state is maintained!

---

## Congratulations!

You now understand:
- What this repository does
- Why Go is used instead of Java
- What Kubernetes operators are
- How Custom Resources work
- How the reconciliation loop operates
- How Apache Druid is deployed on Kubernetes

**Next steps:**
1. Try deploying the operator locally
2. Create a simple Druid cluster
3. Modify the CR and watch the operator react
4. Read the actual code with this understanding
5. Consider contributing to the project!
