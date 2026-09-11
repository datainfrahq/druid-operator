# 2. Prerequisites & Learning Path

## What You Need to Learn

As a Java developer with basic Kubernetes knowledge, here's your learning roadmap:

## 1. Go Programming Language (Essential)

### Why Go for Kubernetes?
- Kubernetes itself is written in Go
- All official Kubernetes libraries are in Go
- Go compiles to a single binary (easy to deploy)
- Go has excellent concurrency support (goroutines)
- Go is simpler than Java (no classes, no inheritance)

### What to Learn in Go

| Topic | Priority | Time Estimate |
|-------|----------|---------------|
| Basic syntax (variables, functions, loops) | High | 2-3 hours |
| Structs and methods | High | 2 hours |
| Interfaces | High | 2 hours |
| Pointers | High | 1-2 hours |
| Error handling | High | 1 hour |
| Packages and modules | High | 1 hour |
| Goroutines and channels | Medium | 2-3 hours |
| Context package | Medium | 1 hour |

### Recommended Resources
1. **A Tour of Go** (official): https://go.dev/tour/
2. **Go by Example**: https://gobyexample.com/
3. **Effective Go**: https://go.dev/doc/effective_go

---

## 2. Kubernetes Concepts (Essential)

### Core Concepts You Must Know

| Concept | What It Is | Why It Matters |
|---------|------------|----------------|
| **Pod** | Smallest deployable unit | Druid nodes run in Pods |
| **Deployment** | Manages stateless Pods | Some Druid nodes use this |
| **StatefulSet** | Manages stateful Pods | Most Druid nodes use this |
| **Service** | Network endpoint for Pods | How Druid nodes communicate |
| **ConfigMap** | Configuration storage | Druid configuration files |
| **Secret** | Sensitive data storage | Passwords, credentials |
| **PersistentVolumeClaim** | Storage request | Druid data storage |
| **Namespace** | Resource isolation | Organize resources |

### Advanced Concepts for Operators

| Concept | What It Is | Why It Matters |
|---------|------------|----------------|
| **Custom Resource Definition (CRD)** | Extends K8s API | Defines what "Druid" means |
| **Custom Resource (CR)** | Instance of CRD | Your Druid cluster definition |
| **Controller** | Watches and acts on resources | The operator's brain |
| **Finalizer** | Cleanup hook | Clean up when Druid is deleted |
| **Owner Reference** | Parent-child relationship | Automatic garbage collection |

### Recommended Resources
1. **Kubernetes Basics**: https://kubernetes.io/docs/tutorials/kubernetes-basics/
2. **Kubernetes Concepts**: https://kubernetes.io/docs/concepts/

---

## 3. Controller-Runtime Library (Important)

This is the Go library used to build Kubernetes operators.

### Key Concepts

| Concept | Description |
|---------|-------------|
| **Manager** | Runs controllers, handles leader election |
| **Controller** | Watches resources, triggers reconciliation |
| **Reconciler** | Your logic - what to do when resource changes |
| **Client** | Talks to Kubernetes API |
| **Scheme** | Knows about resource types |

### Recommended Resources
1. **Kubebuilder Book**: https://book.kubebuilder.io/
2. **Controller-Runtime Docs**: https://pkg.go.dev/sigs.k8s.io/controller-runtime

---

## 4. Apache Druid (Good to Know)

Understanding Druid helps you understand why the operator is designed this way.

### Druid Components

| Component | Role | Stateful? |
|-----------|------|-----------|
| **Coordinator** | Manages data distribution | Yes |
| **Overlord** | Manages ingestion tasks | Yes |
| **Broker** | Query routing | No |
| **Router** | API gateway | No |
| **Historical** | Serves historical data | Yes |
| **MiddleManager** | Runs ingestion tasks | Yes |

### Recommended Resources
1. **Druid Introduction**: https://druid.apache.org/docs/latest/design/
2. **Druid Architecture**: https://druid.apache.org/docs/latest/design/architecture.html

---

## Learning Order Recommendation

```
Week 1: Go Basics
├── Day 1-2: Go syntax, types, functions
├── Day 3-4: Structs, methods, interfaces
└── Day 5-7: Packages, error handling, pointers

Week 2: Kubernetes Deep Dive
├── Day 1-2: Pods, Deployments, StatefulSets
├── Day 3-4: Services, ConfigMaps, PVCs
└── Day 5-7: CRDs, Controllers concept

Week 3: Operator Development
├── Day 1-3: Kubebuilder tutorial
├── Day 4-5: Controller-runtime basics
└── Day 6-7: Study this codebase

Week 4: Apache Druid
├── Day 1-3: Druid architecture
├── Day 4-5: Druid configuration
└── Day 6-7: Run Druid locally
```

---

## Quick Self-Assessment

Before diving into the code, make sure you can answer:

### Go
- [ ] What's the difference between `var x int` and `x := 0`?
- [ ] How do you define a method on a struct?
- [ ] What's an interface in Go?
- [ ] How does error handling work in Go?

### Kubernetes
- [ ] What's the difference between Deployment and StatefulSet?
- [ ] What is a Service and why do we need it?
- [ ] What is a ConfigMap used for?
- [ ] What happens when you `kubectl apply -f file.yaml`?

### Operators
- [ ] What is a CRD?
- [ ] What is reconciliation?
- [ ] What is a controller?

If you can't answer these, spend more time on the prerequisites before diving into the code.

---

## Next Steps

Continue to [Go vs Java](../03-go-vs-java/README.md) to understand Go from a Java developer's perspective.
