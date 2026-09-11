# 4. Kubernetes Fundamentals

## What is Kubernetes?

Kubernetes (K8s) is a container orchestration platform. Think of it as:
- **A data center operating system** - manages compute, storage, networking
- **A declarative system** - you describe WHAT you want, K8s figures out HOW
- **Self-healing** - automatically restarts failed containers

---

## Core Concepts

### 1. Pod - The Smallest Unit

A Pod is one or more containers that:
- Share the same network (localhost)
- Share the same storage volumes
- Are scheduled together on the same node

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - name: my-container
    image: nginx:latest
    ports:
    - containerPort: 80
```

**Key Points:**
- Pods are ephemeral (temporary) - they can be killed and recreated
- Pods get a unique IP address
- Don't create Pods directly - use Deployments or StatefulSets

---

### 2. Deployment - Stateless Applications

A Deployment manages a set of identical Pods:
- Ensures desired number of replicas are running
- Handles rolling updates
- Supports rollback

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.14.2
        ports:
        - containerPort: 80
```

**When to use:**
- Stateless applications (web servers, APIs)
- Applications where any Pod can handle any request
- Applications that don't need stable network identity

---

### 3. StatefulSet - Stateful Applications

A StatefulSet is like a Deployment but for stateful applications:
- Pods get stable, unique network identities (pod-0, pod-1, pod-2)
- Pods get stable storage (PersistentVolumeClaims)
- Pods are created/deleted in order

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: mysql
spec:
  serviceName: mysql
  replicas: 3
  selector:
    matchLabels:
      app: mysql
  template:
    metadata:
      labels:
        app: mysql
    spec:
      containers:
      - name: mysql
        image: mysql:5.7
        volumeMounts:
        - name: data
          mountPath: /var/lib/mysql
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 10Gi
```

**When to use:**
- Databases (MySQL, PostgreSQL, Druid!)
- Applications that need stable network identity
- Applications that need persistent storage

**Why Druid uses StatefulSets:**
- Historical nodes need persistent storage for data
- Nodes need stable identities for cluster coordination

---

### 4. Service - Network Access to Pods

A Service provides a stable network endpoint for Pods:
- Pods come and go, but Service IP stays the same
- Load balances traffic across Pods
- Provides DNS name for discovery

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  selector:
    app: my-app
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP  # Internal only
```

**Service Types:**

| Type | Description | Use Case |
|------|-------------|----------|
| `ClusterIP` | Internal IP only | Internal communication |
| `NodePort` | Exposes on each node's IP | Development/testing |
| `LoadBalancer` | Cloud load balancer | Production external access |
| `Headless` | No cluster IP, DNS only | StatefulSet pod discovery |

**Headless Service (important for StatefulSets):**
```yaml
apiVersion: v1
kind: Service
metadata:
  name: mysql-headless
spec:
  clusterIP: None  # This makes it headless!
  selector:
    app: mysql
  ports:
  - port: 3306
```

With headless service, you can access individual pods:
- `mysql-0.mysql-headless.namespace.svc.cluster.local`
- `mysql-1.mysql-headless.namespace.svc.cluster.local`

---

### 5. ConfigMap - Configuration Data

A ConfigMap stores non-sensitive configuration:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  database.host: "mysql.default.svc"
  database.port: "3306"
  config.properties: |
    key1=value1
    key2=value2
```

**Using ConfigMap in a Pod:**
```yaml
spec:
  containers:
  - name: app
    env:
    - name: DB_HOST
      valueFrom:
        configMapKeyRef:
          name: app-config
          key: database.host
    volumeMounts:
    - name: config
      mountPath: /etc/config
  volumes:
  - name: config
    configMap:
      name: app-config
```

---

### 6. Secret - Sensitive Data

A Secret stores sensitive data (passwords, tokens):

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: db-secret
type: Opaque
data:
  username: YWRtaW4=      # base64 encoded "admin"
  password: cGFzc3dvcmQ=  # base64 encoded "password"
```

---

### 7. PersistentVolumeClaim (PVC) - Storage

A PVC requests storage from the cluster:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: standard
```

**Access Modes:**
- `ReadWriteOnce` (RWO) - Single node read/write
- `ReadOnlyMany` (ROX) - Multiple nodes read-only
- `ReadWriteMany` (RWX) - Multiple nodes read/write

---

### 8. Namespace - Resource Isolation

Namespaces provide logical isolation:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: production
```

```bash
# Create resources in a namespace
kubectl apply -f deployment.yaml -n production

# List resources in a namespace
kubectl get pods -n production
```

---

## How Kubernetes Works

### The Control Loop

```
┌─────────────────────────────────────────────────────────────┐
│                    Kubernetes Control Plane                  │
│                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │  API Server  │◄───│   etcd       │    │  Scheduler   │  │
│  │              │    │  (Database)  │    │              │  │
│  └──────┬───────┘    └──────────────┘    └──────────────┘  │
│         │                                                    │
│         │                                                    │
│  ┌──────▼───────┐                                           │
│  │  Controller  │  ◄── Watches resources, takes action      │
│  │  Manager     │                                           │
│  └──────────────┘                                           │
└─────────────────────────────────────────────────────────────┘
         │
         │ Creates/Updates/Deletes
         ▼
┌─────────────────────────────────────────────────────────────┐
│                    Worker Nodes                              │
│                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │   kubelet    │    │   kubelet    │    │   kubelet    │  │
│  │   + Pods     │    │   + Pods     │    │   + Pods     │  │
│  └──────────────┘    └──────────────┘    └──────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Declarative vs Imperative

**Imperative (telling K8s what to DO):**
```bash
kubectl create deployment nginx --image=nginx
kubectl scale deployment nginx --replicas=3
kubectl set image deployment/nginx nginx=nginx:1.16
```

**Declarative (telling K8s what you WANT):**
```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.16
```
```bash
kubectl apply -f deployment.yaml
```

**Declarative is preferred because:**
- Configuration is version controlled
- Easy to reproduce
- Self-documenting
- Operators work declaratively

---

## Labels and Selectors

Labels are key-value pairs attached to resources:

```yaml
metadata:
  labels:
    app: druid
    component: broker
    environment: production
```

Selectors find resources by labels:

```yaml
spec:
  selector:
    matchLabels:
      app: druid
      component: broker
```

---

## Resource Lifecycle

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Pending   │────►│   Running   │────►│  Succeeded  │
└─────────────┘     └──────┬──────┘     └─────────────┘
                           │
                           ▼
                    ┌─────────────┐
                    │   Failed    │
                    └─────────────┘
```

---

## Common kubectl Commands

```bash
# Get resources
kubectl get pods
kubectl get pods -n my-namespace
kubectl get pods -o wide  # More details
kubectl get pods -o yaml  # Full YAML

# Describe (detailed info)
kubectl describe pod my-pod

# Logs
kubectl logs my-pod
kubectl logs my-pod -f  # Follow
kubectl logs my-pod -c my-container  # Specific container

# Execute command in pod
kubectl exec -it my-pod -- /bin/bash

# Apply configuration
kubectl apply -f my-resource.yaml

# Delete
kubectl delete pod my-pod
kubectl delete -f my-resource.yaml

# Watch for changes
kubectl get pods -w
```

---

## Next Steps

Continue to [What is an Operator?](../05-operators/README.md) to understand the operator pattern.
