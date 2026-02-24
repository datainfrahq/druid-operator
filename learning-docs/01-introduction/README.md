# 1. Introduction - What is This Project?

## The Problem This Solves

Imagine you want to run Apache Druid (a real-time analytics database) on Kubernetes. Druid is a **distributed system** with multiple components:

- **Coordinator** - Manages data availability
- **Overlord** - Manages data ingestion tasks
- **Broker** - Handles queries from clients
- **Router** - Routes requests to the right service
- **Historical** - Stores and serves historical data
- **MiddleManager/Indexer** - Handles data ingestion

To run Druid manually on Kubernetes, you would need to create:
- Multiple StatefulSets (one for each component)
- Multiple ConfigMaps (configuration files)
- Multiple Services (for networking)
- PersistentVolumeClaims (for storage)
- And more...

This could be **50+ YAML files** that you need to manage, update, and keep in sync!

## The Solution: An Operator

This operator lets you define your entire Druid cluster in **ONE simple YAML file**:

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
    historicals:
      nodeType: historical
      replicas: 3
    # ... other nodes
```

The operator then:
1. **Reads** this YAML file
2. **Creates** all necessary Kubernetes resources automatically
3. **Monitors** the cluster continuously
4. **Heals** the cluster if something goes wrong
5. **Updates** the cluster when you change the YAML

## Key Concepts in This Project

### 1. Custom Resource Definition (CRD)
A CRD extends Kubernetes with new resource types. This project defines two CRDs:
- `Druid` - Represents a Druid cluster
- `DruidIngestion` - Represents a data ingestion job

### 2. Custom Resource (CR)
A CR is an instance of a CRD. When you create a YAML file with `kind: Druid`, you're creating a CR.

### 3. Controller
The controller is the "brain" that watches for CRs and takes action. It runs in a loop:
1. Watch for changes to Druid CRs
2. Compare desired state (what the CR says) vs actual state (what exists in K8s)
3. Take action to make actual state match desired state

### 4. Reconciliation
The process of making the actual state match the desired state is called "reconciliation."

## Project Components

```
druid-operator/
├── apis/                    # CRD definitions (what a Druid CR looks like)
├── controllers/             # Controller logic (what to do when CR changes)
├── config/                  # Kubernetes manifests for deploying the operator
├── chart/                   # Helm chart for easy installation
├── examples/                # Example Druid cluster configurations
├── main.go                  # Entry point - starts the operator
└── docs/                    # Documentation
```

## How It Works (High Level)

```
┌─────────────────────────────────────────────────────────────────┐
│                        Kubernetes Cluster                        │
│                                                                   │
│  ┌─────────────┐         ┌─────────────────────────────────┐    │
│  │   You       │         │      Druid Operator             │    │
│  │  (User)     │         │                                 │    │
│  └──────┬──────┘         │  ┌───────────────────────────┐  │    │
│         │                │  │    Controller             │  │    │
│         │ kubectl apply  │  │                           │  │    │
│         │                │  │  1. Watch Druid CRs       │  │    │
│         ▼                │  │  2. Compare states        │  │    │
│  ┌─────────────┐         │  │  3. Create/Update/Delete  │  │    │
│  │  Druid CR   │◄────────┤  │     K8s resources         │  │    │
│  │  (YAML)     │         │  └───────────────────────────┘  │    │
│  └─────────────┘         └─────────────────────────────────┘    │
│         │                                                        │
│         │ Operator creates these automatically:                  │
│         ▼                                                        │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  StatefulSets, Services, ConfigMaps, PVCs, etc.         │    │
│  │  (All the resources needed to run Druid)                │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

## Next Steps

Continue to [Prerequisites & Learning Path](../02-prerequisites/README.md) to understand what technologies you need to learn.
