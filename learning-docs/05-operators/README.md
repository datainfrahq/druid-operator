# 5. What is a Kubernetes Operator?

## The Problem Operators Solve

Kubernetes is great at managing stateless applications. But complex, stateful applications like databases need:
- **Domain knowledge** - How to properly scale, backup, upgrade
- **Operational procedures** - What to do when things go wrong
- **Configuration management** - Complex interdependent settings

Traditionally, this required a human operator (SRE/DevOps engineer) who:
1. Understands the application deeply
2. Monitors the application
3. Takes corrective actions when needed
4. Performs upgrades, backups, scaling

**An Operator automates this human operator's knowledge into software!**

---

## The Operator Pattern

An Operator is a Kubernetes controller that:
1. **Extends Kubernetes** with Custom Resource Definitions (CRDs)
2. **Watches** for changes to Custom Resources
3. **Reconciles** the actual state to match the desired state
4. **Encodes operational knowledge** in code

```
┌─────────────────────────────────────────────────────────────┐
│                     Operator Pattern                         │
│                                                              │
│   ┌─────────────┐                    ┌─────────────────┐    │
│   │   Custom    │   Watches          │    Controller   │    │
│   │  Resource   │◄──────────────────►│   (Reconciler)  │    │
│   │   (CR)      │                    │                 │    │
│   └─────────────┘                    └────────┬────────┘    │
│         │                                     │              │
│         │ Desired State                       │ Creates/     │
│         │                                     │ Updates/     │
│         │                                     │ Deletes      │
│         │                                     ▼              │
│         │                            ┌─────────────────┐    │
│         │                            │  Kubernetes     │    │
│         │                            │  Resources      │    │
│         │                            │  (Pods, Svcs,   │    │
│         └───────────────────────────►│   ConfigMaps)   │    │
│                  Actual State        └─────────────────┘    │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Key Components of an Operator

### 1. Custom Resource Definition (CRD)

A CRD tells Kubernetes about a new resource type:

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: druids.druid.apache.org
spec:
  group: druid.apache.org
  versions:
    - name: v1alpha1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                image:
                  type: string
                nodes:
                  type: object
  scope: Namespaced
  names:
    plural: druids
    singular: druid
    kind: Druid
```

After applying this CRD, Kubernetes understands what a "Druid" is!

### 2. Custom Resource (CR)

A CR is an instance of a CRD - your actual configuration:

```yaml
apiVersion: druid.apache.org/v1alpha1
kind: Druid
metadata:
  name: my-druid-cluster
spec:
  image: apache/druid:25.0.0
  nodes:
    brokers:
      nodeType: broker
      replicas: 2
```

### 3. Controller

The controller watches CRs and takes action:

```go
// Simplified controller logic
func (r *DruidReconciler) Reconcile(ctx context.Context, req Request) (Result, error) {
    // 1. Get the Druid CR
    druid := &v1alpha1.Druid{}
    err := r.Get(ctx, req.NamespacedName, druid)
    
    // 2. Create/Update Kubernetes resources based on CR
    // - Create ConfigMaps for Druid configuration
    // - Create StatefulSets for Druid nodes
    // - Create Services for networking
    
    // 3. Return and requeue after some time
    return Result{RequeueAfter: 10 * time.Second}, nil
}
```

---

## The Reconciliation Loop

The heart of an operator is the **reconciliation loop**:

```
┌─────────────────────────────────────────────────────────────┐
│                   Reconciliation Loop                        │
│                                                              │
│    ┌──────────┐                                             │
│    │  Event   │  (CR created, updated, deleted,             │
│    │ Trigger  │   or periodic requeue)                      │
│    └────┬─────┘                                             │
│         │                                                    │
│         ▼                                                    │
│    ┌──────────┐                                             │
│    │  Fetch   │  Get the CR from Kubernetes API             │
│    │   CR     │                                             │
│    └────┬─────┘                                             │
│         │                                                    │
│         ▼                                                    │
│    ┌──────────┐                                             │
│    │ Compare  │  Desired State (CR) vs Actual State (K8s)   │
│    │  States  │                                             │
│    └────┬─────┘                                             │
│         │                                                    │
│         ▼                                                    │
│    ┌──────────┐                                             │
│    │  Take    │  Create, Update, or Delete resources        │
│    │  Action  │                                             │
│    └────┬─────┘                                             │
│         │                                                    │
│         ▼                                                    │
│    ┌──────────┐                                             │
│    │ Requeue  │  Schedule next reconciliation               │
│    └──────────┘                                             │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Key Principles

1. **Idempotent** - Running reconcile multiple times has the same effect as running once
2. **Level-triggered** - Reacts to current state, not events
3. **Eventually consistent** - May take multiple reconciliations to reach desired state

---

## What Makes a Good Operator?

### Level 1: Basic Install
- Automated application provisioning
- Basic configuration

### Level 2: Seamless Upgrades
- Patch and minor version upgrades
- Rolling updates

### Level 3: Full Lifecycle
- Backup and restore
- Application-aware scaling

### Level 4: Deep Insights
- Metrics and alerts
- Log processing

### Level 5: Auto Pilot
- Auto-scaling
- Auto-healing
- Auto-tuning

**This Druid Operator is approximately Level 3** - it handles installation, upgrades, scaling, and some self-healing.

---

## Operator Frameworks

Several frameworks help build operators:

| Framework | Language | Complexity | Use Case |
|-----------|----------|------------|----------|
| **Kubebuilder** | Go | Medium | Production operators |
| **Operator SDK** | Go/Ansible/Helm | Medium | Various approaches |
| **Metacontroller** | Any (webhooks) | Low | Simple operators |
| **KUDO** | YAML | Low | Stateful apps |

**This project uses Kubebuilder** - the most common choice for Go operators.

---

## Operator vs Helm Chart

| Aspect | Helm Chart | Operator |
|--------|------------|----------|
| **What it is** | Package manager | Controller |
| **When it runs** | At install time | Continuously |
| **Day 2 operations** | Manual | Automated |
| **Self-healing** | No | Yes |
| **Upgrades** | Manual | Can be automated |
| **Complexity** | Lower | Higher |

**Use Helm when:**
- Simple, stateless applications
- One-time deployment
- No special operational needs

**Use Operator when:**
- Complex, stateful applications
- Need continuous management
- Application-specific operational knowledge

---

## Real-World Example: Druid Operator

Without operator (manual):
```bash
# Create namespace
kubectl create namespace druid

# Create ConfigMaps (multiple files)
kubectl apply -f common-config.yaml
kubectl apply -f broker-config.yaml
kubectl apply -f historical-config.yaml
# ... more configs

# Create Services
kubectl apply -f broker-service.yaml
kubectl apply -f historical-service.yaml
# ... more services

# Create StatefulSets
kubectl apply -f broker-statefulset.yaml
kubectl apply -f historical-statefulset.yaml
# ... more statefulsets

# Monitor and fix issues manually
# Upgrade manually
# Scale manually
```

With operator:
```bash
# Install operator (once)
helm install druid-operator datainfra/druid-operator

# Deploy Druid cluster
kubectl apply -f my-druid-cluster.yaml

# That's it! Operator handles everything else.
```

---

## Next Steps

Continue to [Custom Resources](../06-custom-resources/README.md) to understand CRDs and CRs in detail.
