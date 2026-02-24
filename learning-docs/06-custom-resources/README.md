# 6. Custom Resources (CR) and Custom Resource Definitions (CRD)

## What is a CRD?

A Custom Resource Definition (CRD) extends the Kubernetes API with new resource types. It's like teaching Kubernetes a new language.

**Before CRD:** Kubernetes only knows about Pods, Services, Deployments, etc.
**After CRD:** Kubernetes also knows about Druid, DruidIngestion, etc.

---

## CRDs in This Project

This project defines two CRDs:

### 1. Druid CRD
Represents a Druid cluster.

**Location:** `config/crd/bases/druid.apache.org_druids.yaml`

### 2. DruidIngestion CRD
Represents a data ingestion job.

**Location:** `config/crd/bases/druid.apache.org_druidingestions.yaml`

---

## Anatomy of a CRD

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: druids.druid.apache.org  # plural.group
spec:
  group: druid.apache.org        # API group
  names:
    kind: Druid                  # Resource kind
    listKind: DruidList          # List kind
    plural: druids               # Plural name (used in URLs)
    singular: druid              # Singular name
  scope: Namespaced              # Namespaced or Cluster
  versions:
    - name: v1alpha1             # Version
      served: true               # Is this version served?
      storage: true              # Is this the storage version?
      schema:
        openAPIV3Schema:         # Validation schema
          type: object
          properties:
            spec:
              type: object
              # ... field definitions
```

---

## Go Type Definitions

CRDs are defined in Go code, then generated into YAML.

**Location:** `apis/druid/v1alpha1/druid_types.go`

### The Druid Type

```go
// Druid is the Schema for the druids API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type Druid struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   DruidSpec          `json:"spec"`
    Status DruidClusterStatus `json:"status,omitempty"`
}
```

**Key Parts:**

| Field | Purpose |
|-------|---------|
| `TypeMeta` | API version and kind |
| `ObjectMeta` | Name, namespace, labels, annotations |
| `Spec` | Desired state (what user wants) |
| `Status` | Actual state (what exists) |

### The DruidSpec Type

```go
type DruidSpec struct {
    // Image for druid
    // +required
    Image string `json:"image,omitempty"`
    
    // CommonRuntimeProperties - common configuration
    // +required
    CommonRuntimeProperties string `json:"common.runtime.properties"`
    
    // Nodes - list of Druid node types
    // +required
    Nodes map[string]DruidNodeSpec `json:"nodes"`
    
    // RollingDeploy - enable rolling updates
    // +optional
    // +kubebuilder:default:=true
    RollingDeploy bool `json:"rollingDeploy"`
    
    // ... many more fields
}
```

### The DruidNodeSpec Type

```go
type DruidNodeSpec struct {
    // NodeType - broker, historical, coordinator, etc.
    // +required
    // +kubebuilder:validation:Enum:=historical;overlord;middleManager;indexer;broker;coordinator;router
    NodeType string `json:"nodeType"`
    
    // DruidPort - port for the node
    // +required
    DruidPort int32 `json:"druid.port"`
    
    // Replicas - number of replicas
    // +optional
    Replicas int32 `json:"replicas"`
    
    // RuntimeProperties - node-specific configuration
    // +required
    RuntimeProperties string `json:"runtime.properties"`
    
    // ... many more fields
}
```

---

## Kubebuilder Markers

The comments starting with `+kubebuilder:` are special markers:

| Marker | Purpose |
|--------|---------|
| `+kubebuilder:object:root=true` | This is a root object (has its own API endpoint) |
| `+kubebuilder:subresource:status` | Enable status subresource |
| `+kubebuilder:validation:Enum` | Allowed values |
| `+kubebuilder:default` | Default value |
| `+required` | Field is required |
| `+optional` | Field is optional |

---

## Example Custom Resource

Here's a complete Druid CR:

```yaml
apiVersion: druid.apache.org/v1alpha1
kind: Druid
metadata:
  name: tiny-cluster
  namespace: druid
spec:
  image: apache/druid:25.0.0
  startScript: /druid.sh
  
  # Common configuration for all nodes
  commonConfigMountPath: "/opt/druid/conf/druid/cluster/_common"
  
  common.runtime.properties: |
    druid.zk.service.host=zookeeper:2181
    druid.metadata.storage.type=derby
    druid.storage.type=local
    druid.extensions.loadList=["druid-kafka-indexing-service"]
  
  jvm.options: |-
    -server
    -XX:MaxDirectMemorySize=10240g
    -Duser.timezone=UTC
  
  # Node definitions
  nodes:
    # Broker nodes
    brokers:
      nodeType: broker
      druid.port: 8088
      replicas: 2
      nodeConfigMountPath: "/opt/druid/conf/druid/cluster/query/broker"
      runtime.properties: |
        druid.service=druid/broker
        druid.broker.http.numConnections=5
      extra.jvm.options: |-
        -Xmx512M
        -Xms512M
    
    # Historical nodes
    historicals:
      nodeType: historical
      druid.port: 8088
      replicas: 3
      nodeConfigMountPath: "/opt/druid/conf/druid/cluster/data/historical"
      runtime.properties: |
        druid.service=druid/historical
        druid.segmentCache.locations=[{"path":"/druid/data/segments","maxSize":10737418240}]
      volumeClaimTemplates:
        - metadata:
            name: data-volume
          spec:
            accessModes: ["ReadWriteOnce"]
            resources:
              requests:
                storage: 10Gi
    
    # Coordinator nodes
    coordinators:
      nodeType: coordinator
      druid.port: 8088
      replicas: 1
      nodeConfigMountPath: "/opt/druid/conf/druid/cluster/master/coordinator-overlord"
      runtime.properties: |
        druid.service=druid/coordinator
        druid.coordinator.asOverlord.enabled=true
```

---

## DruidIngestion CR

For data ingestion:

```yaml
apiVersion: druid.apache.org/v1alpha1
kind: DruidIngestion
metadata:
  name: kafka-ingestion
spec:
  suspend: false
  druidCluster: tiny-cluster  # Reference to Druid CR
  ingestion:
    type: kafka
    nativeSpec:
      type: kafka
      spec:
        dataSchema:
          dataSource: my-data
          timestampSpec:
            column: timestamp
            format: auto
        ioConfig:
          topic: my-topic
          consumerProperties:
            bootstrap.servers: kafka:9092
```

---

## How CRs are Processed

```
┌─────────────────────────────────────────────────────────────┐
│                    CR Processing Flow                        │
│                                                              │
│  1. User creates CR                                          │
│     kubectl apply -f druid.yaml                              │
│                    │                                         │
│                    ▼                                         │
│  2. API Server validates CR against CRD schema               │
│     - Checks required fields                                 │
│     - Validates enum values                                  │
│     - Applies defaults                                       │
│                    │                                         │
│                    ▼                                         │
│  3. CR stored in etcd                                        │
│                    │                                         │
│                    ▼                                         │
│  4. Controller receives event                                │
│     - Watch detects new/updated CR                           │
│                    │                                         │
│                    ▼                                         │
│  5. Reconcile function called                                │
│     - Reads CR spec                                          │
│     - Creates/updates K8s resources                          │
│     - Updates CR status                                      │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Status Subresource

The status subresource tracks the actual state:

```go
type DruidClusterStatus struct {
    DruidNodeStatus        DruidNodeTypeStatus `json:"druidNodeStatus,omitempty"`
    StatefulSets           []string            `json:"statefulSets,omitempty"`
    Deployments            []string            `json:"deployments,omitempty"`
    Services               []string            `json:"services,omitempty"`
    ConfigMaps             []string            `json:"configMaps,omitempty"`
    Pods                   []string            `json:"pods,omitempty"`
    PersistentVolumeClaims []string            `json:"persistentVolumeClaims,omitempty"`
}
```

**Why separate status?**
- Users update `spec` (desired state)
- Controller updates `status` (actual state)
- Prevents conflicts between user and controller

---

## Owner References

When the operator creates resources, it sets owner references:

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: druid-tiny-cluster-brokers
  ownerReferences:
    - apiVersion: druid.apache.org/v1alpha1
      kind: Druid
      name: tiny-cluster
      uid: abc-123-def
      controller: true
```

**Benefits:**
- Automatic garbage collection (delete Druid CR → all resources deleted)
- Clear ownership hierarchy
- Prevents orphaned resources

---

## Finalizers

Finalizers ensure cleanup before deletion:

```yaml
apiVersion: druid.apache.org/v1alpha1
kind: Druid
metadata:
  name: tiny-cluster
  finalizers:
    - druid.apache.org/finalizer
```

**Flow:**
1. User deletes CR: `kubectl delete druid tiny-cluster`
2. K8s sets `deletionTimestamp` but doesn't delete
3. Controller sees `deletionTimestamp`, runs cleanup
4. Controller removes finalizer
5. K8s actually deletes the CR

---

## Next Steps

Continue to [Project Structure](../07-project-structure/README.md) to understand how the codebase is organized.
