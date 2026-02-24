# 7. Project Structure

## Directory Overview

```
druid-operator/
├── apis/                    # API definitions (CRDs in Go)
│   └── druid/
│       └── v1alpha1/        # Version v1alpha1
│           ├── druid_types.go           # Druid CR definition
│           ├── druidingestion_types.go  # DruidIngestion CR definition
│           ├── groupversion_info.go     # API group registration
│           └── zz_generated.deepcopy.go # Auto-generated deep copy
│
├── controllers/             # Controller logic
│   ├── druid/               # Druid controller
│   │   ├── druid_controller.go    # Main reconciler
│   │   ├── handler.go             # Resource creation logic
│   │   ├── finalizers.go          # Cleanup logic
│   │   ├── status.go              # Status updates
│   │   └── ...
│   └── ingestion/           # Ingestion controller
│       ├── ingestion_controller.go
│       └── reconciler.go
│
├── config/                  # Kubernetes manifests
│   ├── crd/                 # CRD definitions
│   │   └── bases/           # Generated CRD YAML
│   ├── rbac/                # RBAC (permissions)
│   ├── manager/             # Operator deployment
│   └── samples/             # Example CRs
│
├── chart/                   # Helm chart
│   ├── Chart.yaml
│   ├── values.yaml
│   ├── templates/
│   └── crds/
│
├── pkg/                     # Shared packages
│   ├── druidapi/            # Druid API client
│   ├── http/                # HTTP utilities
│   └── util/                # General utilities
│
├── examples/                # Example configurations
├── docs/                    # Documentation
├── e2e/                     # End-to-end tests
├── main.go                  # Entry point
├── go.mod                   # Go module definition
└── Makefile                 # Build commands
```

---

## Key Files Explained

### 1. main.go - Entry Point

```go
func main() {
    // 1. Parse command-line flags
    flag.Parse()
    
    // 2. Create the manager
    mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
        Scheme:             scheme,
        MetricsBindAddress: metricsAddr,
        // ...
    })
    
    // 3. Register controllers
    if err = (druid.NewDruidReconciler(mgr)).SetupWithManager(mgr); err != nil {
        // Handle error
    }
    
    if err = (ingestion.NewDruidIngestionReconciler(mgr)).SetupWithManager(mgr); err != nil {
        // Handle error
    }
    
    // 4. Start the manager
    if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
        // Handle error
    }
}
```

**What it does:**
- Creates a controller-runtime Manager
- Registers both controllers (Druid and DruidIngestion)
- Starts the manager (which starts all controllers)

---

### 2. apis/druid/v1alpha1/druid_types.go - CR Definition

```go
// DruidSpec defines the desired state of Druid
type DruidSpec struct {
    Image                   string                    `json:"image,omitempty"`
    CommonRuntimeProperties string                    `json:"common.runtime.properties"`
    Nodes                   map[string]DruidNodeSpec  `json:"nodes"`
    RollingDeploy           bool                      `json:"rollingDeploy"`
    // ... many more fields
}

// Druid is the Schema for the druids API
type Druid struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec              DruidSpec          `json:"spec"`
    Status            DruidClusterStatus `json:"status,omitempty"`
}
```

**What it does:**
- Defines the structure of a Druid CR
- Used to generate CRD YAML
- Used by controller to work with Druid objects

---

### 3. controllers/druid/druid_controller.go - Main Controller

```go
// DruidReconciler reconciles a Druid object
type DruidReconciler struct {
    client.Client
    Log           logr.Logger
    Scheme        *runtime.Scheme
    ReconcileWait time.Duration
    Recorder      record.EventRecorder
}

// Reconcile is the main reconciliation function
func (r *DruidReconciler) Reconcile(ctx context.Context, request reconcile.Request) (ctrl.Result, error) {
    // 1. Fetch the Druid instance
    instance := &druidv1alpha1.Druid{}
    err := r.Get(ctx, request.NamespacedName, instance)
    if err != nil {
        if errors.IsNotFound(err) {
            return ctrl.Result{}, nil  // CR deleted, nothing to do
        }
        return ctrl.Result{}, err
    }
    
    // 2. Deploy the Druid cluster
    if err := deployDruidCluster(ctx, r.Client, instance, emitEvent); err != nil {
        return ctrl.Result{}, err
    }
    
    // 3. Requeue after wait time
    return ctrl.Result{RequeueAfter: r.ReconcileWait}, nil
}
```

**What it does:**
- Watches for Druid CR changes
- Calls `deployDruidCluster` to create/update resources
- Requeues for periodic reconciliation

---

### 4. controllers/druid/handler.go - Resource Creation

```go
func deployDruidCluster(ctx context.Context, sdk client.Client, m *v1alpha1.Druid, emitEvents EventEmitter) error {
    // 1. Validate the spec
    if err := verifyDruidSpec(m); err != nil {
        return nil
    }
    
    // 2. Get node specs in order
    allNodeSpecs := getNodeSpecsByOrder(m)
    
    // 3. Create common ConfigMap
    if _, err := sdkCreateOrUpdateAsNeeded(ctx, sdk,
        func() (object, error) { return makeCommonConfigMap(ctx, sdk, m, ls) },
        // ...
    ); err != nil {
        return err
    }
    
    // 4. For each node type, create resources
    for _, elem := range allNodeSpecs {
        // Create node ConfigMap
        // Create Services
        // Create StatefulSet or Deployment
        // Create PodDisruptionBudget
        // Create HPA
        // Create Ingress
    }
    
    // 5. Delete unused resources
    // 6. Update status
    
    return nil
}
```

**What it does:**
- Creates all Kubernetes resources for a Druid cluster
- Handles create, update, and delete operations
- Manages rolling updates

---

### 5. controllers/ingestion/reconciler.go - Ingestion Logic

```go
func (r *DruidIngestionReconciler) do(ctx context.Context, di *v1alpha1.DruidIngestion) error {
    // 1. Get Druid router service URL
    svcName, err := druidapi.GetRouterSvcUrl(di.Namespace, di.Spec.DruidClusterName, r.Client)
    
    // 2. Create or update ingestion task
    _, err = r.CreateOrUpdate(di, svcName, *build, auth)
    
    // 3. Handle finalizers for cleanup
    if di.ObjectMeta.DeletionTimestamp.IsZero() {
        // Add finalizer if not present
    } else {
        // Cleanup: shutdown ingestion task
    }
    
    return nil
}
```

**What it does:**
- Manages Druid ingestion tasks via Druid's REST API
- Creates/updates/deletes ingestion supervisors
- Handles compaction and rules

---

## Package Dependencies

```
main.go
    │
    ├── apis/druid/v1alpha1
    │   └── Druid, DruidIngestion types
    │
    ├── controllers/druid
    │   ├── druid_controller.go (Reconciler)
    │   ├── handler.go (Resource creation)
    │   ├── finalizers.go (Cleanup)
    │   └── status.go (Status updates)
    │
    ├── controllers/ingestion
    │   ├── ingestion_controller.go (Reconciler)
    │   └── reconciler.go (Ingestion logic)
    │
    └── pkg/
        ├── druidapi/ (Druid REST API client)
        ├── http/ (HTTP utilities)
        └── util/ (General utilities)
```

---

## Configuration Files

### config/crd/bases/druid.apache.org_druids.yaml
Generated CRD for Druid. Created by running `make manifests`.

### config/rbac/role.yaml
RBAC permissions the operator needs:
```yaml
rules:
- apiGroups: ["druid.apache.org"]
  resources: ["druids"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["apps"]
  resources: ["statefulsets", "deployments"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
# ... more permissions
```

### config/manager/manager.yaml
Deployment for the operator itself.

---

## Helm Chart Structure

```
chart/
├── Chart.yaml           # Chart metadata
├── values.yaml          # Default values
├── crds/                # CRD definitions
│   ├── druid.apache.org_druids.yaml
│   └── druid.apache.org_druidingestions.yaml
└── templates/           # Kubernetes manifests
    ├── deployment.yaml  # Operator deployment
    ├── service.yaml     # Operator service
    ├── rbac_*.yaml      # RBAC resources
    └── service_account.yaml
```

---

## Build Commands (Makefile)

```bash
# Generate CRD manifests from Go types
make manifests

# Generate deep copy functions
make generate

# Build the operator binary
make build

# Build Docker image
make docker-build

# Run tests
make test

# Deploy to cluster
make deploy

# Install CRDs only
make install
```

---

## Next Steps

Continue to [The Reconciliation Loop](../08-reconciliation-loop/README.md) to understand the core operator logic in detail.
