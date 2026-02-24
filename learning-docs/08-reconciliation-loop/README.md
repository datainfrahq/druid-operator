# 8. The Reconciliation Loop

## What is Reconciliation?

Reconciliation is the process of making the **actual state** match the **desired state**.

```
Desired State (CR spec)  →  Reconcile  →  Actual State (K8s resources)
```

The reconciler runs in a loop:
1. **Observe** - Read the current state
2. **Analyze** - Compare with desired state
3. **Act** - Create, update, or delete resources
4. **Repeat** - Requeue and do it again

---

## The Reconcile Function

**Location:** `controllers/druid/druid_controller.go`

```go
func (r *DruidReconciler) Reconcile(ctx context.Context, request reconcile.Request) (ctrl.Result, error) {
    // request contains the name and namespace of the CR that triggered reconciliation
    
    // Step 1: Fetch the Druid CR
    instance := &druidv1alpha1.Druid{}
    err := r.Get(ctx, request.NamespacedName, instance)
    if err != nil {
        if errors.IsNotFound(err) {
            // CR was deleted, nothing to do
            // Owned resources will be garbage collected automatically
            return ctrl.Result{}, nil
        }
        // Error reading the object - requeue
        return ctrl.Result{}, err
    }
    
    // Step 2: Initialize event emitter for logging
    var emitEvent EventEmitter = EmitEventFuncs{r.Recorder}
    
    // Step 3: Deploy/Update the Druid cluster
    if err := deployDruidCluster(ctx, r.Client, instance, emitEvent); err != nil {
        return ctrl.Result{}, err
    }
    
    // Step 4: Update dynamic configurations
    if err := updateDruidDynamicConfigs(ctx, r.Client, instance, emitEvent); err != nil {
        return ctrl.Result{}, err
    }
    
    // Step 5: Requeue after wait time (default 10 seconds)
    return ctrl.Result{RequeueAfter: r.ReconcileWait}, nil
}
```

---

## Return Values

The `Reconcile` function returns `(ctrl.Result, error)`:

| Return | Meaning |
|--------|---------|
| `Result{}, nil` | Success, don't requeue |
| `Result{Requeue: true}, nil` | Success, requeue immediately |
| `Result{RequeueAfter: 10s}, nil` | Success, requeue after 10 seconds |
| `Result{}, err` | Error, requeue with backoff |

---

## The deployDruidCluster Function

**Location:** `controllers/druid/handler.go`

This is where the magic happens:

```go
func deployDruidCluster(ctx context.Context, sdk client.Client, m *v1alpha1.Druid, emitEvents EventEmitter) error {
    
    // 1. VALIDATE the spec
    if err := verifyDruidSpec(m); err != nil {
        emitEvents.EmitEventGeneric(m, "DruidOperatorInvalidSpec", "", err)
        return nil  // Don't retry invalid specs
    }
    
    // 2. GET NODE SPECS in correct order
    // Order matters for rolling updates!
    allNodeSpecs := getNodeSpecsByOrder(m)
    // Returns: historicals → middleManagers → indexers → brokers → coordinators → overlords → routers
    
    // 3. TRACK created resources
    statefulSetNames := make(map[string]bool)
    serviceNames := make(map[string]bool)
    configMapNames := make(map[string]bool)
    // ... more maps
    
    // 4. CREATE COMMON CONFIGMAP
    // Contains common.runtime.properties shared by all nodes
    commonConfig, err := makeCommonConfigMap(ctx, sdk, m, ls)
    commonConfigSHA, _ := getObjectHash(commonConfig)  // For change detection
    
    sdkCreateOrUpdateAsNeeded(ctx, sdk,
        func() (object, error) { return makeCommonConfigMap(ctx, sdk, m, ls) },
        func() object { return &v1.ConfigMap{} },
        // ...
    )
    
    // 5. HANDLE DELETION
    if m.GetDeletionTimestamp() != nil {
        return executeFinalizers(ctx, sdk, m, emitEvents)
    }
    
    // 6. UPDATE FINALIZERS
    if err := updateFinalizers(ctx, sdk, m, emitEvents); err != nil {
        return err
    }
    
    // 7. FOR EACH NODE TYPE
    for _, elem := range allNodeSpecs {
        key := elem.key
        nodeSpec := elem.spec
        
        // Create unique identifier
        nodeSpecUniqueStr := makeNodeSpecificUniqueString(m, key)
        // e.g., "druid-tiny-cluster-brokers"
        
        // Create labels
        lm := makeLabelsForNodeSpec(&nodeSpec, m, m.Name, nodeSpecUniqueStr)
        
        // 7a. Create node ConfigMap
        nodeConfig, _ := makeConfigMapForNodeSpec(&nodeSpec, m, lm, nodeSpecUniqueStr)
        sdkCreateOrUpdateAsNeeded(ctx, sdk, ...)
        
        // 7b. Create Services
        for _, svc := range services {
            sdkCreateOrUpdateAsNeeded(ctx, sdk,
                func() (object, error) { return makeService(&svc, &nodeSpec, m, lm, nodeSpecUniqueStr) },
                // ...
            )
        }
        
        // 7c. Create StatefulSet or Deployment
        if nodeSpec.Kind == "Deployment" {
            sdkCreateOrUpdateAsNeeded(ctx, sdk,
                func() (object, error) { return makeDeployment(&nodeSpec, m, ...) },
                // ...
            )
        } else {
            // Default: StatefulSet
            sdkCreateOrUpdateAsNeeded(ctx, sdk,
                func() (object, error) { return makeStatefulSet(&nodeSpec, m, ...) },
                // ...
            )
        }
        
        // 7d. Create optional resources
        // - Ingress
        // - PodDisruptionBudget
        // - HorizontalPodAutoscaler
        // - PersistentVolumeClaims
    }
    
    // 8. DELETE UNUSED RESOURCES
    // If a node was removed from spec, delete its resources
    deleteUnusedResources(ctx, sdk, m, statefulSetNames, ...)
    
    // 9. UPDATE STATUS
    updatedStatus := v1alpha1.DruidClusterStatus{
        StatefulSets: ...,
        Services: ...,
        ConfigMaps: ...,
        Pods: ...,
    }
    druidClusterStatusPatcher(ctx, sdk, updatedStatus, m, emitEvents)
    
    return nil
}
```

---

## Create or Update Pattern

The `sdkCreateOrUpdateAsNeeded` function handles idempotent resource management:

```go
func sdkCreateOrUpdateAsNeeded(
    ctx context.Context,
    sdk client.Client,
    objFn func() (object, error),      // Function to create the desired object
    emptyObjFn func() object,          // Function to create empty object for Get
    isEqualFn func(prev, curr object) bool,  // Compare function
    updaterFn func(prev, curr object), // Update function
    drd *v1alpha1.Druid,
    names map[string]bool,
    emitEvent EventEmitter,
) (DruidNodeStatus, error) {
    
    // 1. Create the desired object
    obj, err := objFn()
    names[obj.GetName()] = true  // Track this resource
    
    // 2. Add owner reference (for garbage collection)
    addOwnerRefToObject(obj, asOwner(drd))
    
    // 3. Add hash annotation (for change detection)
    addHashToObject(obj)
    
    // 4. Try to get existing object
    prevObj := emptyObjFn()
    err := sdk.Get(ctx, namespacedName, prevObj)
    
    if err != nil && apierrors.IsNotFound(err) {
        // 5a. Object doesn't exist - CREATE it
        return writers.Create(ctx, sdk, drd, obj, emitEvent)
    }
    
    // 5b. Object exists - check if UPDATE needed
    if obj.GetAnnotations()[druidOpResourceHash] != prevObj.GetAnnotations()[druidOpResourceHash] {
        // Hash changed - UPDATE
        obj.SetResourceVersion(prevObj.GetResourceVersion())
        return writers.Update(ctx, sdk, drd, obj, emitEvent)
    }
    
    // 5c. No change needed
    return "", nil
}
```

---

## Rolling Updates

The operator supports Druid's recommended rolling update order:

```go
func getNodeSpecsByOrder(m *v1alpha1.Druid) []keyAndNodeSpec {
    // Order defined by Druid documentation:
    // https://druid.apache.org/docs/latest/operations/rolling-updates.html
    
    nodeSpecsByOrder := []string{
        "historicals",      // 1. Update historicals first
        "middleManagers",   // 2. Then middle managers
        "indexers",         // 3. Then indexers
        "brokers",          // 4. Then brokers
        "coordinators",     // 5. Then coordinators
        "overlords",        // 6. Then overlords
        "routers",          // 7. Finally routers
    }
    
    // Return specs in this order
}
```

During rolling updates:
```go
if m.Spec.RollingDeploy {
    // Check if previous node type is fully deployed
    done, err := isObjFullyDeployed(ctx, sdk, nodeSpec, ...)
    if !done {
        return err  // Wait for previous to complete
    }
}
```

---

## Finalizers

Finalizers ensure cleanup when a Druid CR is deleted:

```go
func executeFinalizers(ctx context.Context, sdk client.Client, m *v1alpha1.Druid, emitEvents EventEmitter) error {
    // 1. Check if finalizer exists
    if !controllerutil.ContainsFinalizer(m, finalizerName) {
        return nil
    }
    
    // 2. Perform cleanup
    // - Delete PVCs if configured
    // - Any other cleanup needed
    
    // 3. Remove finalizer
    controllerutil.RemoveFinalizer(m, finalizerName)
    if err := sdk.Update(ctx, m); err != nil {
        return err
    }
    
    return nil
}
```

---

## Status Updates

The operator updates the CR status to reflect actual state:

```go
func druidClusterStatusPatcher(ctx context.Context, sdk client.Client, updatedStatus v1alpha1.DruidClusterStatus, m *v1alpha1.Druid, emitEvents EventEmitter) error {
    // Get current CR
    currentDruid := &v1alpha1.Druid{}
    sdk.Get(ctx, types.NamespacedName{Name: m.Name, Namespace: m.Namespace}, currentDruid)
    
    // Update status
    currentDruid.Status = updatedStatus
    
    // Patch status subresource
    return sdk.Status().Update(ctx, currentDruid)
}
```

---

## Event Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                     Reconciliation Flow                              │
│                                                                      │
│  ┌──────────┐                                                       │
│  │  Event   │  CR created/updated/deleted                           │
│  │ (Watch)  │  or periodic requeue                                  │
│  └────┬─────┘                                                       │
│       │                                                              │
│       ▼                                                              │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                    Reconcile()                                │   │
│  │                                                               │   │
│  │  1. Get Druid CR                                              │   │
│  │     └─► Not found? Return (garbage collection handles rest)   │   │
│  │                                                               │   │
│  │  2. deployDruidCluster()                                      │   │
│  │     ├─► Validate spec                                         │   │
│  │     ├─► Check deletion timestamp                              │   │
│  │     │   └─► If deleting, run finalizers                       │   │
│  │     ├─► Create common ConfigMap                               │   │
│  │     ├─► For each node type (in order):                        │   │
│  │     │   ├─► Create node ConfigMap                             │   │
│  │     │   ├─► Create Services                                   │   │
│  │     │   ├─► Create StatefulSet/Deployment                     │   │
│  │     │   ├─► Wait for rollout (if rolling deploy)              │   │
│  │     │   └─► Create optional resources (PDB, HPA, Ingress)     │   │
│  │     ├─► Delete unused resources                               │   │
│  │     └─► Update status                                         │   │
│  │                                                               │   │
│  │  3. Return Result{RequeueAfter: 10s}                          │   │
│  │                                                               │   │
│  └──────────────────────────────────────────────────────────────┘   │
│       │                                                              │
│       │ After 10 seconds                                            │
│       ▼                                                              │
│  ┌──────────┐                                                       │
│  │ Requeue  │  Start again                                          │
│  └──────────┘                                                       │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Key Concepts

### 1. Idempotency
Running reconcile multiple times produces the same result:
- If resource exists and matches, do nothing
- If resource exists but differs, update it
- If resource doesn't exist, create it

### 2. Level-Triggered vs Edge-Triggered
- **Edge-triggered**: React to events (what changed)
- **Level-triggered**: React to state (what is)

Operators are level-triggered - they look at current state, not events.

### 3. Eventual Consistency
The system may take multiple reconciliations to reach desired state:
- First reconcile: Create ConfigMaps
- Second reconcile: Create StatefulSets
- Third reconcile: Wait for pods to be ready
- Fourth reconcile: All good, just monitor

---

## Next Steps

Continue to [Apache Druid Overview](../09-druid-overview/README.md) to understand what Druid is and why it needs an operator.
