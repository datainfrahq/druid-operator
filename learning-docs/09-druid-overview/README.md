# 9. Apache Druid Overview

## What is Apache Druid?

Apache Druid is a **real-time analytics database** designed for:
- **Fast queries** on large datasets (sub-second response times)
- **Real-time data ingestion** (streaming from Kafka, Kinesis)
- **High concurrency** (thousands of queries per second)
- **Time-series data** (events with timestamps)

**Use cases:**
- Clickstream analytics
- Network monitoring
- IoT sensor data
- Business intelligence dashboards

---

## Druid Architecture

Druid is a **distributed system** with multiple node types:

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Druid Cluster                                 │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │                    Master Nodes                              │    │
│  │  ┌─────────────┐    ┌─────────────┐                         │    │
│  │  │ Coordinator │    │  Overlord   │                         │    │
│  │  │             │    │             │                         │    │
│  │  │ - Manages   │    │ - Manages   │                         │    │
│  │  │   segments  │    │   ingestion │                         │    │
│  │  │ - Balances  │    │   tasks     │                         │    │
│  │  │   data      │    │             │                         │    │
│  │  └─────────────┘    └─────────────┘                         │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │                    Query Nodes                               │    │
│  │  ┌─────────────┐    ┌─────────────┐                         │    │
│  │  │   Broker    │    │   Router    │                         │    │
│  │  │             │    │             │                         │    │
│  │  │ - Receives  │    │ - Routes    │                         │    │
│  │  │   queries   │    │   requests  │                         │    │
│  │  │ - Merges    │    │ - API       │                         │    │
│  │  │   results   │    │   gateway   │                         │    │
│  │  └─────────────┘    └─────────────┘                         │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │                    Data Nodes                                │    │
│  │  ┌─────────────┐    ┌─────────────┐                         │    │
│  │  │ Historical  │    │MiddleManager│                         │    │
│  │  │             │    │  /Indexer   │                         │    │
│  │  │ - Stores    │    │             │                         │    │
│  │  │   segments  │    │ - Ingests   │                         │    │
│  │  │ - Serves    │    │   data      │                         │    │
│  │  │   queries   │    │ - Creates   │                         │    │
│  │  │             │    │   segments  │                         │    │
│  │  └─────────────┘    └─────────────┘                         │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │                  External Dependencies                       │    │
│  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐     │    │
│  │  │  ZooKeeper  │    │  Metadata   │    │    Deep     │     │    │
│  │  │             │    │   Store     │    │   Storage   │     │    │
│  │  │ - Service   │    │ (MySQL/     │    │ (S3/HDFS/   │     │    │
│  │  │   discovery │    │  PostgreSQL)│    │  local)     │     │    │
│  │  └─────────────┘    └─────────────┘    └─────────────┘     │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Node Types Explained

### 1. Coordinator
**Role:** Manages data availability and distribution

**Responsibilities:**
- Assigns segments to Historical nodes
- Balances data across the cluster
- Manages data retention rules
- Handles segment compaction

**Stateful:** Yes (needs to track segment assignments)

### 2. Overlord
**Role:** Manages data ingestion tasks

**Responsibilities:**
- Accepts ingestion task submissions
- Distributes tasks to MiddleManagers
- Monitors task status
- Handles task failures

**Stateful:** Yes (tracks task state)

**Note:** Often co-located with Coordinator (`coordinator.asOverlord.enabled=true`)

### 3. Broker
**Role:** Query router and result merger

**Responsibilities:**
- Receives queries from clients
- Determines which nodes have relevant data
- Fans out queries to Historical/MiddleManager nodes
- Merges partial results

**Stateful:** No (can be scaled horizontally)

### 4. Router
**Role:** API gateway

**Responsibilities:**
- Routes requests to appropriate services
- Provides unified API endpoint
- Hosts the web console

**Stateful:** No (can be scaled horizontally)

### 5. Historical
**Role:** Stores and serves historical data

**Responsibilities:**
- Downloads segments from deep storage
- Caches segments locally
- Serves queries for historical data

**Stateful:** Yes (needs persistent storage for segment cache)

### 6. MiddleManager / Indexer
**Role:** Handles data ingestion

**Responsibilities:**
- Runs ingestion tasks (Peons)
- Creates new segments
- Uploads segments to deep storage

**Stateful:** Yes (needs storage for intermediate data)

---

## Data Flow

### Ingestion Flow
```
Data Source (Kafka/Files)
        │
        ▼
   ┌─────────┐
   │Overlord │  Accepts ingestion spec
   └────┬────┘
        │
        ▼
┌──────────────┐
│MiddleManager │  Runs ingestion task
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Deep Storage │  Stores segments (S3/HDFS)
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Coordinator  │  Assigns segments to Historicals
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  Historical  │  Downloads and serves segments
└──────────────┘
```

### Query Flow
```
   Client
     │
     ▼
┌─────────┐
│ Router  │  Routes to Broker
└────┬────┘
     │
     ▼
┌─────────┐
│ Broker  │  Determines which nodes have data
└────┬────┘
     │
     ├──────────────┬──────────────┐
     ▼              ▼              ▼
┌─────────┐  ┌─────────┐  ┌──────────────┐
│Historical│  │Historical│  │MiddleManager │
│  Node 1  │  │  Node 2  │  │(real-time)   │
└────┬─────┘  └────┬─────┘  └──────┬───────┘
     │              │               │
     └──────────────┴───────────────┘
                    │
                    ▼
              ┌─────────┐
              │ Broker  │  Merges results
              └────┬────┘
                   │
                   ▼
                Client
```

---

## Why Druid Needs an Operator

### Complexity
- 6+ different node types
- Each with different configuration
- Complex interdependencies
- Specific startup/shutdown order

### Stateful Components
- Historical nodes need persistent storage
- Coordinators need stable identity
- Data must survive pod restarts

### Operational Knowledge
- Rolling updates must follow specific order
- Scaling requires understanding of data distribution
- Failure recovery needs domain knowledge

### Configuration Management
- Common configuration shared across nodes
- Node-specific configuration
- Runtime properties, JVM options, logging

---

## Druid Configuration in the Operator

### Common Configuration
Shared by all nodes:
```yaml
common.runtime.properties: |
  # ZooKeeper connection
  druid.zk.service.host=zookeeper:2181
  
  # Metadata store
  druid.metadata.storage.type=mysql
  druid.metadata.storage.connector.connectURI=jdbc:mysql://mysql:3306/druid
  
  # Deep storage
  druid.storage.type=s3
  druid.storage.bucket=druid-segments
  
  # Extensions
  druid.extensions.loadList=["druid-kafka-indexing-service", "druid-s3-extensions"]
```

### Node-Specific Configuration
Each node type has its own:
```yaml
nodes:
  brokers:
    runtime.properties: |
      druid.service=druid/broker
      druid.broker.http.numConnections=5
      druid.processing.buffer.sizeBytes=100000000
    extra.jvm.options: |-
      -Xmx4g
      -Xms4g
```

---

## Segments: Druid's Data Unit

Druid stores data in **segments**:
- Immutable chunks of data
- Typically cover a time range (hour, day)
- Stored in deep storage (S3, HDFS)
- Cached locally on Historical nodes

```
┌─────────────────────────────────────────────────────────────┐
│                    Segment Lifecycle                         │
│                                                              │
│  1. MiddleManager creates segment from ingested data         │
│                    │                                         │
│                    ▼                                         │
│  2. Segment uploaded to Deep Storage (S3)                    │
│                    │                                         │
│                    ▼                                         │
│  3. Coordinator assigns segment to Historical                │
│                    │                                         │
│                    ▼                                         │
│  4. Historical downloads and caches segment                  │
│                    │                                         │
│                    ▼                                         │
│  5. Historical serves queries from cached segment            │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Druid Extensions

Druid is extensible via extensions:

| Extension | Purpose |
|-----------|---------|
| `druid-kafka-indexing-service` | Kafka ingestion |
| `druid-kinesis-indexing-service` | Kinesis ingestion |
| `druid-s3-extensions` | S3 deep storage |
| `druid-hdfs-storage` | HDFS deep storage |
| `druid-mysql-metadata-storage` | MySQL metadata |
| `druid-postgresql-metadata-storage` | PostgreSQL metadata |

Configured in common.runtime.properties:
```properties
druid.extensions.loadList=["druid-kafka-indexing-service", "druid-s3-extensions"]
```

---

## Next Steps

Continue to [Complete Flow Walkthrough](../10-complete-flow/README.md) to see how everything works together end-to-end.
